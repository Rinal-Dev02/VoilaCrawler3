package request

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/nsqio/go-nsq"
	historyCtrl "github.com/voiladev/VoilaCrawl/internal/controller/request/history"
	"github.com/voiladev/VoilaCrawl/internal/model/request"
	reqManager "github.com/voiladev/VoilaCrawl/internal/model/request/manager"
	"github.com/voiladev/VoilaCrawl/internal/pkg/config"
	pbCrawl "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/redis"
	pbError "github.com/voiladev/protobuf/protoc-gen-go/errors"
	"xorm.io/xorm"
)

type RequestControllerOptions struct {
	NsqdAddress         string
	NsqLookupdAddresses []string
}

type RequestController struct {
	ctx            context.Context
	engine         *xorm.Engine
	historyCtrl    *historyCtrl.RequestHistoryController
	requestManager *reqManager.RequestManager
	redisClient    *redis.RedisClient

	nsqConsumer       *nsq.Consumer
	nsqStatusConsumer *nsq.Consumer
	producer          *nsq.Producer

	options RequestControllerOptions
	logger  glog.Log
}

func NewRequestController(
	ctx context.Context,
	engine *xorm.Engine,
	historyCtrl *historyCtrl.RequestHistoryController,
	requestManager *reqManager.RequestManager,
	redisClient *redis.RedisClient,
	options RequestControllerOptions,
	logger glog.Log) (*RequestController, error) {

	if engine == nil {
		return nil, errors.New("invalid xorm engine")
	}
	if historyCtrl == nil {
		return nil, errors.New("invalid history controller")
	}
	if requestManager == nil {
		return nil, errors.New("invalid requestManager")
	}
	if redisClient == nil {
		return nil, errors.New("invalid redis client")
	}
	if options.NsqdAddress == "" {
		return nil, errors.New("invalid nsqd address")
	}
	if len(options.NsqLookupdAddresses) == 0 {
		return nil, errors.New("invalid nsqlookupd address")
	}

	ctrl := RequestController{
		ctx:            ctx,
		engine:         engine,
		historyCtrl:    historyCtrl,
		requestManager: requestManager,
		redisClient:    redisClient,
		options:        options,
		logger:         logger.New("RequestController"),
	}

	var err error
	conf := nsq.NewConfig()
	conf.MsgTimeout = time.Minute * 5
	conf.MaxAttempts = 3

	if ctrl.producer, err = nsq.NewProducer(options.NsqdAddress, conf); err != nil {
		return nil, err
	}

	if ctrl.nsqConsumer, err = nsq.NewConsumer(config.CrawlRequestTopic, "crawl-api", conf); err != nil {
		return nil, err
	}
	ctrl.nsqConsumer.AddHandler(&RequestHander{ctrl: &ctrl, logger: ctrl.logger.New("RequestHander")})
	ctrl.nsqConsumer.ConnectToNSQLookupds(options.NsqLookupdAddresses)

	if ctrl.nsqStatusConsumer, err = nsq.NewConsumer(config.CrawlRequestStatusTopic, "crawl-api", conf); err != nil {
		return nil, err
	}
	ctrl.nsqStatusConsumer.AddHandler(&RequestStatusHander{ctrl: &ctrl, logger: ctrl.logger.New("RequestStatusHander")})
	ctrl.nsqStatusConsumer.ConnectToNSQLookupds(options.NsqLookupdAddresses)

	go func() {
		select {
		case <-ctrl.ctx.Done():
			ctrl.nsqConsumer.Stop()
			return
		}
	}()
	return &ctrl, nil
}

func (ctrl *RequestController) PublishRequest(ctx context.Context, session *xorm.Session, r *request.Request, force bool) error {
	if ctrl == nil {
		return nil
	}
	logger := ctrl.logger.New("PublishRequest")

	if r == nil {
		return pbError.ErrInvalidArgument.New("invalid request")
	}

	var req pbCrawl.Request
	if err := r.Unmarshal(&req); err != nil {
		logger.Error(err)
		return pbError.ErrInternal.New(err)
	}

	if session == nil {
		session = ctrl.engine.NewSession()
		defer session.Close()
	}

	if succeed, err := ctrl.requestManager.UpdateStatus(ctx, session, r.GetId(), 1, false); err != nil {
		logger.Errorf("update status of request %s failed, error=%s", r.GetId(), err)
		session.Rollback()
		return err
	} else if succeed {
		key := fmt.Sprintf("%s-%s", config.CrawlRequestTopic, r.GetStoreId())
		data, _ := proto.Marshal(&req)
		if err := ctrl.producer.Publish(key, data); err != nil {
			logger.Error(err)
			return pbError.ErrInternal.New(err)
		}
		return nil
	}
	return pbError.ErrFailedPrecondition.New("max retry count reached")
}

const (
	defaultCheckTimeoutRequestInterval = time.Second * 30
)

func (ctrl *RequestController) Run(ctx context.Context) error {
	if ctrl == nil {
		return nil
	}

	timer := time.NewTimer(defaultCheckTimeoutRequestInterval)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
			reqs, err := ctrl.requestManager.List(ctx, nil, reqManager.ListRequest{
				Page: 1, Count: 1000, Retryable: true,
			})
			if err != nil {
				ctrl.logger.Errorf("list request failed, error=%s", err)
				timer.Reset(defaultCheckTimeoutRequestInterval)
				continue
			}

			for _, req := range reqs {
				if err := ctrl.PublishRequest(ctx, nil, req, false); err != nil {
					ctrl.logger.Errorf("publish request failed, error=%s", err)
				}
			}
			timer.Reset(defaultCheckTimeoutRequestInterval)
		}
	}
}
