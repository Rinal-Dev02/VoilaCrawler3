package request

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/nsqio/go-nsq"
	crawlerCtrl "github.com/voiladev/VoilaCrawl/internal/controller/crawler"
	historyCtrl "github.com/voiladev/VoilaCrawl/internal/controller/request/history"
	"github.com/voiladev/VoilaCrawl/internal/model/request"
	reqManager "github.com/voiladev/VoilaCrawl/internal/model/request/manager"
	"github.com/voiladev/VoilaCrawl/internal/pkg/config"
	"github.com/voiladev/VoilaCrawl/pkg/pigate"
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
	crawlerCtrl    *crawlerCtrl.CrawlerController
	historyCtrl    *historyCtrl.RequestHistoryController
	requestManager *reqManager.RequestManager
	redisClient    *redis.RedisClient
	pigate         *pigate.PigateClient

	nsqConsumer *nsq.Consumer
	producer    *nsq.Producer

	options RequestControllerOptions
	logger  glog.Log
}

func NewRequestController(
	ctx context.Context,
	engine *xorm.Engine,
	historyCtrl *historyCtrl.RequestHistoryController,
	crawlerCtrl *crawlerCtrl.CrawlerController,
	requestManager *reqManager.RequestManager,
	redisClient *redis.RedisClient,
	pigate *pigate.PigateClient,
	options RequestControllerOptions,
	logger glog.Log) (*RequestController, error) {

	if engine == nil {
		return nil, errors.New("invalid xorm engine")
	}
	if historyCtrl == nil {
		return nil, errors.New("invalid history controller")
	}
	if crawlerCtrl == nil {
		return nil, errors.New("invalid crawler controller")
	}
	if requestManager == nil {
		return nil, errors.New("invalid requestManager")
	}
	if redisClient == nil {
		return nil, errors.New("invalid redis client")
	}
	if pigate == nil {
		return nil, errors.New("invalid pigate client")
	}
	if len(options.NsqLookupdAddresses) == 0 {
		return nil, errors.New("invalid nsq lookupd address")
	}

	ctrl := RequestController{
		ctx:            ctx,
		engine:         engine,
		historyCtrl:    historyCtrl,
		crawlerCtrl:    crawlerCtrl,
		requestManager: requestManager,
		redisClient:    redisClient,
		pigate:         pigate,
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

	if ctrl.nsqConsumer, err = nsq.NewConsumer(config.CrawlRequestTopic, "crawl-api", nsq.NewConfig()); err != nil {
		return nil, err
	}
	ctrl.nsqConsumer.AddHandler(&RequestHander{ctrl: &ctrl, logger: ctrl.logger.New("RequestHander")})
	ctrl.nsqConsumer.ConnectToNSQLookupds(options.NsqLookupdAddresses)

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
		ctrl.logger.Error(err)
		return pbError.ErrInternal.New(err)
	}

	if !force {
		isMeb, err := redis.Bool(ctrl.redisClient.Do("SISMEMBER", config.CrawlRequestQueueSet, r.GetId()))
		if err != nil {
			logger.Errorf("check if request %s is queued failed, error=%s", r.GetId(), err)
			return err
		}
		if isMeb {
			return nil
		}
	}

	if session == nil {
		session = ctrl.engine.NewSession()
		defer session.Close()
	}

	if succeed, err := ctrl.requestManager.UpdateStatus(ctx, session, r.GetId(), 1, false); err != nil {
		ctrl.logger.Errorf("update status of request %s failed, error=%s", r.GetId(), err)
		session.Rollback()
		return err
	} else if succeed {
		key := fmt.Sprintf("%s-%s", config.CrawlRequestTopic, r.GetStoreId())
		data, _ := proto.Marshal(&req)
		if err := ctrl.producer.Publish(key, data); err != nil {
			ctrl.logger.Error(err)
			return pbError.ErrInternal.New(err)
		}
		return nil
	}
	return pbError.ErrFailedPrecondition.New("max retry count reached")
}
