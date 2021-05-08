package v1

import (
	"context"
	"errors"
	"time"

	reqCtrl "github.com/voiladev/VoilaCrawl/internal/controller/request"
	"github.com/voiladev/VoilaCrawl/internal/model/request"
	hisManager "github.com/voiladev/VoilaCrawl/internal/model/request/history/manager"
	reqManager "github.com/voiladev/VoilaCrawl/internal/model/request/manager"
	pbCrawl "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/protoutil"
	"github.com/voiladev/go-framework/redis"
	"github.com/voiladev/go-framework/timeutil"
	pbError "github.com/voiladev/protobuf/protoc-gen-go/errors"
	"xorm.io/xorm"
)

type GatewayServer struct {
	pbCrawl.UnimplementedGatewayServer

	ctx            context.Context
	engine         *xorm.Engine
	redisClient    *redis.RedisClient
	requestManager *reqManager.RequestManager
	hisManager     *hisManager.HistoryManager
	requestCtrl    *reqCtrl.RequestController
	crawlerClient  pbCrawl.CrawlerManagerClient
	logger         glog.Log
}

func NewGatewayServer(
	ctx context.Context,
	engine *xorm.Engine,
	crawlerClient pbCrawl.CrawlerManagerClient,
	redisClient *redis.RedisClient,
	requestCtrl *reqCtrl.RequestController,
	requestManager *reqManager.RequestManager,
	hisManager *hisManager.HistoryManager,
	logger glog.Log,
) (pbCrawl.GatewayServer, error) {
	if crawlerClient == nil {
		return nil, errors.New("invalid crawler client")
	}
	if requestCtrl == nil {
		return nil, errors.New("invalid request controller")
	}
	if redisClient == nil {
		return nil, errors.New("invalid redis client")
	}
	if requestManager == nil {
		return nil, errors.New("invalid request manager")
	}
	if hisManager == nil {
		return nil, errors.New("invalid history manager")
	}
	if logger == nil {
		return nil, errors.New("invalid logger")
	}

	s := GatewayServer{
		ctx:            ctx,
		engine:         engine,
		crawlerClient:  crawlerClient,
		requestCtrl:    requestCtrl,
		redisClient:    redisClient,
		requestManager: requestManager,
		hisManager:     hisManager,
		logger:         logger.New("GatewayServer"),
	}
	return &s, nil
}

// Crawlers
func (s *GatewayServer) GetCrawlers(ctx context.Context, req *pbCrawl.GetCrawlersRequest) (*pbCrawl.GetCrawlersResponse, error) {
	if s == nil {
		return nil, nil
	}
	return s.crawlerClient.GetCrawlers(ctx, req)
}

// GetCrawler
func (s *GatewayServer) GetCrawler(ctx context.Context, req *pbCrawl.GetCrawlerRequest) (*pbCrawl.GetCrawlerResponse, error) {
	if s == nil {
		return nil, nil
	}
	return s.crawlerClient.GetCrawler(ctx, req)
}

// GetCanonicalUrl
func (s *GatewayServer) GetCanonicalUrl(ctx context.Context, req *pbCrawl.GetCanonicalUrlRequest) (*pbCrawl.GetCanonicalUrlResponse, error) {
	if s == nil {
		return nil, nil
	}
	return s.crawlerClient.GetCanonicalUrl(ctx, req)
}

// GetRequest
func (s *GatewayServer) GetRequest(ctx context.Context, req *pbCrawl.GetRequestRequest) (*pbCrawl.GetRequestResponse, error) {
	if s == nil {
		return nil, nil
	}
	if req.GetId() == "" {
		return nil, pbError.ErrInvalidArgument.New("invalid request id")
	}
	rawreq, err := s.requestManager.GetById(ctx, nil, req.GetId())
	if err != nil {
		s.logger.Error(err)
		return nil, err
	}

	var r pbCrawl.Request
	if err := rawreq.Unmarshal(&r); err != nil {
		return nil, pbError.ErrInternal.New(err)
	}
	data := pbCrawl.GetRequestResponse_Data{
		RetryCount: rawreq.GetRetryCount(),
		Status:     pbCrawl.GetRequestResponse_RequestStatus(rawreq.GetStatus()),
		IsSucceed:  rawreq.GetIsSucceed(),
		Request:    &r,
	}
	return &pbCrawl.GetRequestResponse{Data: &data}, nil
}

// GetCrawlerLogs
func (s *GatewayServer) GetCrawlerLogs(ctx context.Context, req *pbCrawl.GetCrawlerLogsRequest) (*pbCrawl.GetCrawlerLogsResponse, error) {
	if s == nil {
		return nil, nil
	}
	logger := s.logger.New("GetCrawlerLogs")

	var (
		startTime *time.Time
		endTime   *time.Time
	)
	if req.GetStartTime() != "" {
		t := timeutil.TimeParse(req.GetStartTime())
		if t.IsZero() {
			logger.Errorf("unable to parse time %s", req.GetStartTime())
			return nil, pbError.ErrInvalidArgument.New("unsupported time format")
		}
		startTime = &t
	}
	if req.GetEndTime() != "" {
		t := timeutil.TimeParse(req.GetEndTime())
		if t.IsZero() {
			logger.Errorf("unable to parse time %s", req.GetEndTime())
			return nil, pbError.ErrInvalidArgument.New("unsupported time format")
		}
		endTime = &t
	}
	if startTime != nil && endTime != nil {
		if startTime.After(*endTime) {
			return nil, pbError.ErrInvalidArgument.New("start time must before end time")
		}
	}

	listReq := hisManager.ListRequest{
		Limit:   req.GetLimit(),
		Cursor:  req.GetCursor(),
		StoreId: req.GetStoreId(),
		JobId:   req.GetJobId(),
		Order:   req.GetOrder(),
	}
	if startTime != nil {
		listReq.StartUtc = startTime.UnixNano() / 1000000
	}
	if endTime != nil {
		listReq.EndUtc = endTime.UnixNano() / 1000000
	}

	listResp, err := s.hisManager.List(ctx, listReq)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	resp := pbCrawl.GetCrawlerLogsResponse{}
	if listResp.Cursor != "" {
		resp.Pagination = &pbCrawl.GetCrawlerLogsResponse_Pagination{Cursor: listResp.Cursor}
	}
	for _, item := range listResp.Data {
		resp.Data = append(resp.Data, &pbCrawl.Error{
			ReqId:     item.GetId(),
			Timestamp: item.GetTimestamp(),
			JobId:     item.GetJobId(),
			TracingId: item.GetTracingId(),
			ErrMsg:    item.GetErrMsg(),
			Code:      item.GetCode(),
			Duration:  int64(item.GetDuration()),
		})
	}
	return &resp, nil
}

func (s *GatewayServer) Fetch(ctx context.Context, req *pbCrawl.FetchRequest) (*pbCrawl.FetchResponse, error) {
	if s == nil {
		return nil, nil
	}
	logger := s.logger.New("Fetch")

	r, err := request.NewRequest(req)
	if err != nil {
		logger.Errorf("load request failed, error=%s", err)
		return nil, pbError.ErrInvalidArgument.New(err)
	}
	if r.GetJobId() == "" {
		return nil, pbError.ErrInvalidArgument.New("invalid job id")
	}

	session := s.engine.NewSession()
	defer session.Close()

	resp := pbCrawl.FetchResponse{}
	if req.GetOptions().GetEnableBlockForItems() ||
		req.GetEnableBlockForItems() {

		r.Status = 3
		r.IsSucceed = true // disable retry
		if r, err = s.requestManager.Create(ctx, session, r); err != nil && err != pbError.ErrAlreadyExists {
			logger.Errorf("save request failed, error=%s", err)
			return nil, err
		}

		var creq pbCrawl.Request
		if err := r.Unmarshal(&creq); err != nil {
			logger.Errorf("unmarshal request failed, error=%s", err)
			return nil, pbError.ErrInternal.New(err)
		}
		resp, err := s.crawlerClient.DoParse(ctx, &pbCrawl.DoParseRequest{
			Request:             &creq,
			EnableBlockForItems: true,
		})
		if err != nil {
			logger.Error(err)
			return nil, err
		}
		for _, item := range resp.GetData() {
			if item.GetTypeUrl() != protoutil.GetTypeUrl(&pbCrawl.Item{}) {
				continue
			}
			resp.Data = append(resp.Data, item)
		}
	} else {
		if r, err = s.requestManager.Create(ctx, session, r); err != nil && err != pbError.ErrAlreadyExists {
			logger.Errorf("save request failed, error=%s", err)
			return nil, err
		}

		if err := session.Begin(); err != nil {
			logger.Errorf("begin tx failed, error=%s", err)
			return nil, pbError.ErrDatabase.New("begin tx failed")
		}
		if err := s.requestCtrl.PublishRequest(ctx, session, r, true); err != nil {
			logger.Errorf("publish request failed, error=%s", err)
			session.Rollback()
			return nil, err
		}
		if err := session.Commit(); err != nil {
			logger.Errorf("commit tx failed, error=%s", err)
			return nil, pbError.ErrDatabase.New("commit tx failed")
		}
	}
	return &resp, nil
}
