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
	pbError "github.com/voiladev/protobuf/protoc-gen-go/errors"
)

type Sender interface {
	Send(ctx context.Context, msg protoreflect.ProtoMessage) error
}

type RequestControllerOptions struct {
	NsqdAddress         string
	NsqLookupdAddresses []string
}

type RequestController struct {
	ctx             context.Context
	producer        *nsq.Producer
	requestConsumer *nsq.Consumer
	options         RequestControllerOptions
	logger          glog.Log
}

func NewRequestController(
	ctx context.Context,
	sender Sender,
	options *RequestControllerOptions,
	logger glog.Log) (*RequestController, error) {
	if sender == nil {
		return nil, errors.New("invalid sender")
	}
	if options == nil {
		return nil, errors.New("invalid options")
	}
	if len(options.NsqLookupdAddresses) == 0 {
		return nil, errors.New("invalid nsq lookup address")
	}
	ctrl := RequestController{
		ctx:     ctx,
		options: *options,
		logger:  logger.New("RequestController"),
	}

	var (
		err  error
		conf = nsq.NewConfig()
	)
	conf.MaxAttempts = 20
	conf.MaxBackoffDuration = time.Hour
	if ctrl.requestConsumer, err = nsq.NewConsumer(config.CrawlRequestTopic, "crawl-api", conf); err != nil {
		ctrl.logger.Errorf("create consumer failed, error=%s", err)
		return nil, err
	}
	ctrl.requestConsumer.AddHandler(&RequestHandler{ctx: ctx, sender: sender, logger: logger})
	if err = ctrl.requestConsumer.ConnectToNSQLookupds(ctrl.options.NsqLookupdAddresses); err != nil {
		ctrl.logger.Errorf("connect to nsq fialed, error=%s", err)
		panic(err)
	}
	return &ctrl, nil
}

type RequestHandler struct {
	ctx    context.Context
	ctrl   *RequestController
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
		return err
	}
	if err := proto.Unmarshal(event.GetData(), &req); err != nil {
		h.logger.Errorf("unmarshal request data failed, error=%s", err)
		return err
	}
	if h.sender == nil {
		msg.Requeue(time.Second * 30 * time.Duration(msg.Attempts+1))
		return nil
	}

	if err := h.sender.Send(h.ctx, &req); err == pbError.ErrUnavailable {
		// NOTE: for service unavailable, backoff for 5mins
		// there exists case that if the crawlet always offline, some message may dropped
		msg.Requeue(time.Second * 30 * time.Duration(msg.Attempts+1))
		return nil
	} else if err != nil {
		return err
	}
	msg.Finish()
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

	// requeue for failed message
	if err := h.ctrl.producer.Publish(config.CrawlRequestTopic, msg.Body); err != nil {
		h.logger.Errorf("republish %s failed, error=%s", data, err)
	}
	msg.Finish()
}
