package store

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nsqio/go-nsq"
	crawlerCtrl "github.com/voiladev/VoilaCrawl/internal/controller/crawler"
	"github.com/voiladev/VoilaCrawl/internal/pkg/config"
	"github.com/voiladev/VoilaCrawl/pkg/types"
	pbCrawl "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl"
	"github.com/voiladev/go-framework/glog"
	pbError "github.com/voiladev/protobuf/protoc-gen-go/errors"
	"golang.org/x/time/rate"
	"google.golang.org/protobuf/proto"
)

const (
	defaultMsgTimeout = time.Minute * 10
	maxMsgTimeout     = time.Minute * 30

	DefaultHandlerConcurrency = 2
)

type parseRequest struct {
	ctx context.Context
	req *pbCrawl.Request
}

type StoreRequestHandlerOptions struct {
	NsqLookupAddresses []string
	MaxAPIConcurrency  int32
	MaxMQConcurrency   int32
}

type StoreRequestHandler struct {
	ctx         context.Context
	hostname    string
	storeId     string
	storeCtrl   *StoreController
	crawlerCtrl *crawlerCtrl.CrawlerController
	producer    *nsq.Producer

	currentConcurrency int32
	// API
	maxAPILimiter int32
	apiLimiter    *rate.Limiter

	// MQ
	consumer             *nsq.Consumer
	maxMQLimiter         int32
	maxMQConcurrency     int32
	currentMQConcurrency int32

	continusSucceesCount int32
	continusErrorCount   int32
	mutex                sync.RWMutex

	parseRequestChan chan *parseRequest

	logger  glog.Log
	options StoreRequestHandlerOptions
}

func NewStoreRequestHandler(ctx context.Context, hostname, storeId string,
	storeCtrl *StoreController, crawlerCtrl *crawlerCtrl.CrawlerController, producer *nsq.Producer,
	options StoreRequestHandlerOptions, logger glog.Log) (*StoreRequestHandler, error) {

	if hostname == "" {
		return nil, errors.New("invalid hostname")
	}
	if storeId == "" {
		return nil, errors.New("invalid storeId")
	}
	if storeCtrl == nil {
		return nil, errors.New("invalid store ctrl")
	}
	if crawlerCtrl == nil {
		return nil, errors.New("invalid crawler ctrl")
	}
	if producer == nil {
		return nil, errors.New("invalid nsq producer")
	}
	if options.MaxAPIConcurrency < 0 || options.MaxAPIConcurrency > 10 {
		return nil, errors.New("invalid api concurrency")
	}
	if options.MaxMQConcurrency <= 0 || options.MaxMQConcurrency > 50 {
		return nil, errors.New("invalid max concurrency")
	}
	if logger == nil {
		return nil, errors.New("invalid logger")
	}
	if len(options.NsqLookupAddresses) == 0 {
		return nil, errors.New("invalid nsqlookupd addresses")
	}

	rateLimiter := rate.NewLimiter(rate.Every(time.Second/time.Duration(options.MaxAPIConcurrency)), 1)
	h := StoreRequestHandler{
		ctx:         ctx,
		hostname:    hostname,
		storeId:     storeId,
		storeCtrl:   storeCtrl,
		crawlerCtrl: crawlerCtrl,
		producer:    producer,

		maxAPILimiter: options.MaxAPIConcurrency,
		apiLimiter:    rateLimiter,

		maxMQLimiter: options.MaxMQConcurrency,
		logger:       logger.New(fmt.Sprintf("StoreRequestHandler %s", storeId)),
		options:      options,
	}

	var (
		err   error
		topic = fmt.Sprintf("%s-%s", config.CrawlStoreRequestTopicPrefix, storeId)
		conf  = nsq.NewConfig()
	)
	conf.MsgTimeout = defaultMsgTimeout
	conf.MaxAttempts = 30 // this attempts only used when meet with crawler unavaiable error
	conf.MaxInFlight = 0
	if h.consumer, err = nsq.NewConsumer(topic, "crawlet", conf); err != nil {
		return nil, err
	}
	h.consumer.AddConcurrentHandlers(&h, int(options.MaxMQConcurrency))
	h.consumer.SetLogger(nil, nsq.LogLevelError)
	if err = h.consumer.ConnectToNSQLookupds(h.options.NsqLookupAddresses); err != nil {
		return nil, err
	}

	go func() {
		const speedCheckInterval = time.Minute * 5 * 60 // 5mins
		var (
			ticker = time.NewTicker(speedCheckInterval)
		)
		defer func() {
			ticker.Stop()
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				func() {
					// This logic is used to change the concurrency of mq
					succeedCount, errorCount := atomic.LoadInt32(&h.continusSucceesCount), atomic.LoadInt32(&h.continusErrorCount)
					currentCon := atomic.LoadInt32(&h.currentMQConcurrency)
					atomic.StoreInt32(&h.continusSucceesCount, 0)
					atomic.StoreInt32(&h.continusErrorCount, 0)

					succeedRate := 0.0
					total := succeedCount + errorCount
					if total > 0 {
						succeedRate = float64(succeedCount) / float64(total)
					}
					if total <= 2*DefaultHandlerConcurrency && currentCon > DefaultHandlerConcurrency {
						h.SetMQConcurrency(DefaultHandlerConcurrency)
						return
					}

					switch {
					case succeedRate > 0.9:
						h.IncreMQConcurrency(1)
					case succeedRate < 0.1:
						h.SetMQConcurrency(DefaultHandlerConcurrency)
					case succeedRate < 0.4:
						// down to default thread count
						if currentCon > DefaultHandlerConcurrency {
							h.IncreMQConcurrency(-1)
						}
					}
				}()
			}
		}
	}()
	return &h, nil
}

func (h *StoreRequestHandler) ConcurrencyStatus() *types.Crawler_Status {
	if h == nil {
		return nil
	}
	status := types.Crawler_Status{}
	status.MaxAPIConcurrency = atomic.LoadInt32(&h.maxAPILimiter)
	status.MaxMQConcurrency = atomic.LoadInt32(&h.maxMQConcurrency)
	status.CurrentConcurrency = atomic.LoadInt32(&h.currentConcurrency)
	status.CurrentMQConcurrency = atomic.LoadInt32(&h.currentMQConcurrency)

	return &status
}

func (h *StoreRequestHandler) SetMQConcurrency(count int32) {
	if h == nil || h.consumer == nil {
		return
	}
	if count < 0 {
		return
	}
	if count > h.maxMQLimiter {
		count = h.maxMQLimiter
	}
	h.mutex.Lock()
	defer h.mutex.Unlock()

	atomic.StoreInt32(&h.maxMQConcurrency, count)
	h.consumer.ChangeMaxInFlight(int(count))
}

func (h *StoreRequestHandler) IncreMQConcurrency(step int32) {
	if h == nil || h.consumer == nil {
		return
	}
	if step == 0 {
		return
	}

	h.mutex.Lock()
	defer h.mutex.Unlock()

	maxCon := atomic.LoadInt32(&h.maxMQConcurrency)
	if maxCon+step > h.maxMQLimiter {
		step = h.maxMQLimiter - maxCon
	}
	if maxCon+step < 0 {
		step = -maxCon
	}

	atomic.StoreInt32(&h.maxMQConcurrency, maxCon+step)
	h.consumer.ChangeMaxInFlight(int(maxCon + step))
}

type parseOptions struct {
	EnableBlockForItems bool
}

func (h *StoreRequestHandler) parse(ctx context.Context, req *pbCrawl.Request, callback func(context.Context, proto.Message) error) (int, int, error) {
	if h == nil || req == nil {
		return 0, 0, nil
	}

	atomic.AddInt32(&h.currentConcurrency, 1)
	defer func() {
		atomic.AddInt32(&h.currentConcurrency, -1)
	}()

	var (
		retCount    int
		subreqCount int
		itemCount   int
	)
	err := h.crawlerCtrl.Parse(ctx, h.storeId, req, func(ctx context.Context, data proto.Message) error {
		if data == nil {
			return nil
		}

		var (
			topic   string
			msgData []byte
		)
		switch v := data.(type) {
		case *pbCrawl.Request:
			if callback != nil {
				if err := callback(ctx, v); err != nil {
					return err
				}
			} else {
				topic = config.CrawlRequestTopic
				msgData, _ = proto.Marshal(v)
			}
			retCount += 1
			subreqCount += 1
		case *pbCrawl.Item:
			if callback != nil {
				if err := callback(ctx, v); err != nil {
					return err
				}
			} else {
				topic = config.CrawlItemTopic
				msgData, _ = proto.Marshal(v)
			}
			retCount += 1
			itemCount += 1
		case *pbCrawl.Error:
			topic = config.CrawlErrorTopic
			msgData, _ = proto.Marshal(v)
			retCount += 1
		}

		if topic != "" {
			if err := h.producer.Publish(topic, msgData); err != nil {
				h.logger.Error(err)
				return err
			}
		}
		return nil
	})
	if err == nil {
		if retCount == 0 && callback == nil {
			err = fmt.Errorf("no item or subrequest got of url %s", req.GetUrl())
		}
	}
	if err != nil {
		return itemCount, subreqCount, err
	}
	return itemCount, subreqCount, nil
}

// Parse note that this func may exists change race
func (h *StoreRequestHandler) Parse(ctx context.Context, req *pbCrawl.Request, callback func(context.Context, proto.Message) error) (int, int, error) {
	if h == nil {
		return 0, 0, nil
	}

	if err := h.apiLimiter.Wait(ctx); err != nil {
		if err == context.DeadlineExceeded || err == context.Canceled {
			return 0, 0, nil
		}
		return 0, 0, err
	}
	return h.parse(ctx, req, callback)
}

func (h *StoreRequestHandler) HandleMessage(msg *nsq.Message) error {
	if h == nil {
		return nil
	}
	msg.DisableAutoResponse()

	var req pbCrawl.Request
	if err := proto.Unmarshal(msg.Body, &req); err != nil {
		h.logger.Error(err)
		msg.Finish()
		return err
	}

	atomic.AddInt32(&h.currentMQConcurrency, 1)
	defer func() {
		atomic.AddInt32(&h.currentMQConcurrency, -1)
	}()

	ctx, cancel := context.WithTimeout(h.ctx, time.Duration(req.GetOptions().MaxTtlPerRequest)*time.Second)
	defer cancel()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			// here not check the max timeout
			time.Sleep(defaultMsgTimeout - time.Second)
			msg.Touch()
		}
	}()

	var (
		isSucceed              = true
		err                    error
		itemCount, subReqCount int
	)
	if itemCount, subReqCount, err = h.parse(ctx, &req, nil); err != nil {
		isSucceed = false
		e := pbError.NewFromError(err)
		if e.Code() == pbError.ErrUnavailable.Code() {
			// stop the handler and requeue the message
			// the store controller will check the crawler state and restart handler
			h.SetMQConcurrency(0)
			msg.RequeueWithoutBackoff(time.Second * 60)

			return nil
		}

		// disable retry for aborted and unimplemented error
		switch e.Code() {
		case int(pbError.Code_Aborted):
			isSucceed = true
		case int(pbError.Code_Unimplemented):
			isSucceed = true
		}

		// record the error message
		errItemData, _ := proto.Marshal(&pbCrawl.Error{
			StoreId:   req.GetStoreId(),
			TracingId: req.GetTracingId(),
			JobId:     req.GetJobId(),
			ReqId:     req.GetReqId(),
			Timestamp: time.Now().Unix(),
			Code:      int32(e.Code()),
			ErrMsg:    err.Error(),
		})
		if err := h.producer.Publish(config.CrawlErrorTopic, errItemData); err != nil {
			h.logger.Error(err)
		}
	}

	reqStatusData, _ := proto.Marshal(&pbCrawl.RequestStatus{
		StoreId:     req.GetStoreId(),
		TracingId:   req.GetTracingId(),
		JobId:       req.GetJobId(),
		ReqId:       req.GetReqId(),
		Timestamp:   time.Now().Unix(),
		IsSucceed:   isSucceed,
		SubReqCount: int32(subReqCount),
		ItemCount:   int32(itemCount),
	})
	if err := h.producer.Publish(config.CrawlRequestStatusTopic, reqStatusData); err != nil {
		h.logger.Error(err)
	}
	msg.Finish()
	return err
}

func (h *StoreRequestHandler) LogFailedMessage(msg *nsq.Message) {
	if h == nil {
		return
	}

	// pass
}
