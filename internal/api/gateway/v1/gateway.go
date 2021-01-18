package v1

import (
	"context"

	crawlerCtrl "github.com/voiladev/VoilaCrawl/internal/controller/crawler"
	nodeCtrl "github.com/voiladev/VoilaCrawl/internal/controller/node"
	crawlerManager "github.com/voiladev/VoilaCrawl/internal/model/crawler/manager"
	pbCrawl "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl"
	"github.com/voiladev/go-framework/glog"
)

type GatewayServer struct {
	pbCrawl.UnimplementedGatewayServer

	ctx            context.Context
	nodeCtrl       *nodeCtrl.NodeController
	crawlerCtrl    *crawlerCtrl.CrawlerController
	crawlerManager *crawlerManager.CrawlerManager
	logger         glog.Log
}

func NewGatewayServer(
	ctx context.Context,
	nodeCtrl *nodeCtrl.NodeController,
	crawlerCtrl *crawlerCtrl.CrawlerController,
	crawlerManager *crawlerManager.CrawlerManager,
	logger glog.Log,
) (pbCrawl.GatewayServer, error) {
	s := GatewayServer{
		ctx:         ctx,
		nodeCtrl:    nodeCtrl,
		crawlerCtrl: crawlerCtrl,
		logger:      logger.New("GatewayServer"),
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

	// TODO

	return nil, nil
}
