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
	http "github.com/voiladev/VoilaCrawl/pkg/net/http"
	pbCrawl "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl"
	pbItem "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/strconv"
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
					logger.Infof("reconnect after %s", tryDuration)
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

					for {
						select {
						case <-ctx.Done():
							ctrl.logger.Debugf("done")
							return
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
								return
							}
							logger.Debugf("send heartbeta max: %d, idle: %d", msg.GetMaxConcurrency(), msg.GetIdleConcurrency())
						case msg, ok := <-handler.conn.msgBuffer:
							if !ok {
								return
							}
							if err := handler.client.Send(msg); err != nil {
								logger.Errorf("send msg failed, error=%s", err)
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
					ErrMsg:    err.Error(),
				})
				continue
			}

			var crawler *Crawler
			if crawlers, err := ctrl.crawlerManager.GetByHost(ctx, u.Host); err != nil {
				logger.Errorf("get crawlers by host failed, error=%s", err)
				ctrl.Send(ctx, &pbCrawl.Command_Error{
					TracingId: r.GetTracingId(),
					JobId:     r.GetJobId(),
					ReqId:     r.GetReqId(),
					ErrMsg:    err.Error(),
				})
				continue
			} else {
				for _, c := range crawlers {
					if c.IsUrlMatch(u) {
						crawler = c
						break
					}
				}
			}
			if crawler == nil {
				logger.Errorf("no crawler found for %s", r.Url)
				count, _ := ctrl.crawlerManager.Count(ctx)
				ctrl.Send(ctx, &pbCrawl.Command_Error{
					TracingId: r.GetTracingId(),
					JobId:     r.GetJobId(),
					ReqId:     r.GetReqId(),
					ErrMsg:    fmt.Sprintf("0/%d crawler found", count),
				})
				continue
			}

			jobFunc := func() {
				var (
					shareCtx    context.Context
					sharingData []string
				)
				sharingData = append(sharingData,
					"tracing_id", r.TracingId,
					"job_id", r.JobId,
					"req_id", r.ReqId,
				)
				for _, item := range r.SharingData {
					sharingData = append(sharingData, item.Key, item.Value)
				}
				// share context is used to share data between crawlers
				shareCtx = ctxUtil.WithValues(ctx, sharingData...)

				duration, err := func() (int64, error) {
					req, err := NewRequest(r)
					if err != nil {
						return 0, err
					}
					req = crawler.SetHeader(req)

					maxTtlPerRequest := defaultTtlPerRequest
					if r.Options.MaxTtlPerRequest > 0 {
						maxTtlPerRequest = r.Options.MaxTtlPerRequest
					}
					requestCtx, cancel := context.WithTimeout(shareCtx, time.Duration(maxTtlPerRequest)*time.Second)
					defer cancel()

					startTime := time.Now()
					resp, err := ctrl.httpClient.DoWithOptions(requestCtx, req, http.Options{
						EnableProxy:    !r.Options.DisableProxy,
						EnableHeadless: crawler.CrawlOptions().EnableHeadless,
					})
					duration := (time.Now().UnixNano() - startTime.UnixNano()) / 1000000 // in millseconds
					if err != nil {
						return duration, err
					}
					defer resp.Body.Close()

					return duration, crawler.Parse(shareCtx, resp, func(c context.Context, i interface{}) error {
						sharingData := ctxUtil.RetrieveAllValues(c)
						switch val := i.(type) {
						case *http.Request:
							// convert http.Request to pbCrawl.Command_Request and forward
							subreq := pbCrawl.Command_Request{
								TracingId:     r.GetTracingId(),
								JobId:         r.GetJobId(),
								ReqId:         r.GetReqId(),
								Url:           val.URL.String(),
								Method:        val.Method,
								Parent:        r,
								CustomHeaders: map[string]string{},
								CustomCookies: r.CustomCookies,
								Options:       r.Options,
								SharingData:   r.SharingData,
							}
							if val.Body != nil {
								defer val.Body.Close()
								if data, err := ioutil.ReadAll(val.Body); err != nil {
									return err
								} else {
									subreq.Body = fmt.Sprintf("%s", data)
								}
							}
							// over write header
							for k := range val.Header {
								subreq.CustomHeaders[k] = val.Header.Get(k)
							}
							for k, v := range r.CustomHeaders {
								if _, ok := subreq.CustomHeaders[k]; ok {
									continue
								}
								subreq.CustomHeaders[k] = v
							}

							for k, v := range sharingData {
								key, ok1 := k.(string)
								val, ok2 := v.(string)
								if !ok1 || !ok2 {
									continue
								}

								if strings.HasSuffix(key, "tracing_id") ||
									strings.HasSuffix(key, "job_id") ||
									strings.HasSuffix(key, "req_id") {
									continue
								}
								found := false
								for _, item := range subreq.SharingData {
									if item.Key == key {
										item.Value = val
										found = true
										break
									}
								}
								if !found {
									subreq.SharingData = append(subreq.SharingData, &pbCrawl.Command_Request_KeyValue{
										Key: key, Value: val,
									})
								}
							}

							cmd := pbCrawl.Command{Timestamp: time.Now().UnixNano(), NodeId: NodeId()}
							cmd.Data, _ = anypb.New(&subreq)
							ctrl.Send(shareCtx, &cmd)
						case *pbItem.Product:
							var index int64
							if indexVal, ok := sharingData["item.index"]; ok && indexVal != nil {
								ctrl.logger.Errorf("%v", indexVal)
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
							item.Data, _ = anypb.New(val)

							ctrl.logger.Infof("#####################", ctrl.Send(shareCtx, &item))
						default:
							return errors.New("unsupported response data type")
						}
						return nil
					})
				}()

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
