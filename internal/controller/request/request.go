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

const (
	defaultCheckTimeoutRequestInterval = time.Second * 5

	// KEYS[1]-Stores, KEYS[2]-StoreQueue
	// ARGV[1]-reqId, ARGV[2]-req
	requestPushScript = `local ret = redis.call("LPUSH", KEYS[2], ARGV[2])
local count = redis.call("LLEN", KEYS[2])
redis.call("ZADD", KEYS[1], count, KEYS[2])
return ret`

	// KEYS[1]-Stores, KEYS[2]-StoreQueue
	requestPopScript = `local ret = redis.call("RPOP", KEYS[2])
local count = redis.call("LLEN", KEYS[2])
if count == nil or count == 0 then
    redis.call("ZREM", KEYS[1], KEYS[2])
else
    redis.call("ZADD", KEYS[1], count, KEYS[2])
end
return ret`
)

func (ctrl *RequestController) Run(ctx context.Context) error {
	if ctrl == nil {
		return nil
	}

	if ctx == nil {
		return errors.New("invalidcontext")
	}

	var (
		timer = time.NewTimer(defaultCheckTimeoutRequestInterval)
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

			stores, err := redis.Strings(ctrl.redisClient.Do("ZRANGE", config.CrawlStoreList, 0, -1))
			if err != nil {
				ctrl.logger.Errorf("get queues failed, error=%s", err)
				time.Sleep(time.Millisecond * 200)
				continue
			}

			isSend := false
			for _, key := range stores {
				for i := 0; i < 3; i++ {
					data, err := redis.Bytes(ctrl.redisClient.Do("EVAL", requestPopScript, 2, config.CrawlStoreList, key))
					if err == redis.ErrNil {
						break
					} else if err != nil {
						ctrl.logger.Errorf("pop data from %s failed, error=%s", key, err)
						continue
					}

					msg.Reset()
					if err = protojson.Unmarshal(data, &msg); err != nil {
						ctrl.logger.Errorf("unmarshal data failed, error=%s", err)
						continue
					}

					// lock at most 5mins
					if ctrl.threadCtrl.Lock(ctrl.ctx, msg.GetStoreId(), msg.GetReqId(), int64(msg.GetOptions().GetMaxTtlPerRequest())) {
						ctrl.logger.Debugf("%s %s", msg.GetMethod(), msg.GetUrl())

						cmd := pbCrawl.Command{Timestamp: time.Now().UnixNano()}
						cmd.Data, _ = anypb.New(&msg)
						if err = node.Send(ctx, &cmd); err != nil {
							ctrl.threadCtrl.Unlock(ctrl.ctx, msg.GetStoreId(), msg.GetReqId())

							ctrl.logger.Errorf("send msg failed, error=%s", err)

							if _, err := ctrl.redisClient.Do("EVAL", requestPushScript, 2,
								config.CrawlStoreList, key, msg.GetReqId(), data); err != nil {
								ctrl.logger.Errorf("requeue request %s failed, error=%s", msg.GetReqId(), err)
							}
							continue
						}
						isSend = true

						if _, err := ctrl.redisClient.Do("SREM", config.CrawlRequestQueueSet, msg.GetReqId()); err != nil {
							ctrl.logger.Errorf("remove req cache key failed, error=%s", err)
						}
						if _, err := ctrl.requestManager.UpdateStatus(ctx, nil, msg.GetReqId(), 2, 0, false, ""); err != nil {
							ctrl.logger.Errorf("update status of request %s failed, error=%s", msg.GetReqId(), err)
						}

						break
					} else if _, err := ctrl.redisClient.Do("EVAL", requestPushScript, 2,
						config.CrawlStoreList, key, msg.GetReqId(), data); err != nil {
						ctrl.logger.Errorf("requeue request %s failed, error=%s", msg.GetReqId(), err)
					}
				}
				if isSend {
					break
				}
			}

			if len(stores) == 0 {
				time.Sleep(time.Millisecond * 100)
			}
		}
	}
}
