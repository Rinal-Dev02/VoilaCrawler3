package main

import (
	"context"
	"errors"

	"github.com/nsqio/go-nsq"
	"github.com/voiladev/VoilaCrawl/internal/pkg/config"
	"github.com/voiladev/VoilaCrawl/pkg/crawler"
	pbHttp "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/http"
	pbCrawl "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl"
	"github.com/voiladev/go-framework/glog"
	pbError "github.com/voiladev/protobuf/protoc-gen-go/errors"
	"google.golang.org/protobuf/proto"
	anypb "google.golang.org/protobuf/types/known/anypb"
)

type CrawlerServer struct {
	pbCrawl.UnimplementedCrawlerManagerServer

	producer       *nsq.Producer
	crawlerManager *CrawlerManager
	crawlerCtrl    *CrawlerController
	logger         glog.Log
}

func NewCrawlerServer(crawlerCtrl *CrawlerController, crawlerManager *CrawlerManager, logger glog.Log) (pbCrawl.CrawlerManagerServer, error) {
	if crawlerCtrl == nil {
		return nil, errors.New("invalid crawler controller")
	}
	if crawlerManager == nil {
		return nil, errors.New("invalid crawler manager")
	}
	if logger == nil {
		return nil, errors.New("invalid logger")
	}
	s := CrawlerServer{
		crawlerCtrl:    crawlerCtrl,
		crawlerManager: crawlerManager,
		logger:         logger,
	}
	return &s, nil
}

func (s *CrawlerServer) GetCrawlers(ctx context.Context, req *pbCrawl.GetCrawlersRequest) (*pbCrawl.GetCrawlersResponse, error) {
	if req.GetHost() != "" {
		return nil, pbError.ErrInvalidArgument
	}

	crawlers, err := s.crawlerManager.GetByHost(ctx, req.GetHost())
	if err != nil {
		s.logger.Errorf("get crawlers of host %s failed, error=%s", req.GetHost(), err)
		return nil, err
	}
	var resp pbCrawl.GetCrawlersResponse
	for _, c := range crawlers {
		options := c.CrawlOptions()
		if options == nil {
			options = &crawler.CrawlOptions{}
		}
		item := pbCrawl.Crawler{
			Id:             c.GlobalID(),
			Version:        c.Version(),
			AllowedDomains: c.AllowedDomains(),
			Options: &pbCrawl.Crawler_Options{
				EnableHeadless:    options.EnableHeadless,
				EnableSessionInit: options.EnableSessionInit,
				KeepSession:       options.KeepSession,
				SessoinTtl:        int64(options.SessionTtl),
				DisableRedirect:   options.DisableRedirect,
				LoginRequired:     options.LoginRequired,
				Headers:           map[string]string{},
				Reliability:       options.Reliability,
			},
		}
		for k, vs := range options.MustHeader {
			v := ""
			if len(vs) > 0 {
				v = vs[0]
			}
			item.Options.Headers[k] = v
		}
		for _, c := range options.MustCookies {
			item.Options.Cookies = append(item.Options.Cookies, &pbHttp.Cookie{
				Name:   c.Name,
				Value:  c.Value,
				Domain: c.Domain,
				Path:   c.Path,
			})
		}
		resp.Data = append(resp.Data, &item)
	}
	return &resp, nil
}

func (s *CrawlerServer) Parse(ctx context.Context, req *pbCrawl.ParseRequest) (*pbCrawl.ParseResponse, error) {
	if s == nil {
		return nil, nil
	}
	logger := s.logger.New("Parse")

	var (
		shareCtx = ctx
	)
	for k, v := range req.GetRequest().SharingData {
		shareCtx = context.WithValue(shareCtx, k, v)
	}
	shareCtx = context.WithValue(shareCtx, "tracing_id", req.GetRequest().GetTracingId())
	shareCtx = context.WithValue(shareCtx, "job_id", req.GetRequest().GetJobId())
	shareCtx = context.WithValue(shareCtx, "req_id", req.GetRequest().GetReqId())
	shareCtx = context.WithValue(shareCtx, "store_id", req.GetRequest().GetStoreId())

	var ret []*anypb.Any
	if err := s.crawlerCtrl.Parse(shareCtx, req, func(ctx context.Context, msg proto.Message) error {
		if req.GetEnableBlockForItems() {
			data, err := anypb.New(msg)
			if err != nil {
				return err
			}
			ret = append(ret, data)
			return nil
		}

		var topic string
		switch msg.(type) {
		case *pbCrawl.Command_Request:
			topic = config.CrawlRequestTopic
		case *pbCrawl.Item:
			topic = config.CrawlItemProductTopic
		default:
			return errors.New("unsupported data type")
		}
		data, err := proto.Marshal(msg)
		if err != nil {
			return err
		}
		if err := s.producer.Publish(topic, data); err != nil {
			logger.Errorf("publish ret of %s failed, error=%s", req.GetRequest().GetUrl(), err)
			return err
		}
		return nil
	}); err != nil {
		logger.Errorf("parse response from %s failed, error=%s", req.GetRequest().GetUrl(), err)
		return nil, pbError.ErrInternal.New(err.Error())
	}
	return &pbCrawl.ParseResponse{Data: ret}, nil
}
