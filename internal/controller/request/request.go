package request

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/nsqio/go-nsq"
	crawlerCtrl "github.com/voiladev/VoilaCrawl/internal/controller/crawler"
	historyCtrl "github.com/voiladev/VoilaCrawl/internal/controller/request/history"
	threadCtrl "github.com/voiladev/VoilaCrawl/internal/controller/thread"
	"github.com/voiladev/VoilaCrawl/internal/model/request"
	reqManager "github.com/voiladev/VoilaCrawl/internal/model/request/manager"
	"github.com/voiladev/VoilaCrawl/internal/pkg/config"
	"github.com/voiladev/VoilaCrawl/pkg/pigate"
	pbProxy "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl/proxy"
	pbSession "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl/session"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/redis"
	pbError "github.com/voiladev/protobuf/protoc-gen-go/errors"
	"xorm.io/xorm"
)

type RequestControllerOptions struct {
	NsqLookupdAddresses []string
}

type RequestController struct {
	ctx            context.Context
	engine         *xorm.Engine
	threadCtrl     *threadCtrl.ThreadController
	crawlerCtrl    *crawlerCtrl.CrawlerController
	historyCtrl    *historyCtrl.RequestHistoryController
	requestManager *reqManager.RequestManager
	redisClient    *redis.RedisClient
	pigate         *pigate.PigateClient
	sessionManager pbSession.SessionManagerClient
	nsqConsumer    *nsq.Consumer
	options        RequestControllerOptions
	logger         glog.Log
}

func NewRequestController(
	ctx context.Context,
	engine *xorm.Engine,
	threadCtrl *threadCtrl.ThreadController,
	historyCtrl *historyCtrl.RequestHistoryController,
	crawlerCtrl *crawlerCtrl.CrawlerController,
	requestManager *reqManager.RequestManager,
	redisClient *redis.RedisClient,
	pigate *pigate.PigateClient,
	sessionManager pbSession.SessionManagerClient,
	options RequestControllerOptions,
	logger glog.Log) (*RequestController, error) {

	if engine == nil {
		return nil, errors.New("invalid xorm engine")
	}
	if threadCtrl == nil {
		return nil, errors.New("invalid thread controller")
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
	if sessionManager == nil {
		return nil, errors.New("invalid session manager")
	}
	if len(options.NsqLookupdAddresses) == 0 {
		return nil, errors.New("invalid nsq lookupd address")
	}

	ctrl := RequestController{
		ctx:            ctx,
		engine:         engine,
		threadCtrl:     threadCtrl,
		historyCtrl:    historyCtrl,
		crawlerCtrl:    crawlerCtrl,
		requestManager: requestManager,
		redisClient:    redisClient,
		pigate:         pigate,
		sessionManager: sessionManager,
		options:        options,
		logger:         logger.New("RequestController"),
	}

	var err error
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

	if !force {
		isMem, err := redis.Bool(ctrl.redisClient.Do("SISMEMBER", config.CrawlRequestQueueSet, r.GetId()))
		if err != nil {
			logger.Errorf("check if request %s is queued failed, error=%s", r.GetId(), err)
			return err
		}
		if isMem {
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
		key := fmt.Sprintf(config.CrawlRequestStoreQueue, r.GetStoreId())
		if _, err := ctrl.redisClient.Do("EVAL", requestPushScript, 3,
			config.CrawlStoreList, key, config.CrawlRequestQueueSet, r.GetId()); err != nil {
			ctrl.logger.Errorf("requeue request %s failed, error=%s", r.GetId(), err)
			return err
		}
		return nil
	}
	return pbError.ErrFailedPrecondition.New("max retry count reached")
}

const (
	defaultCheckTimeoutRequestInterval = time.Second * 5

	// KEYS[1]-Stores, KEYS[2]-StoreQueue, KEYS[3]-Set
	// ARGV[1]-reqId
	requestPushScript = `local ret = redis.call("LPUSH", KEYS[2], ARGV[1])
redis.call("SADD", KEYS[3], ARGV[1])
local count = redis.call("LLEN", KEYS[2])
redis.call("ZADD", KEYS[1], count, KEYS[2])
return ret`

	// KEYS[1]-Stores, KEYS[2]-StoreQueue
	requestPopScript = `local ret = redis.call("RPOP", KEYS[2])
local count = redis.call("LLEN", KEYS[2])
if count == nil or count == 0 then
    redis.call("ZREM", KEYS[1], KEYS[2])
else
    redis.call("ZADD", KEYS[1], count, KEYS[2])
end
return ret`
)

func (ctrl *RequestController) Run(ctx context.Context) error {
	if ctrl == nil {
		return nil
	}

	go func() {
		timer := time.NewTimer(defaultCheckTimeoutRequestInterval)

		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				reqs, err := ctrl.requestManager.List(ctx, nil, reqManager.ListRequest{
					Page: 1, Count: 200, Retryable: true,
				})
				if err != nil {
					ctrl.logger.Errorf("list request failed, error=%s", err)
					timer.Reset(defaultCheckTimeoutRequestInterval)
					continue
				}

				for _, req := range reqs {
					if err := ctrl.PublishRequest(ctx, nil, req, false); err != nil {
						ctrl.historyCtrl.Publish(ctx, req.GetId(), 0, 0,
							ctrl.logger.Errorf("publish request failed, error=%s", err).ToError().Error())
					}
				}
				timer.Reset(defaultCheckTimeoutRequestInterval)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			ctrl.logger.Error(ctx.Err())
			return ctx.Err()
		default:
			stores, err := redis.Strings(ctrl.redisClient.Do("ZRANGE", config.CrawlStoreList, 0, -1))
			if err != nil {
				ctrl.logger.Errorf("get queues failed, error=%s", err)
				time.Sleep(time.Millisecond * 200)
				continue
			}

			for _, key := range stores {
				host := strings.TrimPrefix(key, config.CrawlRequestStoreQueuePrefix)
				if !ctrl.threadCtrl.TryLock(host) {
					continue
				}

				reqId, err := redis.String(ctrl.redisClient.Do("EVAL", requestPopScript, 2, config.CrawlStoreList, key))
				if err == redis.ErrNil {
					continue
				} else if err != nil {
					ctrl.logger.Errorf("pop data from %s failed, error=%s", key, err)
					continue
				}

				req, err := func() (*request.Request, error) {
					req, err := ctrl.requestManager.GetById(ctx, reqId)
					if err != nil {
						ctrl.logger.Errorf("get request %s failed, error=%s", reqId, err)
						return nil, err
					}
					if req == nil {
						return nil, ctrl.logger.Errorf("request %s not found", reqId).ToError()
					}
					return req, nil
				}()
				if err != nil {
					ctrl.historyCtrl.Publish(ctx, reqId, 0, 0, err.Error())
					if _, err := ctrl.redisClient.Do("SREM", config.CrawlRequestQueueSet, reqId); err != nil {
						ctrl.logger.Errorf("remove req %s cache set failed, error=%s", reqId, err)
					}
					continue
				}

				// lock at most 5mins
				if ctrl.threadCtrl.Lock(req.GetStoreId(), req.GetId(), req.GetOptions().GetMaxTtlPerRequest()) {
					ctrl.logger.Debugf("locked %s %s", req.GetStoreId(), req.GetId())

					// get crawl options
					// TODO: check crawl options
					go func(ctx context.Context, req *request.Request) {
						defer func() {
							ctrl.threadCtrl.Unlock(req.GetStoreId(), req.GetId())
							ctrl.logger.Debugf("unlocked %s %s", req.GetStoreId(), req.GetId())
						}()
						ctrl.logger.Infof("%s %s", req.GetMethod(), req.GetUrl())

						if _, err := ctrl.requestManager.UpdateStatus(ctx, nil, reqId, 2, false); err != nil {
							ctrl.logger.Errorf("update request status failed, error=%s", err)
						}

						err := func() error {
							crawlers, err := ctrl.crawlerCtrl.GetCrawlerByUrl(ctx, req.GetUrl())
							if err != nil {
								ctrl.logger.Errorf("check crawler for %s failed, error=%s", req.GetUrl(), err)
								ctrl.historyCtrl.Publish(ctx, req.GetId(), 0, 0, err.Error())
								return err
							}
							if len(crawlers) == 0 {
								ctrl.logger.Warnf("not crawler found for %s", req.GetUrl())
								ctrl.historyCtrl.Publish(ctx, req.GetId(), 0, 0, "no useable crawler")
								return err
							}
							options := crawlers[0].GetOptions()

							var proxyReq pbProxy.Request
							if err := req.Unmarshal(&proxyReq); err != nil {
								ctrl.logger.Errorf("unmarshal request to proxy.Request failed, error=%s", err)
								ctrl.historyCtrl.Publish(ctx, req.GetId(), 0, 0, err.Error())
								return err
							}

							// EnableProxy, MaxTtlPerRequest set in Unmarshal
							proxyReq.Options.Reliability = options.Reliability
							proxyReq.Options.EnableHeadless = options.EnableHeadless
							proxyReq.Options.JsWaitDuration = options.JsWaitDuration
							proxyReq.Options.EnableSessionInit = options.EnableSessionInit
							proxyReq.Options.KeepSession = options.KeepSession
							proxyReq.Options.DisableCookieJar = options.DisableCookieJar
							proxyReq.Options.DisableRedirect = options.DisableRedirect
							proxyReq.Options.RequestFilterKeys = options.RequestFilterKeys

							reqCtx := context.WithValue(ctx, "tracing_id", req.GetTracingId())
							reqCtx = context.WithValue(reqCtx, "job_id", req.GetJobId())
							reqCtx = context.WithValue(reqCtx, "req_id", req.GetId())

							reqCtx, cancelFunc := context.WithTimeout(reqCtx,
								time.Duration(proxyReq.GetOptions().GetMaxTtlPerRequest())*time.Second)
							defer cancelFunc()

							startTimestamp := time.Now().UnixNano()
							proxyResp, err := ctrl.pigate.Do(reqCtx, &proxyReq)
							duration := int32((time.Now().UnixNano() - startTimestamp) / 1000000)
							if err != nil {
								ctrl.logger.Errorf("do %s request failed, error=%s", req.GetUrl(), err)
								ctrl.historyCtrl.Publish(ctx, req.GetId(), duration, 0, err.Error())
								return err
							}
							ctrl.historyCtrl.Publish(ctx, req.GetId(), duration, proxyResp.StatusCode, "")

							if proxyResp.GetStatusCode() == -1 {
								return errors.New(proxyResp.GetStatus())
							}
							if proxyResp.GetStatusCode() == http.StatusForbidden {
								// clean cached cookie

								if _, err := ctrl.sessionManager.ClearCookies(ctx, &pbSession.ClearCookiesRequest{
									TracingId: req.GetTracingId(),
									Url:       req.GetUrl(),
								}); err != nil {
									ctrl.logger.Errorf("clear cookie for %s failed, error=%s", req.GetUrl(), err)
								}
								// to requeue again
								return errors.New("access forbidden")
							}
							if proxyResp.GetRequest() == nil {
								return fmt.Errorf("request info missing for %s", req.GetUrl())
							}

							// parse respose, if failed, queue the request again
							if err = ctrl.crawlerCtrl.Parse(ctx, req, proxyResp); err != nil {
								ctrl.historyCtrl.Publish(ctx, req.GetId(), 0, 0, err.Error())
								ctrl.logger.Errorf("parse response from %s failed, error=%s", req.GetUrl(), err)
								return err
							}
							return nil
						}()

						var (
							status        int32 = 3
							isSucceed           = true
							isRepublished       = false
						)
						if err != nil {
							isSucceed = false
							if err := ctrl.PublishRequest(ctx, nil, req, true); err != nil {
								ctrl.logger.Errorf("publish request %s failed, error=%s", req.GetId(), err)
								ctrl.historyCtrl.Publish(ctx, req.GetId(), 0, 0, err.Error())
							} else {
								isRepublished = true
							}
						}
						if !isRepublished {
							if _, err := ctrl.requestManager.UpdateStatus(ctx, nil,
								req.GetId(), status, isSucceed); err != nil {

								ctrl.logger.Errorf("update status of request %s failed, error=%s", req.GetId(), err)
								ctrl.historyCtrl.Publish(ctx, req.GetId(), 0, 0, err.Error())
							}
							if _, err := ctrl.redisClient.Do("SREM", config.CrawlRequestQueueSet, req.GetId()); err != nil {
								ctrl.logger.Errorf("remove req cache key failed, error=%s", err)
								ctrl.historyCtrl.Publish(ctx, req.GetId(), 0, 0, err.Error())
							}
						}
					}(ctx, req)
				} else if _, err := ctrl.redisClient.Do("EVAL", requestPushScript, 3,
					config.CrawlStoreList, key, config.CrawlRequestQueueSet, reqId); err != nil {
					ctrl.logger.Errorf("requeue request %s failed, error=%s", reqId, err)
					ctrl.historyCtrl.Publish(ctx, reqId, 0, 0, err.Error())
				}
			}
			if len(stores) == 0 {
				time.Sleep(time.Millisecond * 100)
			}
		}
	}
}
