package v1

import (
	"context"
	"errors"
	"fmt"
	"time"

	nodeCtrl "github.com/voiladev/VoilaCrawl/internal/controller/node"
	"github.com/voiladev/VoilaCrawl/internal/model/request"
	reqManager "github.com/voiladev/VoilaCrawl/internal/model/request/manager"
	pbCrawl "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/redis"
	pbError "github.com/voiladev/protobuf/protoc-gen-go/errors"
	"google.golang.org/protobuf/proto"
)

type GatewayServer struct {
	pbCrawl.UnimplementedGatewayServer

	ctx            context.Context
	nodeCtrl       *nodeCtrl.NodeController
	redisClient    *redis.RedisClient
	requestManager *reqManager.RequestManager
	logger         glog.Log
}

func NewGatewayServer(
	ctx context.Context,
	nodeCtrl *nodeCtrl.NodeController,
	redisClient *redis.RedisClient,
	requestManager *reqManager.RequestManager,
	logger glog.Log,
) (pbCrawl.GatewayServer, error) {
	if nodeCtrl == nil {
		return nil, errors.New("invalid node controller")
	}
	if redisClient == nil {
		return nil, errors.New("invalid redis client")
	}
	if requestManager == nil {
		return nil, errors.New("invalid request manager")
	}
	s := GatewayServer{
		ctx:            ctx,
		nodeCtrl:       nodeCtrl,
		redisClient:    redisClient,
		requestManager: requestManager,
		logger:         logger.New("GatewayServer"),
	}
	return &s, nil
}

func (s *GatewayServer) Channel(cs pbCrawl.Gateway_ChannelServer) error {
	if s == nil {
		return nil
	}
	logger := s.logger.New("Channel")

	handler, err := s.nodeCtrl.Register(cs.Context(), cs)
	if err != nil {
		logger.Error(err)
		return err
	}

	defer func() {
		s.nodeCtrl.Unregister(cs.Context(), handler.ID())
	}()

	if err := handler.Run(); err != nil {
		logger.Error(err)
		return err
	}
	return nil
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
	if r, err = s.requestManager.Create(ctx, nil, r); err != nil {
		logger.Errorf("save request failed, error=%s", err)
		return nil, err
	}
	if err = s.nodeCtrl.PublishRequest(ctx, r); err != nil {
		logger.Error(err)
		return nil, err
	}

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
