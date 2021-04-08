package request

import (
	"context"
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/voiladev/VoilaCrawl/internal/model/request"
	pbCrawl "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl"
	"github.com/voiladev/go-framework/glog"
	pbError "github.com/voiladev/protobuf/protoc-gen-go/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var (
	_ nsq.Handler             = (*RequestHander)(nil)
	_ nsq.FailedMessageLogger = (*RequestHander)(nil)
)

// RequestHander use to handle new requests emit by crawlet
type RequestHander struct {
	ctrl   *RequestController
	logger glog.Log
}

func (h *RequestHander) HandleMessage(msg *nsq.Message) error {
	if h == nil {
		return nil
	}
	msg.DisableAutoResponse()

	ctx, cancel := context.WithTimeout(h.ctrl.ctx, time.Minute)
	defer cancel()

	var creq pbCrawl.Command_Request
	if err := proto.Unmarshal(msg.Body, &creq); err != nil {
		h.logger.Errorf("unmarshal request failed, error=%s", err)
		msg.Finish()
		return err
	}
	req, err := request.NewRequest(&creq)
	if err != nil {
		h.logger.Errorf("load request instance failed, error=%s", err)
		msg.RequeueWithoutBackoff(time.Second * 10)
		return err
	}

	session := h.ctrl.engine.NewSession()
	defer session.Close()

	if req, err = h.ctrl.requestManager.Create(ctx, session, req); err != nil {
		if err == pbError.ErrAlreadyExists {
			h.logger.Warnf("request %s already exists", creq.GetUrl())
			msg.Finish()
			return nil
		}
		h.logger.Errorf("save request %s failed, error=%s", creq.GetUrl(), err)
		msg.Requeue(time.Second * 5)
		return err
	}
	msg.Finish()

	if err := session.Begin(); err != nil {
		h.logger.Errorf("begin tx failed, error=%s", err)
		return err
	}

	if err := h.ctrl.PublishRequest(ctx, session, req, true); err != nil {
		h.logger.Errorf("publish request failed, error=%s", err)
		session.Rollback()
		return err
	}

	if err := session.Commit(); err != nil {
		h.logger.Errorf("commit tx failed, error=%s", err)
		session.Rollback()
		return err
	}
	return nil
}

func (h *RequestHander) LogFailedMessage(msg *nsq.Message) {
	if h == nil {
		return
	}

	var creq pbCrawl.Command_Request
	if err := proto.Unmarshal(msg.Body, &creq); err != nil {
		h.logger.Errorf("unmarshal request failed, error=%s", err)
		return
	}

	data, _ := protojson.Marshal(&creq)
	h.logger.Errorf("process msg %s failed", data)
}
