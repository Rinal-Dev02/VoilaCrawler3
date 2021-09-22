package proxy

import (
	"context"
	"errors"
	"fmt"
	"io"
	rhttp "net/http"
	"net/url"
	"time"

	ctxutil "github.com/voiladev/VoilaCrawler/pkg/context"
	"github.com/voiladev/VoilaCrawler/pkg/crawler"
	"github.com/voiladev/VoilaCrawler/pkg/net/http"
	pbHttp "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/http"
	pbProxy "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/randutil"
	"google.golang.org/grpc"
)

func NewPbProxyManagerClient(ctx context.Context, proxyAddr string) (pbProxy.ProxyManagerClient, error) {
	if proxyAddr == "" {
		return nil, errors.New("proxy address not specified")
	}
	conn, err := grpc.DialContext(
		ctx,
		proxyAddr,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithBackoffMaxDelay(time.Second),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(100*1024*1024)),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(100*1024*1024)),
	)
	if err != nil {
		return nil, err
	}
	return pbProxy.NewProxyManagerClient(conn), nil
}

// proxyClient
type proxyClient struct {
	pighubClient pbProxy.ProxyManagerClient
	jar          http.CookieJar
	logger       glog.Log
}

// NewProxyClient returns a http client which support native http.Client and
// supports Proxy Backconnect proxy, Crawling API.
func NewProxyClient(pighubClient pbProxy.ProxyManagerClient, cookieJar http.CookieJar, logger glog.Log) (http.Client, error) {
	if pighubClient == nil {
		return nil, errors.New("invalid pighubClient")
	}
	if cookieJar == nil {
		return nil, errors.New("invalid cookiejar")
	}
	if logger == nil {
		return nil, errors.New("invaild logger")
	}
	return &proxyClient{
		pighubClient: pighubClient,
		jar:          cookieJar,
		logger:       logger,
	}, nil
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
	c.logger.Infof("%s %s", r.Method, r.URL)

	var body []byte
	if r.Body != nil {
		body, err = io.ReadAll(r.Body)
		if err != nil {
			c.logger.Debug(err)
			return nil, err
		}
	}

	req := &pbProxy.Request{
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
			DisableCookieJar:  opts.DisableCookieJar,
			MaxTtlPerRequest:  5 * 60, // 5mins
			DisableRedirect:   opts.DisableRedirect,
			RequestFilterKeys: opts.RequestFilterKeys,
			JsWaitDuration:    opts.JsWaitDuration,
		},
	}
	// set ttl per request according to deadline
	if deadline, ok := ctx.Deadline(); ok {
		timeRemain := deadline.Unix() - time.Now().Unix()
		if timeRemain > 30 {
			req.Options.MaxTtlPerRequest = timeRemain
		}
	}

	req.TracingId = ctxutil.GetString(ctx, crawler.TracingIdKey)
	req.JobId = ctxutil.GetString(ctx, crawler.JobIdKey)
	req.ReqId = ctxutil.GetString(ctx, crawler.ReqIdKey)
	if req.ReqId == "" {
		req.ReqId = fmt.Sprintf("req_%s", randutil.MustNewRandomID())
	}
	for key, vals := range r.Header {
		req.Headers[key] = &pbHttp.ListValue{Values: vals}
	}

	proxyResp, err := c.pighubClient.DoRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	c.logger.Infof("%s %s %d", r.Method, r.URL, proxyResp.StatusCode)
	if proxyResp.StatusCode != 200 {
		return nil, fmt.Errorf("do http request failed with status %d %s", proxyResp.StatusCode, proxyResp.Status)
	}
	if proxyResp.GetRequest() == nil {
		proxyResp.Request = req
	}

	var buildResponse func(res *pbProxy.Response, isSub bool) (*rhttp.Response, error)
	buildResponse = func(res *pbProxy.Response, isSub bool) (*rhttp.Response, error) {
		if res == nil {
			return nil, nil
		}
		resp := rhttp.Response{
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
			// body data uncompress by httpproxy
			resp.Header.Del("Content-Encoding")
			resp.Header.Del("Content-Length")
			resp.ContentLength = int64(len(res.Body))
			resp.Uncompressed = true
			resp.Body = http.NewReader(res.GetBody())
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
	if resp, err := buildResponse(proxyResp, false); err != nil {
		return nil, err
	} else {
		return http.NewResponse(resp)
	}
}
