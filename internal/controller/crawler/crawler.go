package crawler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/voiladev/VoilaCrawl/internal/model/crawler"
	crawlerManager "github.com/voiladev/VoilaCrawl/internal/model/crawler/manager"
	pbCrawl "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/protoutil"
	pbError "github.com/voiladev/protobuf/protoc-gen-go/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

const crawlerNodeHeartbeatInterval = time.Second * 10

var (
	heartbeatTypeUrl = protoutil.GetTypeUrl(&pbCrawl.ConnectRequest_Heartbeat{})
	requestTypeUrl   = protoutil.GetTypeUrl(&pbCrawl.Request{})
	itemTypeUrl      = protoutil.GetTypeUrl(&pbCrawl.Item{})
	errorTypeUrl     = protoutil.GetTypeUrl(&pbCrawl.Error{})
)

type CrawlerController struct {
	crawlerManager *crawlerManager.CrawlerManager
	logger         glog.Log
}

func NewCrawlerController(
	crawlerManager *crawlerManager.CrawlerManager,
	logger glog.Log,
) (*CrawlerController, error) {
	if crawlerManager == nil {
		return nil, errors.New("invalid crawler manager")
	}
	ctrl := CrawlerController{
		crawlerManager: crawlerManager,
		logger:         logger.New("CrawlerController"),
	}
	return &ctrl, nil
}

// Watch used to watch the keepalive link to decide that the spider pod is online.
// crawlet use redis to share the spider pod status. this enables start multi crawlet instance.
// when ends, the function need to clean:
// 1. the registed spider pod.
// 2. the client instance
func (ctrl *CrawlerController) Watch(ctx context.Context, srv pbCrawl.Gateway_ConnectServer, cw *crawler.Crawler) error {
	if ctrl == nil {
		return nil
	}
	logger := ctrl.logger.New("Watch")

	if err := func() error {
		nctx, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()

		// if try connect failed, return error
		if _, err := cw.Connect(nctx); err != nil {
			logger.Error(err)
			return err
		}
		return nil
	}(); err != nil {
		return err
	}

	// cache for 10 seconds, the crawler need to ping in 10 seconds
	// the crawler will check the existence of the cached info,
	// if crawler not ping in 10 seconds, the crawler auto offline.
	ctrl.crawlerManager.Cache(ctx, cw, 10)
	defer ctrl.crawlerManager.Delete(ctx, cw.GetStoreId(), cw.GetId())

	var (
		pkgChan = make(chan interface{})
	)
	defer func() {
		close(pkgChan)
	}()

	go func() {
		defer func() {
			// this recover is used in case that send to close chan
			if e := recover(); e != nil {
				logger.Error(e)
			}
		}()

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			data, err := srv.Recv()
			if err != nil {
				if err != io.EOF {
					pkgChan <- err
				}
				return
			}
			pkgChan <- data
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return pbError.ErrDeadlineExceeded
		case msg, ok := <-pkgChan:
			if !ok {
				return nil
			}
			switch v := msg.(type) {
			case error:
				return v
			case *anypb.Any:
				switch v.GetTypeUrl() {
				case heartbeatTypeUrl:
					if err := ctrl.crawlerManager.Cache(ctx, cw, 10); err != nil {
						logger.Error(err)
					}
				default:
					return fmt.Errorf("unsupported data type %s", v.GetTypeUrl())
				}
			}
		}
	}
}

func (ctrl *CrawlerController) getCrawlerClient(ctx context.Context, storeId string) (pbCrawl.CrawlerNodeClient, error) {
	if ctrl == nil {
		return nil, nil
	}
	if storeId == "" {
		return nil, pbError.ErrInvalidArgument.New("invalid store id")
	}

	var (
		crawlers, _ = ctrl.crawlerManager.GetByStore(ctx, storeId)
		client      pbCrawl.CrawlerNodeClient
	)
	for _, crawler := range crawlers {
		if cli, err := crawler.Connect(ctx); err != nil {
			ctrl.logger.Error(err)
			continue
		} else {
			client = cli
			break
		}
	}
	if client == nil {
		return nil, pbError.ErrUnavailable.New("no usable crawler found")
	}
	return client, nil
}

func (ctrl *CrawlerController) CrawlerOptions(ctx context.Context, storeId string, url string) (*pbCrawl.CrawlerOptions, error) {
	if ctrl == nil {
		return nil, nil
	}
	logger := ctrl.logger.New("CrawlerOptions")

	client, err := ctrl.getCrawlerClient(ctx, storeId)
	if err != nil {
		return nil, err
	}

	resp, err := client.CrawlerOptions(ctx, &pbCrawl.CrawlerOptionsRequest{Url: url})
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	return resp.GetData(), nil
}

func (ctrl *CrawlerController) CanonicalUrl(ctx context.Context, storeId string, url string) (string, error) {
	if ctrl == nil {
		return "", nil
	}
	logger := ctrl.logger.New("CanonicalUrl")

	client, err := ctrl.getCrawlerClient(ctx, storeId)
	if err != nil {
		return "", err
	}

	resp, err := client.CanonicalUrl(ctx, &pbCrawl.CanonicalUrlRequest{Url: url})
	if err != nil {
		logger.Error(err)
		return "", err
	}
	return resp.GetData().GetUrl(), nil
}

func (ctrl *CrawlerController) Parse(ctx context.Context, storeId string, r *pbCrawl.Request, yield func(ctx context.Context, data proto.Message) error) error {
	if ctrl == nil {
		return nil
	}
	logger := ctrl.logger.New("Parse")

	client, err := ctrl.getCrawlerClient(ctx, storeId)
	if err != nil {
		return err
	}

	return func(ctx context.Context, client pbCrawl.CrawlerNodeClient) (err error) {
		defer func() {
			if e := recover(); e != nil {
				logger.Error(err)
				err = fmt.Errorf("%v", e)
			}
		}()

		parseClient, err := client.Parse(ctx, r)
		if err != nil {
			logger.Error(err)
			return err
		}

		parentUrl, _ := url.Parse(r.GetUrl())
		for {
			select {
			case <-ctx.Done():
				return nil
			default:
			}

			data, err := parseClient.Recv()
			if err != nil {
				if err == io.EOF {
					return nil
				}
				logger.Error(err)
				return err
			}
			switch data.GetTypeUrl() {
			case requestTypeUrl:
				var req pbCrawl.Request
				if err := proto.Unmarshal(data.GetValue(), &req); err != nil {
					logger.Error(err)
					continue
				}
				u, err := url.Parse(req.GetUrl())
				if err != nil {
					logger.Error(err)
					continue
				}
				if u.Scheme == "" {
					u.Scheme = parentUrl.Scheme
				}
				if u.Host == "" {
					u.Host = parentUrl.Host
				}
				req.Url = u.String()

				if _, ok := req.CustomHeaders["Referer"]; !ok {
					req.CustomHeaders["Referer"] = r.Url
				}

				req.TracingId = r.TracingId
				req.StoreId = r.StoreId
				req.JobId = r.JobId
				req.ReqId = r.ReqId
				req.Parent = r

				if err := yield(ctx, &req); err != nil {
					return err
				}
			case itemTypeUrl:
				var item pbCrawl.Item
				if err := proto.Unmarshal(data.GetValue(), &item); err != nil {
					logger.Error(err)
					return err
				}
				item.TracingId = r.TracingId
				item.StoreId = r.StoreId
				item.JobId = r.JobId
				item.ReqId = r.ReqId

				if err = yield(ctx, &item); err != nil {
					return err
				}
			case errorTypeUrl:
				var item pbCrawl.Error
				if err := proto.Unmarshal(data.GetValue(), &item); err != nil {
					logger.Error(err)
					return err
				}
				item.TracingId = r.TracingId
				item.StoreId = r.StoreId
				item.JobId = r.JobId
				item.ReqId = r.ReqId

				if err := yield(ctx, &item); err != nil {
					return err
				}
			}
		}
	}(ctx, client)
}
