package history

import (
	"context"
	"errors"

	"github.com/nsqio/go-nsq"
	historyManager "github.com/voiladev/VoilaCrawl/internal/model/request/history/manager"
	"github.com/voiladev/VoilaCrawl/internal/pkg/config"
	"github.com/voiladev/go-framework/glog"
)

type RequestHistoryControllerOptions struct {
	NsqLookupdAddresses []string
}

type RequestHistoryController struct {
	ctx            context.Context
	historyManager *historyManager.HistoryManager
	producer       *nsq.Producer
	consumer       *nsq.Consumer
	logger         glog.Log
	options        RequestHistoryControllerOptions
}

func NewRequestHistoryController(ctx context.Context, historyManager *historyManager.HistoryManager, options RequestHistoryControllerOptions, logger glog.Log) (*RequestHistoryController, error) {
	if historyManager == nil {
		return nil, errors.New("invalid history manager")
	}
	if logger == nil {
		return nil, errors.New("invalid logger")
	}
	if len(options.NsqLookupdAddresses) == 0 {
		return nil, errors.New("invalid nsq lookupd address")
	}

	ctrl := RequestHistoryController{
		ctx:            ctx,
		historyManager: historyManager,
		options:        options,
		logger:         logger,
	}

	var err error
	conf := nsq.NewConfig()
	conf.MaxAttempts = 3
	conf.MaxInFlight = 6
	if ctrl.consumer, err = nsq.NewConsumer(config.CrawlErrorTopic, "crawl-api", conf); err != nil {
		return nil, err
	}
	ctrl.consumer.AddConcurrentHandlers(&RequestHistoryHandler{ctrl: &ctrl, logger: ctrl.logger.New("RequestHistoryHandler")}, conf.MaxInFlight)
	ctrl.consumer.ConnectToNSQLookupds(options.NsqLookupdAddresses)

	go func() {
		select {
		case <-ctrl.ctx.Done():
			ctrl.consumer.Stop()
			return
		}
	}()
	return &ctrl, nil
}
