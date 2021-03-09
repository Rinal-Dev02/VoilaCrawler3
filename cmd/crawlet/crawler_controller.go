package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"
	"time"

	ctxUtil "github.com/voiladev/VoilaCrawl/pkg/context"
	crawlerSpec "github.com/voiladev/VoilaCrawl/pkg/crawler"
	http "github.com/voiladev/VoilaCrawl/pkg/net/http"
	pbCrawl "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
)

type CrawlerControllerOptions struct {
	MaxConcurrency int32
}

type CrawlerController struct {
	ctx            context.Context
	crawlerManager *CrawlerManager
	httpClient     http.Client
	conn           *Connection

	gpool         *GPool
	requestBuffer chan *pbCrawl.Command_Request

	options CrawlerControllerOptions
	logger  glog.Log
}

func NewCrawlerController(
	ctx context.Context,
	crawlerManager *CrawlerManager,
	httpClient http.Client,
	conn *Connection,
	options *CrawlerControllerOptions,
	logger glog.Log,
) (*CrawlerController, error) {
	ctrl := CrawlerController{
		ctx:            ctx,
		crawlerManager: crawlerManager,
		httpClient:     httpClient,
		conn:           conn,
		requestBuffer:  make(chan *pbCrawl.Command_Request),
		options:        *options,
		logger:         logger.New("CrawlerController"),
	}

	ctrl.gpool, _ = NewGPool(ctx, options.MaxConcurrency, logger)
	go func() {
		var (
			err         error
			handler     *ChannelHandler
			tryDuration time.Duration = time.Second
			logger                    = ctrl.logger.New("EventLoop")
		)
		for {
			// reconnect after failed
			logger.Infof("reconnect after %s", tryDuration)
			time.Sleep(tryDuration)

			func() {
				ctx, cancel := context.WithCancel(ctrl.ctx)
				defer cancel()

				if handler, err = ctrl.conn.NewChannelHandler(ctx, &ctrl); err != nil {
					ctrl.logger.Errorf("new channel handler failed, error=%s", err)

					tryDuration += time.Second
					if tryDuration > time.Minute {
						tryDuration = time.Second * 10
					}
					return
				}
				tryDuration = time.Second

				if err = ctrl.Send(ctx, &pbCrawl.Join_Ping{
					Timestamp: time.Now().UnixNano(),
					Node: &pbCrawl.Join_Ping_Node{
						Id:              NodeId(),
						Host:            Hostname(),
						MaxConcurrency:  ctrl.gpool.MaxConcurrency(),
						IdleConcurrency: ctrl.gpool.MaxConcurrency() - ctrl.gpool.CurrentConcurrency(),
					},
					Crawlers: []*pbCrawl.Crawler{},
				}); err != nil {
					logger.Errorf("register node failed, error=%s", err)
					return
				}

				go func() {
					defer cancel()

					timeTicker := time.NewTicker(time.Second * 10)
					for {
						select {
						case <-ctx.Done():
							ctrl.logger.Debugf("done")
							return
						case <-timeTicker.C:
							logger.Infof("send heartbeta max: %d, idle: %d",
								handler.ctrl.gpool.MaxConcurrency(),
								handler.ctrl.gpool.MaxConcurrency()-handler.ctrl.gpool.CurrentConcurrency(),
							)
						case <-handler.heartbeatTicker.C:
							msg := pbCrawl.Heartbeat_Ping{
								Timestamp:       time.Now().UnixNano(),
								NodeId:          NodeId(),
								MaxConcurrency:  handler.ctrl.gpool.MaxConcurrency(),
								IdleConcurrency: handler.ctrl.gpool.MaxConcurrency() - handler.ctrl.gpool.CurrentConcurrency(),
							}
							anydata, _ := anypb.New(&msg)
							if err := handler.client.Send(anydata); err != nil {
								logger.Errorf("send heartbeat failed, error=%s", err)
								// TODO: end
								return
							}
						case msg, ok := <-handler.conn.msgBuffer:
							if !ok {
								return
							}
							if err := handler.client.Send(msg); err != nil {
								logger.Errorf("send msg failed, error=%s", err)
								return
							}
						}
					}
				}()

				if err = handler.Watch(ctx, func(ctx context.Context, req *pbCrawl.Command_Request) {
					select {
					case <-ctx.Done():
						logger.Error(ctx.Err())
						return
					case ctrl.requestBuffer <- req:
					}
					return
				}); err != nil {
					logger.Errorf("watch failed, error=%s", err)
				}
				handler.Close()
			}()
		}
	}()
	return &ctrl, nil
}

func (ctrl *CrawlerController) Send(ctx context.Context, msg protoreflect.ProtoMessage) error {
	if ctrl == nil || msg == nil {
		return nil
	}

	switch v := msg.(type) {
	case *pbCrawl.Command_Error:
		cmd := pbCrawl.Command{Timestamp: time.Now().UnixNano(), NodeId: NodeId()}
		cmd.Data, _ = anypb.New(v)
		msg = &cmd
	case *pbCrawl.Command_Request:
		cmd := pbCrawl.Command{Timestamp: time.Now().UnixNano(), NodeId: NodeId()}
		cmd.Data, _ = anypb.New(v)
		msg = &cmd
	}
	anydata, err := anypb.New(msg)
	if err != nil {
		ctrl.logger.Errorf("marshal msg type failed, error=%s", err)
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case ctrl.conn.msgBuffer <- anydata:
	}
	return nil
}

const (
	defaultTtlPerRequest int32 = 300 // seconds
)

func (ctrl *CrawlerController) Run(ctx context.Context) error {
	if ctrl == nil {
		return nil
	}
	logger := ctrl.logger.New("Run")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case r, ok := <-ctrl.requestBuffer:
			if !ok {
				return nil
			}

			u, err := url.Parse(r.Url)
			if err != nil {
				logger.Errorf("parse url %s failed, error=%s", err)
				ctrl.Send(ctx, &pbCrawl.Command_Error{
					TracingId: r.GetTracingId(),
					JobId:     r.GetJobId(),
					ReqId:     r.GetReqId(),
					StoreId:   r.GetStoreId(),
					ErrMsg:    err.Error(),
				})
				continue
			}

			var crawlers []*Crawler
			if crawlers, err = ctrl.crawlerManager.GetByHost(ctx, u.Host); err != nil {
				logger.Errorf("get crawlers by host failed, error=%s", err)
				ctrl.Send(ctx, &pbCrawl.Command_Error{
					TracingId: r.GetTracingId(),
					JobId:     r.GetJobId(),
					ReqId:     r.GetReqId(),
					StoreId:   r.GetStoreId(),
					ErrMsg:    err.Error(),
				})
				continue
			}
			if len(crawlers) == 0 {
				logger.Errorf("no crawler found for %s", r.Url)
				count, _ := ctrl.crawlerManager.Count(ctx)
				ctrl.Send(ctx, &pbCrawl.Command_Error{
					TracingId: r.GetTracingId(),
					JobId:     r.GetJobId(),
					ReqId:     r.GetReqId(),
					StoreId:   r.GetStoreId(),
					ErrMsg:    fmt.Sprintf("0/%d crawler found", count),
				})
				continue
			}

			jobFunc := func() {
				var (
					shareCtx    context.Context
					sharingData []string
					duration    int64
					err         error
				)
				sharingData = append(sharingData,
					"tracing_id", r.TracingId,
					"job_id", r.JobId,
					"req_id", r.ReqId,
					"store_id", r.StoreId,
				)
				for k, v := range r.SharingData {
					sharingData = append(sharingData, k, v)
				}
				// share context is used to share data between crawlers
				shareCtx = ctxUtil.WithValues(ctx, sharingData...)

				for _, crawler := range crawlers {
					duration, err = func(crawler *Crawler) (int64, error) {
						req, err := NewRequest(r)
						if err != nil {
							logger.Debug(err)
							return 0, err
						}
						req = crawler.SetHeader(req)

						maxTtlPerRequest := defaultTtlPerRequest
						if r.Options.MaxTtlPerRequest > 0 {
							maxTtlPerRequest = r.Options.MaxTtlPerRequest
						}

						requestCtx, cancel := context.WithTimeout(shareCtx, time.Duration(maxTtlPerRequest)*time.Second+time.Minute*10)
						defer cancel()

						startTime := time.Now()
						resp, err := ctrl.httpClient.DoWithOptions(requestCtx, req, http.Options{
							EnableProxy:       !r.Options.DisableProxy,
							EnableHeadless:    crawler.CrawlOptions().EnableHeadless,
							EnableSessionInit: crawler.CrawlOptions().EnableSessionInit,
							KeepSession:       crawler.CrawlOptions().KeepSession,
							DisableRedirect:   crawler.CrawlOptions().DisableRedirect,
							Reliability:       crawler.CrawlOptions().Reliability,
						})
						duration := (time.Now().UnixNano() - startTime.UnixNano()) / 1000000 // in millseconds
						if err != nil {
							logger.Infof("Access %s error=%s", req.URL.String(), err)
							return duration, err
						}
						logger.Infof("Access %s %d", req.URL.String(), resp.StatusCode)
						defer resp.Body.Close()

						return duration, crawler.Parse(shareCtx, resp, func(c context.Context, i interface{}) error {
							sharingData := ctxUtil.RetrieveAllValues(c)
							switch val := i.(type) {
							case *http.Request:
								if val.URL.Host == "" {
									val.URL.Scheme = req.URL.Scheme
									val.URL.Host = req.URL.Host
								} else if val.URL.Scheme != "http" && req.URL.Scheme != "https" {
									val.URL.Scheme = req.URL.Scheme
								}

								if val.Header.Get("Referer") == "" && resp.Request != nil {
									val.Header.Set("Referer", resp.Request.URL.String())
								}

								// convert http.Request to pbCrawl.Command_Request and forward
								subreq := pbCrawl.Command_Request{
									TracingId:     r.GetTracingId(),
									JobId:         r.GetJobId(),
									ReqId:         r.GetReqId(),
									StoreId:       r.GetStoreId(),
									Url:           val.URL.String(),
									Method:        val.Method,
									Parent:        r,
									CustomHeaders: r.CustomHeaders,
									CustomCookies: r.CustomCookies,
									Options:       r.Options,
									SharingData:   r.SharingData,
								}
								if subreq.CustomHeaders == nil {
									subreq.CustomHeaders = make(map[string]string)
								}
								if subreq.SharingData == nil {
									subreq.SharingData = map[string]string{}
								}

								if val.Body != nil {
									defer val.Body.Close()
									if data, err := ioutil.ReadAll(val.Body); err != nil {
										return err
									} else {
										subreq.Body = fmt.Sprintf("%s", data)
									}
								}

								for k := range val.Header {
									subreq.CustomHeaders[k] = val.Header.Get(k)
								}

								for k, v := range sharingData {
									key, ok := k.(string)
									if !ok {
										continue
									}
									val := strconv.Format(v)

									if strings.HasSuffix(key, "tracing_id") ||
										strings.HasSuffix(key, "job_id") ||
										strings.HasSuffix(key, "req_id") ||
										strings.HasSuffix(key, "store_id") {
										continue
									}
									subreq.SharingData[key] = val
								}
								cmd := pbCrawl.Command{Timestamp: time.Now().UnixNano(), NodeId: NodeId()}
								cmd.Data, _ = anypb.New(&subreq)
								return ctrl.Send(shareCtx, &cmd)
							default:
								msg, ok := i.(proto.Message)
								if !ok {
									return errors.New("unsupported response data type")
								}
								var index int64
								if indexVal, ok := sharingData["item.index"]; ok && indexVal != nil {
									index = strconv.MustParseInt(indexVal)
								}
								item := pbCrawl.Item{
									Timestamp: time.Now().UnixNano(),
									NodeId:    NodeId(),
									TracingId: r.GetTracingId(),
									JobId:     r.GetJobId(),
									ReqId:     r.GetReqId(),
									Index:     int32(index),
								}
								item.Data, _ = anypb.New(msg)

								return ctrl.Send(shareCtx, &item)
							}
						})
					}(crawler)
					if err != crawlerSpec.ErrNotSupportedPath {
						continue
					}
					break
				}

				var (
					errMsg string
				)
				if err != nil {
					errMsg = err.Error()
				}
				if err := ctrl.Send(ctx, &pbCrawl.Command_Error{
					TracingId: r.GetTracingId(),
					JobId:     r.GetJobId(),
					ReqId:     r.GetReqId(),
					StoreId:   r.GetStoreId(),
					Duration:  duration,
					IsSucceed: err == nil,
					ErrMsg:    errMsg,
				}); err != nil {
					logger.Error("send feedback failed, error=%s", err)
				}
			}
			ctrl.gpool.DoJob(jobFunc)
		}
	}
}
