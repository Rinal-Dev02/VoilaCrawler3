package v1

import (
	"context"

	crawlerCtrl "github.com/voiladev/VoilaCrawler/internal/controller/crawler"
	crawlerManager "github.com/voiladev/VoilaCrawler/internal/model/crawler/manager"
	pbCrawler "github.com/voiladev/VoilaCrawler/protoc-gen-go/chameleon/smelter/v1/crawler"
	"github.com/voiladev/go-framework/glog"
	"google.golang.org/grpc"
)

type GatewayServer struct {
	crawlerManager *crawlerManager.CrawlerManager
	crawlerCtrl    *crawlerCtrl.CrawlerController
	logger         glog.Log
}

func NewGatewayServer(
	crawlerManager crawlerManager.CrawlerManager,
	crawlerCtrl crawlerCtrl.CrawlerController,
	logger glog.Log,
) (*GatewayServer, error) {
	s := GatewayServer{
		logger: logger.New("GatewayServer"),
	}

	return &s, nil
}

func (s *GatewayServer) Channel(ctx context.Context, opts ...grpc.CallOption) (pbCrawler.CrawlerController_ChannelClient, error) {
	return nil, nil
}
