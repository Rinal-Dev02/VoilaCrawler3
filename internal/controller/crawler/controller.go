package crawler

import (
	"context"

	crawlerManager "github.com/voiladev/VoilaCrawler/internal/model/crawler/manager"
	nodeManager "github.com/voiladev/VoilaCrawler/internal/model/node/manager"
	"github.com/voiladev/go-framework/glog"
)

type CrawlerControllerOptions struct {
}

// CrawlerController
type CrawlerController struct {
	ctx            context.Context
	nodeManager    *nodeManager.NodeManager
	crawlerManager *crawlerManager.CrawlerManager

	options CrawlerControllerOptions
	logger  glog.Log
}

func NewCrawlerController(ctx context.Context,
	nodeManager *nodeManager.NodeManager,
	crawlerManager *crawlerManager.CrawlerManager,
	logger glog.Log,
) (*CrawlerController, error) {
	c := CrawlerController{
		logger: logger.New("CrawlerController"),
	}

	return &c, nil
}
