package store

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/nsqio/go-nsq"
	crawlerCtrl "github.com/voiladev/VoilaCrawl/internal/controller/crawler"
	crawlerManager "github.com/voiladev/VoilaCrawl/internal/model/crawler/manager"
	pbCrawl "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl"
	"github.com/voiladev/go-framework/glog"
	pbError "github.com/voiladev/protobuf/protoc-gen-go/errors"
	"google.golang.org/protobuf/proto"
)

// This controller used to manage the handler of the stores.
// update the concurrency of the processor.

type StoreControllerOptions struct {
	NsqLookupdAddresses []string
	NsqdAddress         string
	MaxCurrency         int32
}

type StoreController struct {
	ctx            context.Context
	crawlerManager *crawlerManager.CrawlerManager
	crawlerCtrl    *crawlerCtrl.CrawlerController
	producer       *nsq.Producer
	options        StoreControllerOptions
	logger         glog.Log

	storeHandlers sync.Map
}

func NewStoreController(
	ctx context.Context,
	crawlerCtrl *crawlerCtrl.CrawlerController,
	crawlerManager *crawlerManager.CrawlerManager,
	logger glog.Log,
	options StoreControllerOptions,
) (*StoreController, error) {
	if crawlerCtrl == nil {
		return nil, errors.New("invalid crawler controller")
	}
	if crawlerManager == nil {
		return nil, errors.New("invalid crawler manager")
	}
	if logger == nil {
		return nil, errors.New("invalid logger")
	}
	if len(options.NsqLookupdAddresses) == 0 {
		return nil, errors.New("invalid lookupd address")
	}
	if options.NsqdAddress == "" {
		return nil, errors.New("invalid nsqd address")
	}
	if options.MaxCurrency <= 0 {
		options.MaxCurrency = 20
	}

	ctrl := StoreController{
		ctx:            ctx,
		crawlerCtrl:    crawlerCtrl,
		crawlerManager: crawlerManager,
		logger:         logger.New("StoreController"),
		options:        options,
	}

	var err error
	if ctrl.producer, err = nsq.NewProducer(options.NsqdAddress, nsq.NewConfig()); err != nil {
		return nil, err
	}
	return &ctrl, nil
}

// Parse note that this func may exists change race
func (ctrl *StoreController) Parse(ctx context.Context, storeId string, req *pbCrawl.Request, callback func(context.Context, proto.Message) error) (int, int, error) {
	if ctrl == nil {
		return 0, 0, nil
	}

	if storeId == "" {
		return 0, 0, pbError.ErrInvalidArgument.New("invalid store id")
	}
	if val, ok := ctrl.storeHandlers.Load(storeId); ok {
		handler, _ := val.(*StoreRequestHandler)
		return handler.Parse(ctx, req, callback)
	}
	return 0, 0, pbError.ErrUnavailable.New(fmt.Sprintf("no crawler available for store %s", req.GetStoreId()))
}

func (ctrl *StoreController) Run(ctx context.Context) error {
	if ctrl == nil {
		return nil
	}

	const checkInterval = time.Second * 5
	ticker := time.NewTicker(checkInterval)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			storeIds, err := ctrl.crawlerManager.GetStores(ctx)
			if err != nil {
				ctrl.logger.Error(err)
				continue
			}
			for _, id := range storeIds {
				if id == "" {
					continue
				}

				if err := ctrl.crawlerManager.Clean(ctx, id); err != nil {
					ctrl.logger.Error(err)
				}

				var handler *StoreRequestHandler
				if val, ok := ctrl.storeHandlers.Load(id); ok {
					handler, _ = val.(*StoreRequestHandler)
				} else {
					if handler, err = NewStoreRequestHandler(ctx, id, ctrl, ctrl.crawlerCtrl, ctrl.producer,
						ctrl.options.MaxCurrency, StoreRequestHandlerOptions{
							NsqLookupAddresses: ctrl.options.NsqLookupdAddresses,
						}, ctrl.logger); err != nil {
						ctrl.logger.Error(err)
						continue
					}
					ctrl.storeHandlers.Store(id, handler)
				}
				if count, err := ctrl.crawlerManager.CountOfStore(ctx, id); err != nil {
					ctrl.logger.Error(err)
					handler.SetConcurrency(0)
					continue
				} else if count > 0 {
					if handler.CurrentMaxConcurrency() == 0 {
						handler.SetConcurrency(DefaultHandlerConcurrency)
					}
				} else {
					handler.SetConcurrency(0)
				}
			}
		}
	}
}
