package node

import (
	"context"
	"errors"
	"time"

	"github.com/nsqio/go-nsq"
	crawlerManager "github.com/voiladev/VoilaCrawl/internal/model/crawler/manager"
	nodeManager "github.com/voiladev/VoilaCrawl/internal/model/node/manager"
	"github.com/voiladev/VoilaCrawl/internal/model/request"
	reqManager "github.com/voiladev/VoilaCrawl/internal/model/request/manager"
	"github.com/voiladev/VoilaCrawl/internal/pkg/config"
	pbEvent "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/event"
	pbCrawl "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/types/sortedmap"
	pbError "github.com/voiladev/protobuf/protoc-gen-go/errors"
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
	publisher      *nsq.Producer
	options        NodeControllerOptions
	logger         glog.Log
}

func NewNodeController(
	ctx context.Context,
	engine *xorm.Engine,
	nodeManager *nodeManager.NodeManager,
	crawlerManager *crawlerManager.CrawlerManager,
	requestManager *reqManager.RequestManager,
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
	if options == nil {
		return nil, errors.New("invalid options")
	}
	if options.HeartbeatInternal <= 100 {
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

func (ctrl *NodeController) Send(ctx context.Context, msg protoreflect.ProtoMessage) error {
	if ctrl == nil {
		return nil
	}

	var (
		handler *nodeHanadler
		maxIdle int32
	)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			ctrl.nodeHandlers.Range(func(key string, val interface{}) bool {
				select {
				case <-ctx.Done():
					return false
				default:
					h := val.(*nodeHanadler)
					if h.IdleConcurrency() >= maxIdle {
						handler = h
						maxIdle = h.IdleConcurrency()
					}
					return true
				}
			})
		}
		if handler != nil {
			break
		}
	}
	return handler.Send(ctx, msg)
}

func (ctrl *NodeController) PublishRequest(ctx context.Context, req *request.Request) error {
	if ctrl == nil || req == nil {
		return nil
	}

	var cmdReq pbCrawl.Command_Request
	if err := req.Unmarshal(&cmdReq); err != nil {
		return pbError.ErrInternal.New(err)
	}
	reqData, _ := anypb.New(&cmdReq)

	event := pbEvent.Event{
		Id:   req.GetId(),
		Type: pbEvent.EventType_Created,
		Headers: map[string]string{
			"Type-Url": reqData.GetTypeUrl(),
		},
		Data:      reqData.GetValue(),
		Timestamp: time.Now().UnixNano(),
	}
	data, _ := proto.Marshal(&event)
	return ctrl.publisher.Publish(config.CrawlRequestTopic, data)
}

func (ctrl *NodeController) PublishItem(ctx context.Context, item *pbCrawl.Command_Item) error {
	if ctrl == nil || item == nil || item.GetData() == nil {
		return nil
	}

	itemData, _ := anypb.New(item)
	event := pbEvent.Event{
		Id:   item.GetReqId(),
		Type: pbEvent.EventType_Created,
		Headers: map[string]string{
			"Type-Url": itemData.GetTypeUrl(),
		},
		Data:      itemData.GetValue(),
		Timestamp: time.Now().UnixNano(),
	}
	data, _ := proto.Marshal(&event)
	return ctrl.publisher.Publish(config.CrawledItemTopic, data)
}
