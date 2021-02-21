package proxy

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	rhttp "net/http"
	"net/url"
	"strings"
	"time"

	"github.com/voiladev/VoilaCrawl/pkg/net/http"
	pbHttp "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/http"
	pbProxy "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
	"google.golang.org/protobuf/encoding/protojson"
)

type Options struct {
	ProxyAddr string
}

// proxyClient
type proxyClient struct {
	jar             http.CookieJar
	httpClient      *rhttp.Client
	httpProxyClient *rhttp.Client
	options         Options
	logger          glog.Log
}

// NewProxyClient returns a http client which support native http.Client and
// supports Proxy Backconnect proxy, Crawling API.
func NewProxyClient(proxyAddr string, cookieJar http.CookieJar, logger glog.Log) (http.Client, error) {
	if _, err := url.Parse(proxyAddr); err != nil {
		return nil, errors.New("invalid proxy http address")
	}
	if cookieJar == nil {
		return nil, errors.New("invalid cookiejar")
	}
	if logger == nil {
		return nil, errors.New("invaild logger")
	}
	client := proxyClient{
		jar:        cookieJar,
		httpClient: &rhttp.Client{},
		options: Options{
			ProxyAddr: proxyAddr,
		},
		logger: logger,
	}

	return &client, nil
}

func (c *proxyClient) Jar() http.CookieJar {
	if c == nil {
		return nil
	}
	return c.jar
}

func (c *proxyClient) Do(ctx context.Context, r *http.Request) (*http.Response, error) {
	if c == nil {
		return nil, nil
	}
	return c.DoWithOptions(ctx, r, http.Options{})
}

// DoWithOptions
// If proxy is enabled, the client will try backconnect proxy for at most twice.
// If all the two try failed, then this proxy will try crawling api.
// To remember that timeout is controlled by ctx
func (c *proxyClient) DoWithOptions(ctx context.Context, r *http.Request, opts http.Options) (resp *http.Response, err error) {
	if c == nil || r == nil {
		return nil, nil
	}

	var body []byte
	if r.Body != nil {
		body, err = io.ReadAll(r.Body)
		if err != nil {
			c.logger.Debug(err)
			return nil, err
		}
	}

	req := pbProxy.Request{
		Method:  r.Method,
		Url:     r.URL.String(),
		Body:    body,
		Headers: make(map[string]*pbHttp.ListValue),
		Options: &pbProxy.Request_Options{
			EnableProxy:       opts.EnableProxy,
			Reliability:       opts.Reliability,
			EnableHeadless:    opts.EnableHeadless,
			EnableSessionInit: opts.EnableSessionInit,
			KeepSession:       opts.KeepSession,
			MaxTtlPerRequest:  10 * 60, // 10mins
			DisableRedirect:   opts.DisableRedirect,
		},
	}
	// set ttl per request according to deadline
	if deadline, ok := ctx.Deadline(); ok {
		timeRemain := deadline.Unix() - time.Now().Unix()
		if timeRemain > 0 {
			req.Options.MaxTtlPerRequest = timeRemain
		}
	}

	if ctx.Value("tracing_id") != nil {
		req.TracingId = ctx.Value("tracing_id").(string)
	}
	if ctx.Value("job_id") != nil {
		req.JobId = ctx.Value("job_id").(string)
	}
	if ctx.Value("req_id") != nil {
		req.ReqId = ctx.Value("req_id").(string)
	} else {
		req.ReqId = fmt.Sprintf("req_%d", time.Now().UnixNano())
	}
	for key, vals := range r.Header {
		req.Headers[key] = &pbHttp.ListValue{Values: vals}
	}

	data, _ := protojson.Marshal(&req)
	proxyReq, err := rhttp.NewRequest(rhttp.MethodPost, c.options.ProxyAddr, bytes.NewReader(data))
	if err != nil {
		c.logger.Debug(err)
		return nil, err
	}
	proxyReq = proxyReq.WithContext(ctx)

	proxyResp, err := c.httpClient.Do(proxyReq)
	if err != nil {
		return nil, err
	}
	defer proxyResp.Body.Close()

	if proxyResp.StatusCode != 200 {
		return nil, fmt.Errorf("do http request failed with status %d %s", proxyResp.StatusCode, proxyResp.Status)
	}

	var proxyRespBody pbProxy.Response
	if respBody, err := io.ReadAll(proxyResp.Body); err != nil {
		return nil, err
	} else if err := protojson.Unmarshal(respBody, &proxyRespBody); err != nil {
		return nil, err
	}
	if proxyRespBody.GetStatusCode() == -1 {
		return nil, errors.New(proxyRespBody.Status)
	}

	var buildResponse func(res *pbProxy.Response, isSub bool) (*http.Response, error)
	buildResponse = func(res *pbProxy.Response, isSub bool) (*http.Response, error) {
		if res == nil {
			return nil, nil
		}
		resp := http.Response{
			StatusCode: int(res.GetStatusCode()),
			Status:     res.GetStatus(),
			Proto:      res.GetProto(),
			ProtoMajor: int(res.GetProtoMajor()),
			ProtoMinor: int(res.GetProtoMinor()),
			Header:     http.Header{},
		}
		for k, lv := range res.Headers {
			resp.Header[k] = lv.Values
		}
		if !isSub && len(res.Body) > 0 {
			// try to uncompress ziped data
			if strings.EqualFold(resp.Header.Get("Content-Encoding"), "gzip") {
				if reader, err := gzip.NewReader(bytes.NewReader(res.Body)); err == nil {
					if data, err := io.ReadAll(reader); err == nil {
						resp.Body = http.NewReader(data)
						resp.Header.Del("Content-Encoding")
						resp.Header.Del("Content-Length")
						resp.ContentLength = -1
						resp.Uncompressed = true
					}
				} else {
					resp.ContentLength = int64(len(res.Body))
					resp.Body = http.NewReader(res.GetBody())
				}
			} else {
				resp.ContentLength = int64(len(res.Body))
				resp.Uncompressed = true
				resp.Body = http.NewReader(res.GetBody())
			}
		}

		if res.Request != nil {
			r := res.Request
			u, _ := url.Parse(r.GetUrl())
			resp.Request = &http.Request{
				Method: r.Method,
				URL:    u,
				Header: http.Header{},
			}
			for k, lv := range r.Headers {
				resp.Request.Header[k] = lv.Values
			}

			if res.Request.Response != nil {
				resp.Request.Response, _ = buildResponse(res.Request.Response, true)
			}
		}
		return &resp, nil
	}
	return buildResponse(&proxyRespBody, false)
}
