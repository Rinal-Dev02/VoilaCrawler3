package store

import (
	"context"
	"errors"
	"fmt"
	"os"
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
	MaxAPIConcurrency   int32
	MaxMQConcurrency    int32
}

type StoreController struct {
	ctx            context.Context
	hostname       string
	crawlerManager *crawlerManager.CrawlerManager
	crawlerCtrl    *crawlerCtrl.CrawlerController
	producer       *nsq.Producer
	options        StoreControllerOptions
	logger         glog.Log
	storeHandlers  sync.Map
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
	if options.MaxMQConcurrency <= 0 {
		options.MaxMQConcurrency = 20
	}
	if options.MaxAPIConcurrency <= 0 {
		options.MaxAPIConcurrency = 1
	}

	ctrl := StoreController{
		ctx:            ctx,
		crawlerCtrl:    crawlerCtrl,
		crawlerManager: crawlerManager,
		logger:         logger.New("StoreController"),
		options:        options,
	}

	var err error
	ctrl.hostname, err = os.Hostname()
	if err != nil {
		return nil, err
	}
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
					handler, err = NewStoreRequestHandler(ctx, ctrl.hostname, id, ctrl,
						ctrl.crawlerCtrl, ctrl.producer, StoreRequestHandlerOptions{
							NsqLookupAddresses: ctrl.options.NsqLookupdAddresses,
							MaxAPIConcurrency:  ctrl.options.MaxAPIConcurrency,
							MaxMQConcurrency:   ctrl.options.MaxMQConcurrency,
						}, ctrl.logger)
					if err != nil {
						ctrl.logger.Error(err)
						continue
					}
					ctrl.storeHandlers.Store(id, handler)
				}

				status := handler.ConcurrencyStatus()
				ctrl.crawlerManager.UpdateStatus(ctx, ctrl.hostname, id, status, 10)
				if count, err := ctrl.crawlerManager.CountOfStore(ctx, id); err != nil {
					ctrl.logger.Error(err)
					handler.SetMQConcurrency(0)
					continue
				} else if count > 0 {
					if status.CurrentMQConcurrency == 0 {
						handler.SetMQConcurrency(DefaultHandlerConcurrency)
					}
				} else {
					handler.SetMQConcurrency(0)
				}
			}
		}
	}
}
