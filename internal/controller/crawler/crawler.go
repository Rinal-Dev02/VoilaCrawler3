package crawler

import (
	"context"
	"fmt"

	"github.com/voiladev/VoilaCrawl/internal/model/request"
	pbCrawl "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl"
	pbProxy "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	"github.com/voiladev/go-framework/glog"
)

type CrawlerController struct {
	ctx           context.Context
	crawlerClient pbCrawl.CrawlerManagerClient
	logger        glog.Log
}

func NewCrawlerController(
	ctx context.Context,
	crawlerClient pbCrawl.CrawlerManagerClient,
	logger glog.Log,
) (*CrawlerController, error) {
	ctrl := CrawlerController{
		ctx:           ctx,
		crawlerClient: crawlerClient,
		logger:        logger.New("CrawlerController"),
	}
	return &ctrl, nil
}

// GetCrawlerByUrl
func (ctrl *CrawlerController) GetCrawlerByUrl(ctx context.Context, rawUrl string) ([]*pbCrawl.Crawler, error) {
	if ctrl == nil {
		return nil, nil
	}
	logger := ctrl.logger.New("GetCrawlerByUrl")

	resp, err := ctrl.crawlerClient.GetCrawlers(ctx, &pbCrawl.GetCrawlersRequest{Url: rawUrl})
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	return resp.GetData(), nil
}

func (ctrl *CrawlerController) Parse(ctx context.Context, r *request.Request, rawResp *pbProxy.Response) error {
	if ctrl == nil {
		return nil
	}
	logger := ctrl.logger.New("Parse")

	var crawlReq pbCrawl.Command_Request
	if err := r.Unmarshal(&crawlReq); err != nil {
		logger.Errorf("unmarshal request to crawl.Command_Request failed, error=%s", err)
		return err
	}

	if res, err := ctrl.crawlerClient.Parse(ctx, &pbCrawl.ParseRequest{
		Request:             &crawlReq,
		Response:            rawResp,
		EnableBlockForItems: false,
	}); err != nil {
		logger.Errorf("parse %s failed, error=%s", r.GetUrl(), err)
		return err
	} else if res.ItemCount+res.SubReqCount == 0 {
		err = fmt.Errorf("parse %s failed, error=got no item or subrequest", r.GetUrl())
		logger.Error(err)
		return err
	}
	return nil
}
