package manager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/nsqio/go-nsq"
	crawlerCtrl "github.com/voiladev/VoilaCrawl/internal/controller/crawler"
	storeCtrl "github.com/voiladev/VoilaCrawl/internal/controller/store"
	"github.com/voiladev/VoilaCrawl/internal/model/crawler"
	crawlerManager "github.com/voiladev/VoilaCrawl/internal/model/crawler/manager"
	pbCrawl "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl"
	_ "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/protoutil"
	pbError "github.com/voiladev/protobuf/protoc-gen-go/errors"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type CrawlerServer struct {
	pbCrawl.UnimplementedCrawlerManagerServer
	pbCrawl.UnimplementedCrawlerRegisterServer

	storeCtrl      *storeCtrl.StoreController
	crawlerCtrl    *crawlerCtrl.CrawlerController
	crawlerManager *crawlerManager.CrawlerManager
	producer       *nsq.Producer
	logger         glog.Log
}

func NewCrawlerServer(storeCtrl *storeCtrl.StoreController, crawlerCtrl *crawlerCtrl.CrawlerController,
	crawlerManager *crawlerManager.CrawlerManager, logger glog.Log) (pbCrawl.CrawlerManagerServer, pbCrawl.CrawlerRegisterServer, error) {
	if storeCtrl == nil {
		return nil, nil, errors.New("invalid store controller")
	}
	if crawlerCtrl == nil {
		return nil, nil, errors.New("invalid crawler controller")
	}
	if crawlerManager == nil {
		return nil, nil, errors.New("invalid crawler manager")
	}
	if logger == nil {
		return nil, nil, errors.New("invalid logger")
	}
	s := CrawlerServer{
		storeCtrl:      storeCtrl,
		crawlerCtrl:    crawlerCtrl,
		crawlerManager: crawlerManager,
		logger:         logger.New("CrawlerServer"),
	}
	return &s, &s, nil
}

func (s *CrawlerServer) GetCrawlers(ctx context.Context, req *pbCrawl.GetCrawlersRequest) (*pbCrawl.GetCrawlersResponse, error) {
	if s == nil {
		return nil, nil
	}

	crawlers, _ := s.crawlerManager.List(ctx)
	resp := pbCrawl.GetCrawlersResponse{
		Data: map[string]*pbCrawl.GetCrawlersResponse_CrawlerGroup{},
	}
	for storeId, cws := range crawlers {
		group := pbCrawl.GetCrawlersResponse_CrawlerGroup{
			StoreId: storeId,
		}
		if nodeStatus, err := s.crawlerManager.GetStatus(ctx, "", storeId); err != nil {
			s.logger.Error(err)
		} else {
			for _, status := range nodeStatus {
				group.Status = append(group.Status, &pbCrawl.NodeStatus{
					Host:                 status.Hostname,
					MaxAPIConcurrency:    status.MaxAPIConcurrency,
					MaxMQConcurrency:     status.MaxMQConcurrency,
					CurrentConcurrency:   status.CurrentConcurrency,
					CurrentMQConcurrency: status.CurrentMQConcurrency,
				})
			}
		}

		for _, cw := range cws {
			var item pbCrawl.Crawler
			cw.Unmarshal(&item)
			group.Crawlers = append(group.Crawlers, &item)
		}
		resp.Data[storeId] = &group
	}
	return &resp, nil
}

func (s *CrawlerServer) GetCrawler(ctx context.Context, req *pbCrawl.GetCrawlerRequest) (*pbCrawl.GetCrawlerResponse, error) {
	if s == nil {
		return nil, nil
	}

	crawlers, _ := s.crawlerManager.GetByStore(ctx, req.GetStoreId())

	var ret pbCrawl.GetCrawlerResponse
	for _, cw := range crawlers {
		var item pbCrawl.Crawler
		cw.Unmarshal(&item)
		ret.Data = append(ret.Data, &item)
	}

	if nodeStatus, err := s.crawlerManager.GetStatus(ctx, "", req.GetStoreId()); err != nil {
		s.logger.Error(err)
	} else {
		for _, status := range nodeStatus {
			ret.Status = append(ret.Status, &pbCrawl.NodeStatus{
				Host:                 status.Hostname,
				MaxAPIConcurrency:    status.MaxAPIConcurrency,
				MaxMQConcurrency:     status.MaxMQConcurrency,
				CurrentConcurrency:   status.CurrentConcurrency,
				CurrentMQConcurrency: status.CurrentMQConcurrency,
			})
		}
	}

	return &ret, nil
}

func (s *CrawlerServer) GetCrawlerOptions(ctx context.Context, req *pbCrawl.GetCrawlerOptionsRequest) (*pbCrawl.GetCrawlerOptionsResponse, error) {
	if s == nil {
		return nil, nil
	}
	logger := s.logger.New("GetCrawlerOptions")

	opts, err := s.crawlerCtrl.CrawlerOptions(ctx, req.GetStoreId(), req.GetUrl())
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	return &pbCrawl.GetCrawlerOptionsResponse{Data: opts}, nil
}

func (s *CrawlerServer) GetCanonicalUrl(ctx context.Context, req *pbCrawl.GetCanonicalUrlRequest) (*pbCrawl.GetCanonicalUrlResponse, error) {
	if s == nil {
		return nil, nil
	}
	logger := s.logger.New("GetCanonicalUrl")

	u, err := s.crawlerCtrl.CanonicalUrl(ctx, req.GetStoreId(), req.GetUrl())
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	return &pbCrawl.GetCanonicalUrlResponse{
		Data: &pbCrawl.GetCanonicalUrlResponse_Data{Url: u},
	}, nil
}

func (s *CrawlerServer) DoParse(ctx context.Context, req *pbCrawl.DoParseRequest) (*pbCrawl.DoParseResponse, error) {
	if s == nil {
		return nil, nil
	}
	logger := s.logger.New("DoParse")

	ttl := time.Duration(req.GetRequest().GetOptions().GetMaxTtlPerRequest()) * time.Second
	shareCtx, cancel := context.WithTimeout(ctx, ttl)
	defer cancel()

	for k, v := range req.GetRequest().SharingData {
		shareCtx = context.WithValue(shareCtx, k, v)
	}
	shareCtx = context.WithValue(shareCtx, "tracing_id", req.GetRequest().GetTracingId())
	shareCtx = context.WithValue(shareCtx, "job_id", req.GetRequest().GetJobId())
	shareCtx = context.WithValue(shareCtx, "req_id", req.GetRequest().GetReqId())
	shareCtx = context.WithValue(shareCtx, "store_id", req.GetRequest().GetStoreId())

	var (
		resp      pbCrawl.DoParseResponse
		reqFilter = map[string]struct{}{}
		subReqs   = make(chan *pbCrawl.Request, 100)
	)
	defer func() {
		close(subReqs)
		// clean
		for range subReqs {
		}
	}()

	subReqs <- req.GetRequest()
	reqFilter[req.GetRequest().GetUrl()] = struct{}{}

end:
	for {
		select {
		case <-ctx.Done():
			break end
		case r, ok := <-subReqs:
			if !ok {
				break end
			}
			if itemCount, subReqCount, err := s.storeCtrl.Parse(shareCtx, req.GetRequest().GetStoreId(), r,
				func(ctx context.Context, v proto.Message) error {
					if v == nil {
						return nil
					}

					switch vv := v.(type) {
					case *pbCrawl.Request:
						if _, ok := reqFilter[vv.String()]; !ok {
							select {
							case subReqs <- vv:
								reqFilter[vv.String()] = struct{}{}
							default:
								logger.Errorf("too may sub requests for %s, ignored", r.GetUrl())
							}
						}
					case *pbCrawl.Item:
						data, _ := anypb.New(vv)
						resp.Data = append(resp.Data, data)
						if len(resp.Data) >= int(req.GetRequest().Options.GetMaxItemCount()) {
							return crawlerCtrl.ErrCountMatched
						}
					}
					return nil
				},
			); err != nil && !errors.Is(err, crawlerCtrl.ErrCountMatched) {
				logger.Errorf("parse response from %s failed, error=%v", req.GetRequest().GetUrl(), err)
				return nil, pbError.ErrInternal.New(err.Error())
			} else {
				resp.ItemCount += int32(itemCount)
				resp.SubReqCount += int32(subReqCount)
			}
		default:
			break end
		}
	}
	return &resp, nil
}

func (s *CrawlerServer) Connect(srv pbCrawl.CrawlerRegister_ConnectServer) (err error) {
	if s == nil {
		return nil
	}
	logger := s.logger.New("Connect")

	var (
		ip  string
		ctx = srv.Context()
	)

	if peer, _ := peer.FromContext(ctx); peer != nil {
		ip, _, _ = net.SplitHostPort(peer.Addr.String())
	} else {
		return fmt.Errorf("get peer info failed")
	}

	anyData, err := srv.Recv()
	if err != nil {
		if err == io.EOF {
			logger.Debugf("connect EOF")
			return nil
		}
		logger.Errorf("receive pkg failed, error=%s", err)
		return err
	}

	if anyData.GetTypeUrl() != protoutil.GetTypeUrl(&pbCrawl.ConnectRequest_Ping{}) {
		logger.Errorf("crawler node not registered yet.")
		return pbError.ErrFailedPrecondition.New("crawler node not registered")
	}

	var data pbCrawl.ConnectRequest_Ping
	if err := proto.Unmarshal(anyData.GetValue(), &data); err != nil {
		logger.Error(err)
		return err
	}

	cw, err := crawler.NewCrawler(&data, ip)
	if err != nil {
		logger.Error(err)
		return err
	}
	return s.crawlerCtrl.Watch(ctx, srv, cw)
}
