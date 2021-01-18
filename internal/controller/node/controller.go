package node

import (
	"context"

	crawlerManager "github.com/voiladev/VoilaCrawl/internal/model/crawler/manager"
	nodeManager "github.com/voiladev/VoilaCrawl/internal/model/node/manager"
	reqManager "github.com/voiladev/VoilaCrawl/internal/model/request/manager"
	pbCrawl "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/types/sortedmap"
)

type NodeControllerOptions struct {
	HeartbeatInternal int64 // 单位毫秒
}

// NodeController
type NodeController struct {
	ctx            context.Context
	nodeManager    *nodeManager.NodeManager
	crawlerManager *crawlerManager.CrawlerManager
	requestManager *reqManager.RequestManager
	nodeHandlers   *sortedmap.SortedMap
	options        NodeControllerOptions
	logger         glog.Log
}

func NewNodeController(ctx context.Context,
	nodeManager *nodeManager.NodeManager,
	crawlerManager *crawlerManager.CrawlerManager,
	requestManager *reqManager.RequestManager,
	logger glog.Log,
) (*NodeController, error) {
	c := NodeController{
		ctx:            ctx,
		nodeManager:    nodeManager,
		requestManager: requestManager,
		nodeHandlers:   sortedmap.New(),
		logger:         logger.New("NodeController"),
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
