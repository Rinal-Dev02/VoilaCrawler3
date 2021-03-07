package request

import (
	"context"
	"errors"
	"time"

	nodeCtrl "github.com/voiladev/VoilaCrawl/internal/controller/node"
	threadCtrl "github.com/voiladev/VoilaCrawl/internal/controller/thread"
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
	threadCtrl     *threadCtrl.ThreadController
	requestManager *reqManager.RequestManager
	redisClient    *redis.RedisClient
	logger         glog.Log
}

func NewRequestController(
	ctx context.Context,
	nodeCtrl *nodeCtrl.NodeController,
	requestManager *reqManager.RequestManager,
	redisClient *redis.RedisClient,
	threadCtrl *threadCtrl.ThreadController,
	logger glog.Log) (*RequestController, error) {

	if nodeCtrl == nil {
		return nil, errors.New("invalid node controller")
	}
	if threadCtrl == nil {
		return nil, errors.New("invalid thread controller")
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
		threadCtrl:     threadCtrl,
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
		timer = time.NewTimer(defaultCheckTimeoutRequestInterval)
		data  []byte
		msg   pbCrawl.Command_Request
	)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				reqs, err := ctrl.requestManager.List(ctx, nil, reqManager.ListRequest{
					Page: 1, Count: 200, Retryable: true,
				})
				if err != nil {
					ctrl.logger.Errorf("list request failed, error=%s", err)
					timer.Reset(defaultCheckTimeoutRequestInterval)
					continue
				}
				ctrl.logger.Debugf("got %d requests", len(reqs))

				for _, req := range reqs {
					if err := ctrl.nodeCtrl.PublishRequest(ctx, req); err != nil {
						ctrl.logger.Errorf("publish request failed, error=%s", err)
					}
				}
				timer.Reset(defaultCheckTimeoutRequestInterval)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			ctrl.logger.Error(ctx.Err())
			return ctx.Err()
		default:
			node := ctrl.nodeCtrl.NextNode()
			if node == nil {
				time.Sleep(time.Millisecond * 100)
				continue
			}

			// 根据连接数情况缓存
			items, err := redis.ByteSlices(ctrl.redisClient.Do("BRPOP", config.CrawlRequestQueue, 0))
			if err == redis.ErrNil || err == redis.ErrTimeout {
				continue
			} else if err != nil {
				ctrl.logger.Errorf("pop data from %s failed, error=%s", config.CrawlRequestQueue, err)
				time.Sleep(time.Millisecond * 200)
				continue
			}
			data = items[1]

			msg.Reset()
			if err = protojson.Unmarshal(data, &msg); err != nil {
				ctrl.logger.Errorf("unmarshal data failed, error=%s", err)
				continue
			}

			// lock at most 15mins
			if ctrl.threadCtrl.Lock(ctrl.ctx, msg.GetHost(), msg.GetReqId(), int64(msg.GetOptions().GetMaxTtlPerRequest())) {
				cmd := pbCrawl.Command{Timestamp: time.Now().UnixNano()}
				cmd.Data, _ = anypb.New(&msg)

				ctrl.logger.Debugf("%s %s", msg.GetMethod(), msg.GetUrl())
				if err = node.Send(ctx, &cmd); err != nil {
					ctrl.threadCtrl.Unlock(ctrl.ctx, msg.GetHost(), msg.GetReqId())

					ctrl.logger.Errorf("send msg failed, error=%s", err)
					continue
				}
				if _, err := ctrl.requestManager.UpdateStatus(ctx, nil, msg.GetReqId(), 2, 0, false, ""); err != nil {
					ctrl.logger.Errorf("update status of request %s failed, error=%s", msg.GetReqId(), err)
				}
			} else if _, err = ctrl.redisClient.Do("LPUSH", config.CrawlRequestQueue, data); err != nil {
				ctrl.logger.Errorf("requeue request %s failed, error=%s", msg.GetReqId(), err)
				time.Sleep(time.Millisecond * 200)
			} else if _, err = ctrl.redisClient.Do("SREM", config.CrawlRequestQueueSet, msg.GetReqId()); err != nil {
				ctrl.logger.Errorf("remove req cache key failed, error=%s", err)
			}
		}
	}
}
