package manager

import (
	"context"
	"time"

	"github.com/voiladev/VoilaCrawl/internal/model/request"
	"github.com/voiladev/VoilaCrawl/pkg/types"
	"github.com/voiladev/go-framework/glog"
	pbError "github.com/voiladev/protobuf/protoc-gen-go/errors"
	"xorm.io/xorm"
)

type HistoryManager struct {
	engine *xorm.Engine
	logger glog.Log
}

func NewHistoryManager(engine *xorm.Engine, logger glog.Log) (*HistoryManager, error) {
	m := HistoryManager{
		engine: engine,
		logger: logger.New("HistoryManager"),
	}
	return &m, nil
}

func (m *HistoryManager) GetById(ctx context.Context, id string) (*request.Request, error) {
	if m == nil {
		return nil, nil
	}
	logger := m.logger.New("GetById")

	var req request.Request
	if exists, err := m.engine.Context(ctx).ID(id).Get(&req.Request); err != nil {
		logger.Errorf("get request by id %s failed, error=%s", err)
		return nil, pbError.ErrDatabase.New(err)
	} else if exists {
		return &req, nil
	}
	return nil, nil
}

func (m *HistoryManager) Save(ctx context.Context, session *xorm.Session, his *types.RequestHistory) error {
	if m == nil || his == nil {
		return nil
	}
	logger := m.logger.New("Save")

	if his.GetId() == "" {
		return pbError.ErrInvalidArgument.New("invalid request id")
	}
	if his.Timestamp == 0 {
		his.Timestamp = time.Now().UnixNano() / 1000000
	}

	if session == nil {
		session = m.engine.NewSession()
		defer session.Close()
	}

	sql := `INSERT INTO request_history (id,timestamp,tracing_id,store_id,job_id,code,err_msg,duration) VALUES (?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE code=?,err_msg=?,duration=?`
	if _, err := session.Context(ctx).Exec(sql,
		his.Id, his.Timestamp, his.TracingId, his.StoreId, his.JobId, his.Code, his.ErrMsg, his.Duration,
		his.Code, his.ErrMsg, his.Duration,
	); err != nil {
		logger.Errorf("save request history failed, error=%s", err)
		return pbError.ErrDatabase.New(err)
	}
	return nil
}
