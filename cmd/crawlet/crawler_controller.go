package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	ctxUtil "github.com/voiladev/VoilaCrawl/pkg/context"
	http "github.com/voiladev/VoilaCrawl/pkg/net/http"
	pbCrawl "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl"
	pbProxy "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
	pbError "github.com/voiladev/protobuf/protoc-gen-go/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type CrawlerController struct {
	ctx            context.Context
	crawlerManager *CrawlerManager
	httpClient     http.Client

	logger glog.Log
}

func NewCrawlerController(
	ctx context.Context,
	crawlerManager *CrawlerManager,
	httpClient http.Client,
	logger glog.Log,
) (*CrawlerController, error) {
	ctrl := CrawlerController{
		ctx:            ctx,
		crawlerManager: crawlerManager,
		httpClient:     httpClient,
		logger:         logger.New("CrawlerController"),
	}
	return &ctrl, nil
}

func (ctrl *CrawlerController) Parse(ctx context.Context, rawResp *pbCrawl.ParseRequest, yield func(ctx context.Context, data proto.Message) error) error {
	if ctrl == nil || yield == nil {
		return nil
	}
	logger := ctrl.logger.New("Parse")

	resp, err := buildResponse(rawResp.GetResponse(), false)
	if err != nil {
		logger.Errorf("build standard response failed, error=%s", err)
		return err
	}

	crawlers, err := ctrl.crawlerManager.GetByHost(ctx, resp.Request.URL.Host)
	if err != nil {
		logger.Errorf("get crawlers for host %s failed, error=%s", resp.Request.URL.Host, err)
		return err
	}
	if len(crawlers) == 0 {
		return pbError.ErrNotFound.New("no usable crawler found")
	}

	return func() (err error) {
		defer func() {
			if e := recover(); e != nil {
				err = fmt.Errorf("%v", e)
			}
		}()

		return crawlers[0].Parse(ctx, resp, func(c context.Context, i interface{}) error {
			sharingData := ctxUtil.RetrieveAllValues(c)
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
				subreq := pbCrawl.Command_Request{
					TracingId:     rawResp.GetRequest().GetTracingId(),
					JobId:         rawResp.GetRequest().GetJobId(),
					ReqId:         rawResp.GetRequest().GetReqId(),
					StoreId:       rawResp.GetRequest().GetStoreId(),
					Url:           val.URL.String(),
					Method:        val.Method,
					Parent:        rawResp.GetRequest(),
					CustomHeaders: rawResp.GetRequest().CustomHeaders,
					CustomCookies: rawResp.GetRequest().CustomCookies,
					Options:       rawResp.GetRequest().Options,
					SharingData:   rawResp.GetRequest().SharingData,
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

					if strings.HasSuffix(key, "tracing_id") ||
						strings.HasSuffix(key, "job_id") ||
						strings.HasSuffix(key, "req_id") ||
						strings.HasSuffix(key, "store_id") {
						continue
					}
					subreq.SharingData[key] = val
				}

				return yield(ctx, &subreq)
			default:
				msg, ok := i.(proto.Message)
				if !ok {
					return errors.New("unsupported response data type")
				}
				var index int64
				if indexVal, ok := sharingData["item.index"]; ok && indexVal != nil {
					index = strconv.MustParseInt(indexVal)
				}
				item := pbCrawl.Item{
					Timestamp: time.Now().UnixNano(),
					TracingId: rawResp.GetRequest().GetTracingId(),
					JobId:     rawResp.GetRequest().GetJobId(),
					ReqId:     rawResp.GetRequest().GetReqId(),
					Index:     int32(index),
				}
				item.Data, _ = anypb.New(msg)
				return yield(ctx, &item)
			}
		})
	}()
}

func buildResponse(res *pbProxy.Response, isSub bool) (*http.Response, error) {
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
