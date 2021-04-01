package v1

import (
	"context"
	"errors"
	"fmt"
	"time"

	reqCtrl "github.com/voiladev/VoilaCrawl/internal/controller/request"
	"github.com/voiladev/VoilaCrawl/internal/model/request"
	reqManager "github.com/voiladev/VoilaCrawl/internal/model/request/manager"
	pbCrawl "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/redis"
	pbError "github.com/voiladev/protobuf/protoc-gen-go/errors"
	"google.golang.org/protobuf/proto"
	"xorm.io/xorm"
)

type GatewayServer struct {
	pbCrawl.UnimplementedGatewayServer

	ctx            context.Context
	engine         *xorm.Engine
	redisClient    *redis.RedisClient
	requestManager *reqManager.RequestManager
	requestCtrl    *reqCtrl.RequestController
	logger         glog.Log
}

func NewGatewayServer(
	ctx context.Context,
	engine *xorm.Engine,
	requestCtrl *reqCtrl.RequestController,
	redisClient *redis.RedisClient,
	requestManager *reqManager.RequestManager,
	logger glog.Log,
) (pbCrawl.GatewayServer, error) {
	if engine == nil {
		return nil, errors.New("invalid xorm engine")
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
	s := GatewayServer{
		ctx:            ctx,
		engine:         engine,
		requestCtrl:    requestCtrl,
		redisClient:    redisClient,
		requestManager: requestManager,
		logger:         logger.New("GatewayServer"),
	}
	return &s, nil
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

	// TODO: change this to an api
	if req.GetOptions().GetEnableBlockForItems() {
		ttl := req.GetOptions().GetMaxBlockTtl()
		if ttl <= 0 {
			ttl = 300
		}

		nctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(ttl))
		defer cancel()

		var (
			resp   pbCrawl.FetchResponse
			retKey = fmt.Sprintf("fetch://tracing/%s", r.GetTracingId())
		)
		for req.GetOptions().GetMaxItemCount() == 0 ||
			len(resp.GetData()) < int(req.GetOptions().GetMaxItemCount()) {

			select {
			case <-nctx.Done():
				return &resp, nil
			default:
				if data, err := redis.Bytes(s.redisClient.Do("RPOP", retKey)); err == redis.ErrNil {
					time.Sleep(time.Millisecond * 200)
					continue
				} else if err != nil {
					logger.Errorf("check results of key %s failed, error=%s", retKey, err)
					time.Sleep(time.Millisecond * 200)
					continue
				} else {
					var item pbCrawl.Item
					if err := proto.Unmarshal(data, &item); err != nil {
						logger.Errorf("unmarshal data to any type failed, error=%s", err)
						continue
					}
					resp.Data = append(resp.Data, &item)
				}
			}
		}
		return &resp, nil
	}
	return &pbCrawl.FetchResponse{}, nil
}
