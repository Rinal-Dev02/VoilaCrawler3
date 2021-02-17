package request

import (
	"context"
	"errors"
	"time"

	nodeCtrl "github.com/voiladev/VoilaCrawl/internal/controller/node"
	reqManager "github.com/voiladev/VoilaCrawl/internal/model/request/manager"
	"github.com/voiladev/VoilaCrawl/internal/pkg/config"
	pbCrawl "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/redis"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"
)

type RequestController struct {
	ctx            context.Context
	nodeCtrl       *nodeCtrl.NodeController
	requestManager *reqManager.RequestManager
	redisClient    *redis.RedisClient
	logger         glog.Log
}

func NewRequestController(
	ctx context.Context,
	nodeCtrl *nodeCtrl.NodeController,
	requestManager *reqManager.RequestManager,
	redisClient *redis.RedisClient,
	logger glog.Log) (*RequestController, error) {

	if nodeCtrl == nil {
		return nil, errors.New("invalid node controller")
	}
	if requestManager == nil {
		return nil, errors.New("invalid requestManager")
	}
	if redisClient == nil {
		return nil, errors.New("invalid redis client")
	}

	ctrl := RequestController{
		ctx:            ctx,
		redisClient:    redisClient,
		requestManager: requestManager,
		nodeCtrl:       nodeCtrl,
		logger:         logger.New("RequestController"),
	}
	return &ctrl, nil
}

const defaultCheckTimeoutRequestInterval = time.Second * 5

func (ctrl *RequestController) Run(ctx context.Context) error {
	if ctrl == nil {
		return nil
	}

	if ctx == nil {
		return errors.New("invalidcontext")
	}

	var (
		err   error
		timer = time.NewTimer(defaultCheckTimeoutRequestInterval)
		data  []byte
		msg   pbCrawl.Command_Request
	)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			timer.Reset(defaultCheckTimeoutRequestInterval)

			reqs, err := ctrl.requestManager.List(ctx, nil, reqManager.ListRequest{
				Page: 1, Count: 1000,
				ExpireStatus: 2, Retryable: true,
			})
			if err != nil {
				ctrl.logger.Errorf("list request failed, error=%s", err)
				continue
			}

			for _, req := range reqs {
				if err := ctrl.nodeCtrl.PublishRequest(ctx, req); err != nil {
					ctrl.logger.Errorf("publish request failed, error=%s", err)
				}
			}
		default:
			node := ctrl.nodeCtrl.NextNode()
			if node == nil {
				time.Sleep(time.Millisecond * 100)
				continue
			}

			data, err = redis.Bytes(ctrl.redisClient.Do("RPOP", config.CrawlRequestQueue))
			if err == redis.ErrNil {
				time.Sleep(time.Millisecond * 200)
				continue
			} else if err != nil {
				ctrl.logger.Errorf("pop data from %s failed, error=%s", config.CrawlRequestQueue, err)
				time.Sleep(time.Millisecond * 200)
				continue
			}

			msg.Reset()
			if err = protojson.Unmarshal(data, &msg); err != nil {
				ctrl.logger.Errorf("unmarshal data failed, error=%s", err)
				continue
			}

			cmd := pbCrawl.Command{Timestamp: time.Now().UnixNano()}
			cmd.Data, _ = anypb.New(&msg)
			if err = node.Send(ctx, &cmd); err != nil {
				ctrl.logger.Errorf("send msg failed, error=%s", err)
				continue
			}

			if _, err := ctrl.requestManager.UpdateStatus(ctx, nil, msg.GetReqId(), 2, 0, false, ""); err != nil {
				ctrl.logger.Errorf("update status of request %s failed, error=%s", msg.GetReqId(), err)
			}
		}
	}
}
