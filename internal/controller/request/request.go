package request

import (
	"context"
	"errors"
	"time"

	nodeCtrl "github.com/voiladev/VoilaCrawl/internal/controller/node"
	"github.com/voiladev/VoilaCrawl/internal/pkg/config"
	pbCrawl "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/redis"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"
)

type RequestController struct {
	ctx         context.Context
	nodeCtrl    *nodeCtrl.NodeController
	redisClient *redis.RedisClient
	logger      glog.Log
}

func NewRequestController(
	ctx context.Context,
	nodeCtrl *nodeCtrl.NodeController,
	redisClient *redis.RedisClient,
	logger glog.Log) (*RequestController, error) {
	if redisClient == nil {
		return nil, errors.New("invalid redis client")
	}
	ctrl := RequestController{
		ctx:         ctx,
		redisClient: redisClient,
		nodeCtrl:    nodeCtrl,
		logger:      logger.New("RequestController"),
	}
	return &ctrl, nil
}

func (ctrl *RequestController) Run(ctx context.Context) error {
	if ctrl == nil {
		return nil
	}

	if ctx == nil {
		return errors.New("invalidcontext")
	}

	var (
		err  error
		data []byte
		msg  pbCrawl.Command_Request
	)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
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
		}
	}
}
