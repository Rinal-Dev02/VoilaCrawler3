package request

import (
	"context"
	"errors"
	"time"

	"github.com/nsqio/go-nsq"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/voiladev/VoilaCrawl/internal/pkg/config"
	pbEvent "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/event"
	pbCrawl "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl"
	"github.com/voiladev/go-framework/glog"
)

type Sender interface {
	Send(ctx context.Context, msg protoreflect.ProtoMessage) error
}

type RequestControllerOptions struct {
	NsqLookupdAddresses []string
}

type ReqeustController struct {
	ctx      context.Context
	sender   Sender
	consumer *nsq.Consumer
	options  RequestControllerOptions
	logger   glog.Log
}

func NewRequestController(
	ctx context.Context,
	sender Sender,
	options *RequestControllerOptions,
	logger glog.Log) (*ReqeustController, error) {
	if sender == nil {
		return nil, errors.New("invalid sender")
	}
	if options == nil {
		return nil, errors.New("invalid options")
	}
	if len(options.NsqLookupdAddresses) == 0 {
		return nil, errors.New("invalid nsq lookup address")
	}
	ctrl := ReqeustController{
		ctx:     ctx,
		sender:  sender,
		options: *options,
		logger:  logger.New("RequestController"),
	}

	var (
		err error
	)
	if ctrl.consumer, err = nsq.NewConsumer(config.CrawlRequestTopic, "publisher", nsq.NewConfig()); err != nil {
		ctrl.logger.Errorf("create consumer failed, error=%s", err)
		return nil, err
	}
	ctrl.consumer.AddConcurrentHandlers(&RequestHandler{ctx: ctx, sender: sender, logger: logger}, 3)
	if err = ctrl.consumer.ConnectToNSQLookupds(ctrl.options.NsqLookupdAddresses); err != nil {
		ctrl.logger.Errorf("connect to nsq fialed, error=%s", err)
		return nil, err
	}
	return &ctrl, nil
}

type RequestHandler struct {
	ctx    context.Context
	sender Sender
	logger glog.Log
}

func (h *RequestHandler) HandleMessage(msg *nsq.Message) error {
	if h == nil {
		return nil
	}
	msg.DisableAutoResponse()

	var (
		event pbEvent.Event
		req   pbCrawl.Command_Request
	)
	if err := proto.Unmarshal(msg.Body, &event); err != nil {
		h.logger.Errorf("unmarshal event data failed, error=%s", err)
		msg.RequeueWithoutBackoff(time.Minute)
	}
	if err := proto.Unmarshal(event.GetData(), &req); err != nil {
		h.logger.Errorf("unmarshal request data failed, error=%s", err)
		msg.RequeueWithoutBackoff(time.Minute)
		return err
	}
	if h.sender == nil {
		msg.RequeueWithoutBackoff(time.Minute)
		return nil
	}
	if err := h.sender.Send(h.ctx, &req); err != nil {
		msg.RequeueWithoutBackoff(time.Minute)
		return err
	}
	return nil
}

func (h *RequestHandler) LogFailedMessage(msg *nsq.Message) {
	if h == nil {
		return
	}

	var (
		event pbEvent.Event
		req   pbCrawl.Command_Request
	)
	if err := proto.Unmarshal(msg.Body, &event); err != nil {
		h.logger.Errorf("unmarshal event data failed, error=%s", err)
		return
	}
	if err := proto.Unmarshal(event.GetData(), &req); err != nil {
		h.logger.Errorf("unmarshal request data failed, error=%s", err)
		return
	}
	data, _ := protojson.Marshal(&req)
	h.logger.Errorf("submit request failed: %s", data)
}
