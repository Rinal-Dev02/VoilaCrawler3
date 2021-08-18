package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	ctxutil "github.com/voiladev/VoilaCrawler/pkg/context"
	"github.com/voiladev/VoilaCrawler/pkg/crawler"
	"github.com/voiladev/VoilaCrawler/pkg/net/http"
	pbCrawl "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl"
	pbItem "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/randutil"
	"github.com/voiladev/go-framework/strconv"
	"github.com/voiladev/go-framework/text"
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

	var methods []*pbCrawl.CrawlerMethod
	if _, ok := s.crawler.(crawler.ProductCrawler); ok {
		for i := 0; i < productCrawlerType.NumMethod(); i++ {
			method := productCrawlerType.Method(i)
			cm := pbCrawl.CrawlerMethod{
				Name: method.Name,
			}
			if inArgs := method.Type.NumIn(); inArgs > 0 {
				firstInType := method.Type.In(0)
				if !firstInType.Implements(reflect.TypeOf((*context.Context)(nil))) || inArgs > 1 {
					cm.RequireInput = true
				}
			}
		}
	}
	return &pbCrawl.CrawlerOptionsResponse{
		Data:        &opts,
		RemoteCalls: methods,
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

var (
	protoMessageType   = reflect.TypeOf((*proto.Message)(nil)).Elem()
	productCrawlerType = reflect.TypeOf((*crawler.ProductCrawler)(nil)).Elem()
)

func marshalAny(val reflect.Value) (*anypb.Any, error) {
	if val.IsNil() || val.IsZero() {
		return nil, errors.New("nil/invalid value")
	}
	if !val.Type().Implements(protoMessageType) {
		return nil, errors.New("object not implement interface proto.Message")
	}
	interVal := val.Interface()
	v, _ := interVal.(proto.Message)

	if data, err := anypb.New(v); err != nil {
		return nil, err
	} else {
		return data, nil
	}
}

// Call
func (s *CrawlerServer) Call(ctx context.Context, req *pbCrawl.CallRequest) (*pbCrawl.CallResponse, error) {
	if s == nil || s.crawler == nil {
		return nil, nil
	}
	if req.GetMethod() == "" {
		return nil, pbError.ErrInvalidArgument.New(`method required`)
	}
	if !(req.GetMethod()[0] > 'A' && req.GetMethod()[0] < 'Z') {
		return nil, pbError.ErrPermissionDenied.New(fmt.Sprintf(`private method "%s" is not callable`, req.GetMethod()))
	}

	shareCtx := context.WithValue(ctx, crawler.TracingIdKey, req.GetTracingId())
	shareCtx = context.WithValue(shareCtx, crawler.JobIdKey, req.GetJobId())
	shareCtx = context.WithValue(shareCtx, crawler.ReqIdKey, randutil.MustNewRandomID())
	shareCtx = context.WithValue(shareCtx, crawler.SiteIdKey, s.crawler.ID())

	cw := reflect.ValueOf(s.crawler)
	if !cw.Type().Implements(productCrawlerType) {
		return nil, pbError.ErrUnimplemented.New(fmt.Sprintf(`method "%s" unimplemented or is not callable`, req.GetMethod()))
	}

	if _, exists := productCrawlerType.MethodByName(req.GetMethod()); !exists {
		return nil, pbError.ErrNotFound.New(fmt.Sprintf(`method "%s" not found`, req.GetMethod()))
	}

	caller := cw.MethodByName(req.GetMethod())
	if caller.IsZero() {
		return nil, pbError.ErrNotFound.New(fmt.Sprintf(`method "%s" not found`, req.GetMethod()))
	}
	if caller.Kind() != reflect.Func {
		return nil, pbError.ErrInvalidArgument.New(fmt.Sprintf("%s is not callable", req.GetMethod()))
	}

	var (
		inArgCount  = caller.Type().NumIn()
		outArgCount = caller.Type().NumOut()
	)
	if inArgCount > 2 || outArgCount != 2 {
		return nil, pbError.ErrInvalidArgument.New(fmt.Sprintf(`method "%s" want more in arguments`, req.GetMethod()))
	}
	inputs := []reflect.Value{}
	switch inArgCount {
	case 0:
	case 1:
		inputs = append(inputs, reflect.ValueOf(shareCtx))
	case 2:
		inputs = append(inputs, reflect.ValueOf(shareCtx), reflect.ValueOf(req.GetInput()))
	}

	vals := caller.Call(inputs)
	if len(vals) != 2 {
		return nil, pbError.ErrInternal.New("caller response count not correct")
	}
	if !vals[1].IsNil() {
		val := vals[1].Interface()
		err := val.(error)
		s.logger.Error(err)
		return nil, pbError.ErrInternal.New(err)
	}

	val := vals[0]
	switch val.Kind() {
	case reflect.Slice, reflect.Array:
		var (
			size = val.Len()
			ret  pbCrawl.CallResponse
		)
		for i := 0; i < size; i++ {
			if v, err := marshalAny(val.Index(i)); err != nil {
				return nil, pbError.ErrInvalidArgument.New(err)
			} else {
				ret.Data = append(ret.Data, &pbCrawl.Item{
					Timestamp: time.Now().UnixNano() / 1000000,
					SiteId:    req.GetSiteId(),
					JobId:     req.GetJobId(),
					Index:     int32(i),
					Data:      v,
				})
			}
		}
		return &ret, nil
	case reflect.Ptr:
		if v, err := marshalAny(val); err != nil {
			return nil, pbError.ErrInvalidArgument.New(err)
		} else {
			item := &pbCrawl.Item{
				Timestamp: time.Now().UnixNano() / 1000000,
				SiteId:    req.GetSiteId(),
				JobId:     req.GetJobId(),
				Index:     1,
				Data:      v,
			}
			return &pbCrawl.CallResponse{Data: []*pbCrawl.Item{item}}, nil
		}
	default:
		s.logger.Debugf("%v %v", val.Interface(), val.Kind())
		return nil, pbError.ErrInternal.New("unsuported returned value")
	}
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
	shareCtx = context.WithValue(shareCtx, crawler.SiteIdKey, rawreq.GetSiteId())
	shareCtx = context.WithValue(shareCtx, crawler.TargetTypeKey, strings.Join(rawreq.GetOptions().GetTargetTypes(), ","))

	req, err := buildRequest(rawreq)
	if err != nil {
		logger.Error(err)
		return err
	}

	if !func() bool {
		for _, domain := range s.crawler.AllowedDomains() {
			if matched, _ := filepath.Match(domain, req.URL.Hostname()); matched {
				return true
			}
		}
		return false
	}() {
		logger.Infof("Access %s aborted", rawreq.GetUrl())
		return crawler.ErrAbort
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
		if resp.StatusCode != http.StatusOK {
			logger.Errorf("no response got status: %d", resp.StatusCode)
			return errors.New("no response got")
		}
		return crawler.ErrAbort
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

		if ctxutil.GetString(shareCtx, crawler.MainCategoryKey) != "" ||
			ctxutil.GetString(shareCtx, crawler.CategoryKey) != "" {
			switch val := i.(type) {
			case *http.Request:
				var (
					cateNames      = []string{}
					cateNameFilter = map[string]struct{}{}
				)
				for _, key := range crawler.CategoryKeys {
					vals := strings.Split(text.Clean(ctxutil.GetString(shareCtx, key)), ">")
					for _, val := range vals {
						val = strings.Title(strings.ToLower(val))
						if _, ok := cateNameFilter[val]; ok || val == "" {
							continue
						}
						// Ignore too long names, in most cases the long name is not the cate name
						if len([]rune(val)) > 48 {
							continue
						}
						cateNames = append(cateNames, val)
						cateNameFilter[val] = struct{}{}

						// keep length to 6
						if len(cateNames) == 6 {
							break
						}
					}
				}
				var cate *pbItem.Category
				for i := len(cateNames) - 1; i >= 0; i-- {
					name := cateNames[i]

					c := pbItem.Category{Name: name}
					if i == len(cateNames)-1 {
						c.Url = val.URL.String()
						c.Depth = int32(i + 1)
					}
					if cate != nil {
						c.Children = append(c.Children, cate)
					}
					cate = &c
				}
				i = cate
			}
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
				SiteId:        rawreq.GetSiteId(),
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
			val.SiteId = rawreq.GetSiteId()
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
