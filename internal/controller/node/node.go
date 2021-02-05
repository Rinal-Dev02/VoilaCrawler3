package node

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nsqio/go-nsq"
	crawlerManager "github.com/voiladev/VoilaCrawl/internal/model/crawler/manager"
	nodeManager "github.com/voiladev/VoilaCrawl/internal/model/node/manager"
	"github.com/voiladev/VoilaCrawl/internal/model/request"
	reqManager "github.com/voiladev/VoilaCrawl/internal/model/request/manager"
	"github.com/voiladev/VoilaCrawl/internal/pkg/config"
	pbEvent "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/event"
	pbCrawl "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl"
	pbItem "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/protoutil"
	"github.com/voiladev/go-framework/redis"
	"github.com/voiladev/go-framework/types/sortedmap"
	pbError "github.com/voiladev/protobuf/protoc-gen-go/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
	"xorm.io/xorm"
)

type NodeControllerOptions struct {
	HeartbeatInternal int64 // 单位毫秒
	NsqdAddr          string
}

// NodeController
type NodeController struct {
	ctx            context.Context
	engine         *xorm.Engine
	nodeManager    *nodeManager.NodeManager
	crawlerManager *crawlerManager.CrawlerManager
	requestManager *reqManager.RequestManager
	nodeHandlers   *sortedmap.SortedMap
	redisClient    *redis.RedisClient
	publisher      *nsq.Producer

	options NodeControllerOptions
	logger  glog.Log
}

func NewNodeController(
	ctx context.Context,
	engine *xorm.Engine,
	nodeManager *nodeManager.NodeManager,
	crawlerManager *crawlerManager.CrawlerManager,
	requestManager *reqManager.RequestManager,
	redisClient *redis.RedisClient,
	options *NodeControllerOptions,
	logger glog.Log,
) (*NodeController, error) {
	if engine == nil {
		return nil, errors.New("invalid engine")
	}
	if nodeManager == nil {
		return nil, errors.New("invalid node manager")
	}
	if crawlerManager == nil {
		return nil, errors.New("invalid crawler manager")
	}
	if requestManager == nil {
		return nil, errors.New("invalid request manager")
	}
	if redisClient == nil {
		return nil, errors.New("invalid redis client")
	}
	if options == nil {
		return nil, errors.New("invalid options")
	}
	if options.HeartbeatInternal < 50 {
		return nil, errors.New("heartbeat internal too short")
	}
	if options.NsqdAddr == "" {
		return nil, errors.New("invalid nsqd address")
	}
	c := NodeController{
		ctx:            ctx,
		engine:         engine,
		nodeManager:    nodeManager,
		crawlerManager: crawlerManager,
		requestManager: requestManager,
		redisClient:    redisClient,
		nodeHandlers:   sortedmap.New(),
		options:        *options,
		logger:         logger.New("NodeController"),
	}

	var (
		err  error
		conf = nsq.NewConfig()
	)
	if c.publisher, err = nsq.NewProducer(c.options.NsqdAddr, conf); err != nil {
		c.logger.Errorf("create nsq product failed, error=%s", err)
		return nil, err
	}
	return &c, nil
}

func (ctrl *NodeController) Register(ctx context.Context, conn pbCrawl.Gateway_ChannelServer) (*nodeHanadler, error) {
	if ctrl == nil {
		return nil, nil
	}
	logger := ctrl.logger.New("Register")

	handler, err := NewNodeHandler(ctx, ctrl, conn, ctrl.logger)
	if err != nil {
		logger.Errorf("instance NodeHandler failed, error=%s", err)
		return nil, err
	}
	ctrl.nodeHandlers.Set(handler.ID(), handler)

	return handler, nil
}

func (ctrl *NodeController) Unregister(ctx context.Context, id string) error {
	if ctrl == nil {
		return nil
	}
	val := ctrl.nodeHandlers.Get(id)
	if val == nil {
		return nil
	}

	h := val.(*nodeHanadler)
	ctrl.nodeHandlers.Delete(id)
	ctrl.nodeManager.Delete(ctx, h.node.GetId())
	return nil
}

func (ctrl *NodeController) PublishRequest(ctx context.Context, req *request.Request) error {
	if ctrl == nil || req == nil {
		return nil
	}
	logger := ctrl.logger.New("PublishRequest")

	var cmdReq pbCrawl.Command_Request
	if err := req.Unmarshal(&cmdReq); err != nil {
		logger.Errorf("unmarshal command request failed, error=%s", err)
		return pbError.ErrInternal.New(err)
	}

	reqData, err := protojson.Marshal(&cmdReq)
	if err != nil {
		logger.Errorf("marshal Command_Request failed, error=%s", err)
		return pbError.ErrInternal.New(err)
	}

	if _, err := ctrl.redisClient.Do("LPUSH", config.CrawlRequestQueue, reqData); err != nil {
		logger.Errorf("lpush request failed, error=%s", err)
		return pbError.ErrDatabase.New(err)
	}
	return nil
}

func (ctrl *NodeController) PublishItem(ctx context.Context, item *pbCrawl.Item) error {
	if ctrl == nil {
		return nil
	}
	logger := ctrl.logger.New("PublishItem")

	if item == nil || item.GetData() == nil {
		return errors.New("empty item data")
	}

	itemData, _ := anypb.New(item)

	switch item.GetData().GetTypeUrl() {
	case protoutil.GetTypeUrl(&pbItem.Product{}):
		event := pbEvent.Event{
			Id:   item.GetReqId(),
			Type: pbEvent.EventType_Created,
			Headers: map[string]string{
				"Type-Url": itemData.GetTypeUrl(),
				"Datetime": time.Now().Format(time.RFC3339Nano),
			},
			Data:      itemData.GetValue(),
			Timestamp: time.Now().UnixNano(),
		}
		data, _ := proto.Marshal(&event)
		if err := ctrl.publisher.Publish(config.CrawlItemProductTopic, data); err != nil {
			logger.Errorf("publish item failed, error=%s", err)
			return pbError.ErrInternal.New(err)
		}
	default:
		data, _ := proto.Marshal(item)
		retKey := fmt.Sprintf("fetch://tracing/%s", item.GetTracingId())
		if _, err := ctrl.redisClient.Do("LPUSH", retKey, data); err != nil {
			logger.Errorf("cache item data failed, error=%s", err)
			return pbError.ErrInternal.New(err)
		}
		if _, err := ctrl.redisClient.Do("EXPIRES", retKey, 3600); err != nil {
			logger.Errorf("update key expires ttl failed, error=%s", err)
			return pbError.ErrInternal.New(err)
		}
	}
	return nil
}

// Broadcast
func (ctrl *NodeController) broadcast(ctx context.Context, msg protoreflect.ProtoMessage) error {
	if ctrl == nil || msg == nil {
		return nil
	}

	switch v := msg.(type) {
	case *pbCrawl.Command_Request:
		cmd := pbCrawl.Command{Timestamp: time.Now().UnixNano()}
		cmd.Data, _ = anypb.New(v)
		msg = &cmd
	}
	ctrl.nodeHandlers.Range(func(key string, val interface{}) bool {
		select {
		case <-ctx.Done():
			return false
		default:
			h := val.(*nodeHanadler)
			if h.IsInited() {
				h.Send(ctx, msg)
			}
			return true
		}
	})
	return nil
}

// Send
func (ctrl *NodeController) Send(ctx context.Context, msg protoreflect.ProtoMessage, broadcast bool) error {
	if ctrl == nil {
		return nil
	}

	if broadcast {
		return ctrl.broadcast(ctx, msg)
	}

	var handler *nodeHanadler
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// find the most idle node
			ctrl.nodeHandlers.Range(func(key string, val interface{}) bool {
				h := val.(*nodeHanadler)

				if !h.IsInited() {
					return true
				}

				handler = h
				return false
			})

			if handler != nil {
				// get one request
				switch v := msg.(type) {
				case *pbCrawl.Command_Request:
					cmd := pbCrawl.Command{Timestamp: time.Now().UnixNano()}
					cmd.Data, _ = anypb.New(v)
					msg = &cmd
				}
				return handler.Send(ctx, msg)
			}
			time.Sleep(time.Millisecond * 200)
		}
	}
}

// NextNode
func (ctrl *NodeController) NextNode() *nodeHanadler {
	if ctrl == nil {
		return nil
	}
	var (
		handler *nodeHanadler
		maxIdle int32 = 0
	)
	ctrl.nodeHandlers.Range(func(key string, val interface{}) bool {
		h := val.(*nodeHanadler)
		if !h.IsInited() {
			return true
		}

		idleCount := h.IdleConcurrency()
		if idleCount > maxIdle {
			handler = h
			maxIdle = idleCount
		}
		return true
	})
	return handler
}
