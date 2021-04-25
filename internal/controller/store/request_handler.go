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
	pbCrawl "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl"
	"github.com/voiladev/go-framework/glog"
	pbError "github.com/voiladev/protobuf/protoc-gen-go/errors"
	"google.golang.org/protobuf/proto"
)

const (
	defaultMsgTimeout = time.Minute * 10
	maxMsgTimeout     = time.Minute * 30

	DefaultHandlerConcurrency = 3
)

type parseRequest struct {
	ctx context.Context
	req *pbCrawl.Request
}

type StoreRequestHandlerOptions struct {
	NsqLookupAddresses []string
}

type StoreRequestHandler struct {
	ctx                 context.Context
	storeId             string
	storeCtrl           *StoreController
	crawlerCtrl         *crawlerCtrl.CrawlerController
	producer            *nsq.Producer
	maxConcurrencyLimit int32
	logger              glog.Log
	options             StoreRequestHandlerOptions

	consumer             *nsq.Consumer
	maxConcurrency       int32 // >= currentConcurrency
	currentConcurrency   int32 // >= currentMQConcurrency + api concurrency
	maxMQConcurrency     int32
	currentMQConcurrency int32

	continusSucceesCount int32
	continusErrorCount   int32
	mutex                sync.RWMutex

	parseRequestChan chan *parseRequest
}

func NewStoreRequestHandler(ctx context.Context, storeId string, storeCtrl *StoreController, crawlerCtrl *crawlerCtrl.CrawlerController, producer *nsq.Producer, maxConcurrency int32, options StoreRequestHandlerOptions, logger glog.Log) (*StoreRequestHandler, error) {
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
	if maxConcurrency <= 0 || maxConcurrency > 50 {
		return nil, errors.New("invalid max concurrency")
	}
	if logger == nil {
		return nil, errors.New("invalid logger")
	}
	if len(options.NsqLookupAddresses) == 0 {
		return nil, errors.New("invalid nsqlookupd addresses")
	}

	h := StoreRequestHandler{
		ctx:                 ctx,
		storeId:             storeId,
		storeCtrl:           storeCtrl,
		crawlerCtrl:         crawlerCtrl,
		producer:            producer,
		maxConcurrencyLimit: maxConcurrency,
		logger:              logger.New(fmt.Sprintf("StoreRequestHandler %s", storeId)),
		options:             options,
	}

	var (
		err   error
		topic = fmt.Sprintf("%s-%s", config.CrawlStoreRequestTopicPrefix, storeId)
		conf  = nsq.NewConfig()
	)
	conf.MsgTimeout = defaultMsgTimeout
	conf.MaxAttempts = 6 // this attempts only used when meet with crawler unavaiable error
	conf.MaxInFlight = 0
	if h.consumer, err = nsq.NewConsumer(topic, "crawlet", conf); err != nil {
		return nil, err
	}
	h.consumer.AddConcurrentHandlers(&h, int(maxConcurrency))
	if err = h.consumer.ConnectToNSQLookupds(h.options.NsqLookupAddresses); err != nil {
		return nil, err
	}
	// disable auto push msg, let store controller to start
	h.consumer.ChangeMaxInFlight(0)

	go func() {
		const speedCheckInterval = time.Minute * 5 * 60 // 5mins
		var (
			ticker                   = time.NewTicker(speedCheckInterval)
			mqConcurrencyCheckTicker = time.NewTicker(time.Second * 5)
		)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				func() {
					succeedCount, errorCount := atomic.LoadInt32(&h.continusSucceesCount), atomic.LoadInt32(&h.continusErrorCount)
					currentCon := atomic.LoadInt32(&h.maxConcurrency)
					atomic.StoreInt32(&h.continusSucceesCount, 0)
					atomic.StoreInt32(&h.continusErrorCount, 0)

					total := succeedCount + errorCount
					succeedRate := float64(succeedCount) / float64(total)

					if total <= 2*DefaultHandlerConcurrency && currentCon > DefaultHandlerConcurrency {
						h.SetConcurrency(DefaultHandlerConcurrency)
						return
					}
					switch {
					case succeedRate > 0.9:
						if currentCon < h.maxConcurrencyLimit {
							h.IncreConcurrency(1)
						}
					case succeedRate == 0:
						// down to only one thread
						for i := 0; i < 2 && currentCon > 1; i++ {
							h.IncreConcurrency(-1)
							currentCon = atomic.LoadInt32(&h.maxConcurrency)
						}
					case succeedRate < 0.4:
						// down to default thread count
						if currentCon > DefaultHandlerConcurrency {
							h.IncreConcurrency(-1)
						}
					}

					currentCon = atomic.LoadInt32(&h.maxConcurrency)
					currentRunningCon := atomic.LoadInt32(&h.currentConcurrency)
					currentMQCon := atomic.LoadInt32(&h.currentMQConcurrency)
					h.logger.Debugf("max: %d, current: %d, mq: %d", currentCon, currentRunningCon, currentMQCon)
				}()
			case <-mqConcurrencyCheckTicker.C:
				maxCon := atomic.LoadInt32(&h.maxConcurrency)
				maxMQCon := atomic.LoadInt32(&h.maxMQConcurrency)
				currentCon := atomic.LoadInt32(&h.currentConcurrency)
				currentMQCon := atomic.LoadInt32(&h.currentMQConcurrency)
				if maxMQCon < maxCon && // 并发数没有达到当前最大限制
					currentMQCon >= maxMQCon && // 当前并发数达到或者超过MQ最大限制
					currentCon < maxCon { // 当前并发数没有超过最大并发数

					h.IncreMQConcurrency(1)
				}
			}
		}
	}()
	return &h, nil
}

func (h *StoreRequestHandler) CurrentMaxConcurrency() int32 {
	if h == nil {
		return 0
	}
	return atomic.LoadInt32(&h.maxConcurrency)
}

func (h *StoreRequestHandler) SetConcurrency(count int32) {
	if h == nil || h.consumer == nil {
		return
	}
	if count > h.maxConcurrencyLimit {
		count = h.maxConcurrencyLimit
	}
	if count < 0 {
		count = 0
	}

	h.mutex.Lock()
	defer h.mutex.Unlock()

	currentCon := atomic.LoadInt32(&h.currentConcurrency)
	currentMQCon := atomic.LoadInt32(&h.currentMQConcurrency)
	maxMQCon := count - (currentCon - currentMQCon)
	if maxMQCon < 0 {
		maxMQCon = 0
	}

	atomic.StoreInt32(&h.maxConcurrency, count)
	atomic.StoreInt32(&h.maxMQConcurrency, maxMQCon)
	h.consumer.ChangeMaxInFlight(int(maxMQCon))
}

func (h *StoreRequestHandler) IncreConcurrency(step int32) {
	if h == nil || h.consumer == nil {
		return
	}
	if step == 0 {
		return
	}

	h.mutex.Lock()
	defer h.mutex.Unlock()

	maxCon := atomic.LoadInt32(&h.maxConcurrency)
	if step+maxCon < 0 {
		step = -1 * maxCon
	}
	if step+maxCon > h.maxConcurrencyLimit {
		step = h.maxConcurrencyLimit - maxCon
	}
	maxCon = maxCon + step

	// check max concurrency for mq
	currentCon := atomic.LoadInt32(&h.currentConcurrency)
	currentMQCon := atomic.LoadInt32(&h.currentMQConcurrency)
	maxMQCon := maxCon - (currentCon - currentMQCon)
	if maxMQCon < 0 {
		maxMQCon = 0
	}
	atomic.StoreInt32(&h.maxConcurrency, maxCon)
	atomic.StoreInt32(&h.maxMQConcurrency, maxMQCon)
	h.consumer.ChangeMaxInFlight(int(maxMQCon))
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
	if step+maxCon < 0 {
		// set to zero
		step = -1 * maxCon
	}
	if step+maxCon > h.maxConcurrencyLimit {
		step = h.maxConcurrencyLimit - maxCon
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
		if err := h.producer.Publish(topic, msgData); err != nil {
			h.logger.Error(err)
			return err
		}
		return nil
	})
	if err == nil {
		if retCount == 0 {
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

	decreased := false
	for {
		select {
		case <-ctx.Done():
			return 0, 0, pbError.ErrDeadlineExceeded
		default:
		}

		maxCon := atomic.LoadInt32(&h.maxConcurrency)
		if maxCon == 0 {
			return 0, 0, pbError.ErrUnavailable.New("no crawler available")
		}
		maxMQCon := atomic.LoadInt32(&h.maxMQConcurrency)
		currentCon := atomic.LoadInt32(&h.currentConcurrency)
		currentMQCon := atomic.LoadInt32(&h.currentMQConcurrency)
		gap1, gap2 := maxCon-maxMQCon, currentCon-currentMQCon
		if gap2 < gap1 {
			return h.parse(ctx, req, callback)
		}

		if maxMQCon > 0 && !decreased {
			h.IncreMQConcurrency(-1)
			// here not recover the decreased mq concurrency
			decreased = true
		}
		// else wait idle chance
		time.Sleep(time.Millisecond * 200)
	}
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
		e := pbError.NewFromError(err)
		if e.Code() == pbError.ErrUnavailable.Code() {
			// stop the handler and requeue the message
			// the store controller will check the crawler state and restart handler
			h.SetConcurrency(0)
			msg.RequeueWithoutBackoff(time.Second * 60)
			return nil
		}

		// record the error message
		errItemData, _ := proto.Marshal(&pbCrawl.Error{
			StoreId:   req.GetStoreId(),
			TracingId: req.GetTracingId(),
			JobId:     req.GetJobId(),
			ReqId:     req.GetReqId(),
			Timestamp: time.Now().Unix(),
			ErrMsg:    err.Error(),
		})

		isSucceed = false
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
