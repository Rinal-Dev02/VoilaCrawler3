package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	ctxutil "github.com/voiladev/go-crawler/pkg/context"
	"github.com/voiladev/go-crawler/pkg/crawler"
	"github.com/voiladev/go-crawler/pkg/net/http"
	pbCrawl "github.com/voiladev/go-crawler/protoc-gen-go/chameleon/smelter/v1/crawl"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
	pbError "github.com/voiladev/protobuf/protoc-gen-go/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"
)

var _ pbCrawl.CrawlerNodeServer = (*CrawlerServer)(nil)

type CrawlerServer struct {
	pbCrawl.UnimplementedCrawlerNodeServer

	crawler    crawler.Crawler
	httpClient http.Client
	logger     glog.Log
}

func NewCrawlerServer(cw crawler.Crawler, httpClient http.Client, logger glog.Log) (pbCrawl.CrawlerNodeServer, error) {
	if cw == nil {
		return nil, errors.New("invalid crawler")
	}
	if httpClient == nil {
		return nil, errors.New("invalid http client")
	}
	if logger == nil {
		return nil, errors.New("invalid logger")
	}
	return &CrawlerServer{
		crawler:    cw,
		httpClient: httpClient,
		logger:     logger.New("CrawlerServer"),
	}, nil
}

// Version
func (s *CrawlerServer) Version(ctx context.Context, req *emptypb.Empty) (*pbCrawl.VersionResponse, error) {
	if s == nil || s.crawler == nil {
		return nil, nil
	}
	return &pbCrawl.VersionResponse{
		Version: s.crawler.Version(),
	}, nil
}

// CrawlerOptions
func (s *CrawlerServer) CrawlerOptions(ctx context.Context, req *pbCrawl.CrawlerOptionsRequest) (*pbCrawl.CrawlerOptionsResponse, error) {
	if s == nil || s.crawler == nil {
		return nil, nil
	}
	u, err := url.Parse(req.GetUrl())
	if err != nil {
		return nil, pbError.ErrInvalidArgument.New(err)
	}

	rawOpts := s.crawler.CrawlOptions(u)
	var opts pbCrawl.CrawlerOptions
	if err := rawOpts.Unmarshal(&opts); err != nil {
		return nil, pbError.ErrInternal.New(err)
	}

	return &pbCrawl.CrawlerOptionsResponse{
		Data: &opts,
	}, nil
}

// AllowedDomains
func (s *CrawlerServer) AllowedDomains(ctx context.Context, req *emptypb.Empty) (*pbCrawl.AllowedDomainsResponse, error) {
	if s == nil || s.crawler == nil {
		return nil, nil
	}

	domains := s.crawler.AllowedDomains()
	return &pbCrawl.AllowedDomainsResponse{
		Data: domains,
	}, nil
}

// CanonicalUrl
func (s *CrawlerServer) CanonicalUrl(ctx context.Context, req *pbCrawl.CanonicalUrlRequest) (*pbCrawl.CanonicalUrlResponse, error) {
	if s == nil || s.crawler == nil {
		return nil, nil
	}

	curl, err := s.crawler.CanonicalUrl(req.GetUrl())
	if err != nil {
		return nil, pbError.ErrInvalidArgument.New(err)
	}
	return &pbCrawl.CanonicalUrlResponse{
		Data: &pbCrawl.CanonicalUrlResponse_Data{Url: curl},
	}, nil
}

// Parse do http request first and then do parse
func (s *CrawlerServer) Parse(rawreq *pbCrawl.Request, ps pbCrawl.CrawlerNode_ParseServer) error {
	if s == nil || s.crawler == nil {
		return nil
	}
	logger := s.logger.New("Parse")

	logger.Infof("Access %s", rawreq.GetUrl())

	shareCtx := ps.Context()
	for k, v := range rawreq.SharingData {
		shareCtx = context.WithValue(shareCtx, k, v)
	}
	shareCtx = context.WithValue(shareCtx, crawler.TracingIdKey, rawreq.GetTracingId())
	shareCtx = context.WithValue(shareCtx, crawler.JobIdKey, rawreq.GetJobId())
	shareCtx = context.WithValue(shareCtx, crawler.ReqIdKey, rawreq.GetReqId())
	shareCtx = context.WithValue(shareCtx, crawler.StoreIdKey, rawreq.GetStoreId())
	shareCtx = context.WithValue(shareCtx, crawler.TargetTypeKey, strings.Join(rawreq.GetOptions().GetTargetTypes(), ","))

	req, err := buildRequest(rawreq)
	if err != nil {
		logger.Error(err)
		return err
	}

	opts := s.crawler.CrawlOptions(req.URL)
	resp, err := s.httpClient.DoWithOptions(shareCtx, req, http.Options{
		EnableProxy:       !rawreq.Options.DisableProxy,
		EnableHeadless:    opts.EnableHeadless,
		EnableSessionInit: opts.EnableSessionInit,
		KeepSession:       opts.KeepSession,
		DisableCookieJar:  opts.DisableCookieJar,
		DisableRedirect:   opts.DisableRedirect,
		Reliability:       opts.Reliability,
	})
	if err != nil {
		logger.Error(err)
		return err
	}
	if resp.Body == nil {
		logger.Error("no response got")
		return errors.New("no response got")
	}

	// check and patch the rawurl
	if resp.Request == nil {
		resp.Request = req
	}
	rootResp := resp
	for rootResp.Request.Response != nil {
		rootResp = rootResp.Request.Response
	}
	if rootResp.Request.URL.Path != req.URL.Path {
		patchResp := http.Response{
			StatusCode: http.StatusTemporaryRedirect,
			Request:    req,
		}
		rootResp.Request.Response = &patchResp
	}

	err = s.crawler.Parse(shareCtx, resp, func(c context.Context, i interface{}) error {
		sharingData := ctxutil.RetrieveAllValues(c)
		tracingId := rawreq.GetTracingId()
		if tid := ctxutil.GetString(c, crawler.TracingIdKey); tid != "" {
			if !strings.HasPrefix(tid, "sub_") {
				tid = "sub_" + tid
			}
			tracingId = tid
		}

		switch val := i.(type) {
		case *http.Request:
			if val.URL.Host == "" {
				val.URL.Scheme = resp.Request.URL.Scheme
				val.URL.Host = resp.Request.URL.Host
			} else if val.URL.Scheme != "http" && val.URL.Scheme != "https" {
				val.URL.Scheme = resp.Request.URL.Scheme
			}
			if val.Header.Get("Referer") == "" && resp.Request != nil {
				val.Header.Set("Referer", resp.Request.URL.String())
			}

			// convert http.Request to pbCrawl.Command_Request and forward
			subreq := pbCrawl.Request{
				TracingId:     tracingId,
				JobId:         rawreq.GetJobId(),
				ReqId:         rawreq.GetReqId(),
				StoreId:       rawreq.GetStoreId(),
				Url:           val.URL.String(),
				Method:        val.Method,
				Parent:        rawreq,
				CustomHeaders: rawreq.CustomHeaders,
				CustomCookies: rawreq.CustomCookies,
				Options:       rawreq.Options,
				SharingData:   rawreq.SharingData,
			}

			if subreq.CustomHeaders == nil {
				subreq.CustomHeaders = make(map[string]string)
			}
			if subreq.SharingData == nil {
				subreq.SharingData = map[string]string{}
			}
			if val.Body != nil {
				defer val.Body.Close()
				if data, err := io.ReadAll(val.Body); err != nil {
					return err
				} else {
					subreq.Body = fmt.Sprintf("%s", data)
				}
			}
			for k := range val.Header {
				subreq.CustomHeaders[k] = val.Header.Get(k)
			}

			for k, v := range sharingData {
				key, ok := k.(string)
				if !ok {
					continue
				}
				val := strconv.Format(v)

				subreq.SharingData[key] = val
			}
			data, _ := anypb.New(&subreq)
			return ps.Send(data)
		case *pbCrawl.Error:
			if val.ErrMsg == "" {
				return nil
			}

			val.ReqId = rawreq.GetReqId()
			val.TracingId = tracingId
			val.JobId = rawreq.GetJobId()
			val.StoreId = rawreq.GetStoreId()
			val.Timestamp = time.Now().UnixNano() / 1000000

			data, _ := anypb.New(val)
			return ps.Send(data)
		default:
			msg, ok := i.(proto.Message)
			if !ok {
				return errors.New("unsupported response data type")
			}
			index := ctxutil.GetInt(c, "item.index")
			if index == 0 {
				index = ctxutil.GetInt(c, "index")
			}
			item := pbCrawl.Item{
				Timestamp: time.Now().UnixNano() / 1000000,
				TracingId: rawreq.GetTracingId(),
				JobId:     rawreq.GetJobId(),
				ReqId:     rawreq.GetReqId(),
				Index:     int32(index),
			}
			item.Data, _ = anypb.New(msg)
			data, _ := anypb.New(&item)
			return ps.Send(data)
		}
	})
	if err != nil {
		logger.Error(err)
	}

	if errors.Is(crawler.ErrAbort, err) {
		return pbError.ErrAborted.New(err.Error())
	} else if errors.Is(crawler.ErrUnsupportedPath, err) {
		return pbError.ErrUnimplemented.New(err.Error())
	} else if errors.Is(crawler.ErrUnsupportedTarget, err) {
		return pbError.ErrUnimplemented.New(err.Error())
	}
	return err
}

func buildRequest(r *pbCrawl.Request) (*http.Request, error) {
	if r == nil {
		return nil, errors.New("invalid request")
	}
	var reader io.Reader
	if r.GetBody() != "" {
		reader = bytes.NewReader([]byte(r.GetBody()))
	}
	req, err := http.NewRequest(r.Method, r.Url, reader)
	if err != nil {
		return nil, pbError.ErrInvalidArgument.New(err)
	}
	for k, v := range r.CustomHeaders {
		// ignore cookies in header
		if strings.ToLower(k) == "cookie" {
			continue
		}
		req.Header.Set(k, v)
	}
	for _, cookie := range r.CustomCookies {
		if !strings.HasPrefix(req.URL.Path, cookie.Path) && cookie.Path != "" {
			req.AddCookie(&http.Cookie{
				Name:  cookie.Name,
				Value: cookie.Value,
			})
		}
	}
	return req, nil
}
