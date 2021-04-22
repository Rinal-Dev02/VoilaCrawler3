package history

import (
	"context"
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/voiladev/VoilaCrawl/pkg/types"
	pbCrawl "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl"
	"github.com/voiladev/go-framework/glog"
	pbError "github.com/voiladev/protobuf/protoc-gen-go/errors"
	"google.golang.org/protobuf/proto"
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

	ctx, cancel := context.WithTimeout(h.ctrl.ctx, time.Minute)
	defer cancel()

	var errData pbCrawl.Error
	if err := proto.Unmarshal(msg.Body, &errData); err != nil {
		h.logger.Errorf("invalid data format, error=%s", err)
		return err
	}

	his := types.RequestHistory{
		Id:        errData.GetReqId(),
		Timestamp: errData.GetTimestamp(),
		TracingId: errData.GetTracingId(),
		StoreId:   errData.GetStoreId(),
		JobId:     errData.GetJobId(),
		Code:      errData.GetCode(),
		ErrMsg:    errData.GetErrMsg(),
		Duration:  int32(errData.GetDuration()),
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
