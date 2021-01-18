package manager

import (
	"context"
	"time"

	"github.com/voiladev/VoilaCrawl/internal/model/request"
	"github.com/voiladev/VoilaCrawl/pkg/types"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/randutil"
	pbError "github.com/voiladev/protobuf/protoc-gen-go/errors"
	"xorm.io/xorm"
)

type RequestManager struct {
	engine *xorm.Engine
	logger glog.Log
}

func NewRequestManager(engine *xorm.Engine, logger glog.Log) (*RequestManager, error) {
	m := RequestManager{
		engine: engine,
		logger: logger.New("RequestManager"),
	}
	return &m, nil
}

func (m *RequestManager) GetById(ctx context.Context, id string) (*request.Request, error) {
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

func (m *RequestManager) Create(ctx context.Context, session *xorm.Session, req *request.Request) (*request.Request, error) {
	if m == nil {
		return nil, nil
	}
	logger := m.logger.New("Create")

	if session == nil {
		session = m.engine.NewSession()
		defer session.Close()
	}

	req.Id = randutil.MustNewRandomID()
	req.Metadata = &types.Request_Metadata{
		CreatedUtc: time.Now().Unix(),
		UpdatedUtc: time.Now().Unix(),
	}
	if _, err := session.Context(ctx).InsertOne(&req.Request); err != nil {
		logger.Errorf("create request failed, error=%s", err)
		return nil, pbError.ErrDatabase.New(err)
	}
	return req, nil
}

func (m *RequestManager) UpdateStatus(ctx context.Context, session *xorm.Session, id string, isSucceed bool, msg string) error {
	if m == nil {
		return nil
	}
	logger := m.logger.New("UpdateStatus")

	if session == nil {
		session = m.engine.NewSession()
		defer session.Close()
	}

	sql := `update request set is_succeed=?,err_msg=? where id=?`
	if _, err := session.Context(ctx).Exec(sql, isSucceed, msg, id); err != nil {
		logger.Errorf("update request status failed, error=%s", err)
		return pbError.ErrDatabase.New(err)
	}
	return nil
}

func (m *RequestManager) UpdateRetry(ctx context.Context, session *xorm.Session, id string) (bool, error) {
	if m == nil {
		return false, nil
	}
	logger := m.logger.New("UpdateRetry")

	if session == nil {
		session = m.engine.NewSession()
		defer session.Close()
	}

	sql := `update request set status_retry_count=status_retry_count+1 where id=? and option_max_retry_count>status_retry_count`
	if ret, err := session.Context(ctx).Exec(sql, id); err != nil {
		logger.Errorf("update request status failed, error=%s", err)
		return false, pbError.ErrDatabase.New(err)
	} else if count, err := ret.RowsAffected(); err != nil {
		logger.Errorf("get affected data failed, error=%s", err)
		return false, pbError.ErrDatabase.New(err)
	} else if count > 0 {
		return true, nil
	}
	return false, nil
}

func (m *RequestManager) Delete(ctx context.Context, session *xorm.Session, id string) error {
	if m == nil {
		return nil
	}
	logger := m.logger.New("Delete")

	if session == nil {
		session = m.engine.NewSession()
		defer session.Close()
	}

	sql := `update request set is_deleted=?,deleted_utc=? where id=?`
	if _, err := session.Context(ctx).Exec(sql, true, time.Now().Unix(), id); err != nil {
		logger.Errorf("delete request failed, error=%s", err)
		return pbError.ErrDatabase.New(err)
	}
	return nil
}
