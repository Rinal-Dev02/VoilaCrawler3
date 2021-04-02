package history

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/nsqio/go-nsq"
	historyManager "github.com/voiladev/VoilaCrawl/internal/model/request/history/manager"
	"github.com/voiladev/VoilaCrawl/internal/pkg/config"
	"github.com/voiladev/VoilaCrawl/pkg/types"
	"github.com/voiladev/go-framework/glog"
	pbError "github.com/voiladev/protobuf/protoc-gen-go/errors"
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

func NewRequestHistoryController(ctx context.Context, historyManager *historyManager.HistoryManager, nsqProducer *nsq.Producer, options RequestHistoryControllerOptions, logger glog.Log) (*RequestHistoryController, error) {
	if historyManager == nil {
		return nil, errors.New("invalid history manager")
	}
	if nsqProducer == nil {
		return nil, errors.New("invalid nsq producer")
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
		producer:       nsqProducer,
		options:        options,
		logger:         logger,
	}

	var err error
	if ctrl.consumer, err = nsq.NewConsumer(config.CrawlRequestHistoryTopic, "crawl-api", nsq.NewConfig()); err != nil {
		return nil, err
	}
	ctrl.consumer.AddHandler(&RequestHistoryHandler{ctrl: &ctrl, logger: ctrl.logger.New("RequestHistoryHandler")})
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

func (ctrl *RequestHistoryController) Publish(ctx context.Context, reqId string, duration, statusCode int32, errMsg string) error {
	if reqId == "" {
		return pbError.ErrInvalidArgument.New("invalid request id")
	}
	data, _ := json.Marshal(&types.RequestHistory{
		Id:         reqId,
		Timestamp:  time.Now().UnixNano() / 1000000,
		StatusCode: statusCode,
		ErrMsg:     errMsg,
		Duration:   duration,
	})
	if err := ctrl.producer.Publish(config.CrawlRequestHistoryTopic, data); err != nil {
		ctrl.logger.Error(err)
		return err
	}
	return nil
}
