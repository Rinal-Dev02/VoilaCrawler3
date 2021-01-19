package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"time"

	ctxUtil "github.com/voiladev/VoilaCrawl/pkg/context"
	http "github.com/voiladev/VoilaCrawl/pkg/net/http"
	pbCrawl "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl"
	pbItem "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl/item"
	"github.com/voiladev/go-framework/glog"
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
	handler        *ChannelHandler

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

	var err error
	ctrl.gpool, err = NewGPool(ctx, options.MaxConcurrency, logger)
	if ctrl.handler, err = ctrl.conn.NewChannelHandler(ctrl.ctx, &ctrl); err != nil {
		ctrl.logger.Errorf("new channel handler failed, error=%s", err)
		return nil, err
	}
	go func() {
		for {
			// reconnect after failed
			time.Sleep(time.Second)

			ctrl.handler.Watch(ctrl.ctx, func(ctx context.Context, req *pbCrawl.Command_Request) {
				select {
				case <-ctx.Done():
					ctrl.logger.Error(ctx.Err())
					return
				case ctrl.requestBuffer <- req:
				}
				return
			})
		}
	}()
	return &ctrl, nil
}

const (
	defaultTtlPerRequest int32 = 300 // seconds
)

func (ctrl *CrawlerController) Run(ctx context.Context) error {
	if ctrl == nil {
		return nil
	}
	logger := ctrl.logger.New("Run")

	if err := ctrl.handler.Send(ctx, &pbCrawl.Join_Ping{
		Timestamp: time.Now().UnixNano(),
		Node: &pbCrawl.Join_Ping_Node{
			Id:              NodeId(),
			Host:            Hostname(),
			MaxConcurrency:  ctrl.gpool.MaxConcurrency(),
			IdleConcurrency: ctrl.gpool.MaxConcurrency() - ctrl.gpool.CurrentConcurrency(),
		},
		Crawlers: []*pbCrawl.Crawler{},
	}); err != nil {
		ctrl.logger.Errorf("register node failed, error=%s", err)
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case r, ok := <-ctrl.requestBuffer:
			if !ok {
				return nil
			}

			// TODO: added tracingId, jobId, reqId check

			u, err := url.Parse(r.Url)
			if err != nil {
				logger.Errorf("parse url %s failed, error=%s", err)
				ctrl.handler.Send(ctx, &pbCrawl.Command_Error{
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
				ctrl.handler.Send(ctx, &pbCrawl.Command_Error{
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
				ctrl.handler.Send(ctx, &pbCrawl.Command_Error{
					TracingId: r.GetTracingId(),
					JobId:     r.GetJobId(),
					ReqId:     r.GetReqId(),
					ErrMsg:    "no crawler found",
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

				if err := func() error {
					req, err := NewRequest(r)
					if err != nil {
						return err
					}
					req = crawler.SetHeader(req)

					maxTtlPerRequest := defaultTtlPerRequest
					if r.Options.MaxTtlPerRequest > 0 {
						maxTtlPerRequest = r.Options.MaxTtlPerRequest
					}
					requestCtx, cancel := context.WithTimeout(shareCtx, time.Duration(maxTtlPerRequest)*time.Second)
					defer cancel()

					// here do with http request
					resp, err := ctrl.httpClient.DoWithOptions(requestCtx, req, http.Options{
						EnableProxy:    !crawler.CrawlOptions().DisableProxy,
						EnableHeadless: crawler.CrawlOptions().EnableHeadless,
					})
					if err != nil {
						return err
					}
					defer resp.Body.Close()

					return crawler.Parse(shareCtx, resp, func(c context.Context, i interface{}) error {
						switch val := i.(type) {
						case *http.Request:
							// convert http.Request to pbCrawl.Command_Request and forward
							subreq := pbCrawl.Command_Request{
								TracingId:   r.GetTracingId(),
								JobId:       r.GetJobId(),
								ReqId:       r.GetReqId(),
								Url:         val.URL.String(),
								Method:      val.Method,
								Parent:      r,
								Options:     r.Options,
								SharingData: r.SharingData,
							}
							if val.Body != nil {
								defer val.Body.Close()
								if data, err := ioutil.ReadAll(val.Body); err != nil {
									return err
								} else {
									subreq.Body = fmt.Sprintf("%s", data)
								}
							}
							for k, v := range ctxUtil.RetrieveAllValues(c) {
								key, ok1 := k.(string)
								val, ok2 := v.(string)
								if !ok1 || !ok2 {
									continue
								}
								if key == "tracing_id" || key == "job_id" || key == "req_id" {
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
							ctrl.handler.Send(shareCtx, &cmd)
						case *pbItem.Product:
							item := pbCrawl.Command_Item{
								TracingId: r.GetTracingId(),
								JobId:     r.GetJobId(),
								ReqId:     r.GetReqId(),
							}
							item.Data, _ = anypb.New(val)

							cmd := pbCrawl.Command{Timestamp: time.Now().UnixNano(), NodeId: NodeId()}
							cmd.Data, _ = anypb.New(&item)
							ctrl.handler.Send(shareCtx, &cmd)
						default:
							return errors.New("unsupported response data type")
						}
						return nil
					})
				}(); err != nil {
					ctrl.handler.Send(ctx, &pbCrawl.Command_Error{
						TracingId: r.GetTracingId(),
						JobId:     r.GetJobId(),
						ReqId:     r.GetReqId(),
						ErrMsg:    err.Error(),
					})
				}
			}

			ctrl.gpool.DoJob(jobFunc)
		}
	}
}
