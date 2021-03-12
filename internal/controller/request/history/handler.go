package history

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/voiladev/VoilaCrawl/pkg/types"
	"github.com/voiladev/go-framework/glog"
	pbError "github.com/voiladev/protobuf/protoc-gen-go/errors"
)

var (
	_ nsq.Handler             = (*RequestHistoryHandler)(nil)
	_ nsq.FailedMessageLogger = (*RequestHistoryHandler)(nil)
)

// RequestHistoryHandler use to handle new requests emit by crawlet
type RequestHistoryHandler struct {
	ctrl   *RequestHistoryController
	logger glog.Log
}

func (h *RequestHistoryHandler) HandleMessage(msg *nsq.Message) error {
	if h == nil {
		return nil
	}
	msg.DisableAutoResponse()

	ctx, cancel := context.WithTimeout(h.ctrl.ctx, time.Minute)
	defer cancel()

	var his types.RequestHistory
	if err := json.Unmarshal(msg.Body, &his); err != nil {
		h.logger.Errorf("unmarshal request failed, error=%s", err)
		msg.Finish()
		return err
	}

	if err := h.ctrl.historyManager.Save(ctx, nil, &his); err != nil {
		e := pbError.NewFromError(err)
		if e.Code() == int(pbError.Code_Internal) {
			msg.RequeueWithoutBackoff(time.Second * 10)
			return err
		}
		msg.Finish()
		return err
	}
	msg.Finish()
	return nil
}

func (h *RequestHistoryHandler) LogFailedMessage(msg *nsq.Message) {
	if h == nil {
		return
	}
	h.logger.Errorf("process request history %s failed", msg.Body)
}
