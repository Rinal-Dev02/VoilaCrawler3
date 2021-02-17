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

type ListRequest struct {
	Page         int32
	Count        int32
	TracingId    string
	JobId        string
	ExpireStatus int32 // 0-all, 1-not expired, 2-expired
	Retryable    bool
}

func (m *RequestManager) List(ctx context.Context, session *xorm.Session, req ListRequest) ([]*request.Request, error) {
	if m == nil {
		return nil, nil
	}
	logger := m.logger.New("List")

	if req.Page <= 0 || req.Count <= 0 {
		return nil, pbError.ErrInvalidArgument.New("invalid pagination info")
	}

	if session == nil {
		session = m.engine.NewSession()
		defer session.Close()
	}

	handler := m.engine.Context(ctx)
	if req.TracingId != "" {
		handler = handler.Where("tracing_id=?", req.TracingId)
	}
	if req.JobId != "" {
		handler = handler.And("job_id=?", req.JobId)
	}
	if req.Retryable {
		handler = handler.And("status_retry_count < option_max_retry_count")
	}

	if req.ExpireStatus == 1 {
		// TODO
	} else if req.ExpireStatus == 2 {
		const retryPaddingInterval = 180
		handler = handler.And("(is_succeed=0 and start_utc+option_max_ttl_per_request>?)", time.Now().Unix()+retryPaddingInterval)
	} else if req.ExpireStatus != 0 {
		return nil, pbError.ErrInvalidArgument.New("invalid expire status")
	}

	var reqs []*types.Request
	if err := handler.And("is_deleted=0").Limit(int(req.Count), int((req.Page-1)*req.Count)).Find(&reqs); err != nil {
		logger.Errorf("get requests failed, error=%s", err)
		return nil, pbError.ErrDatabase.New(err)
	}
	var ret []*request.Request
	for _, req := range reqs {
		ret = append(ret, &request.Request{Request: *req})
	}
	return ret, nil
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

// UpdateStatus  status: 1-queued, 2-inprocess, 3-processed
func (m *RequestManager) UpdateStatus(ctx context.Context, session *xorm.Session, id string, status int32, duration int64, isSucceed bool, msg string) (bool, error) {
	if m == nil {
		return false, nil
	}
	logger := m.logger.New("UpdateStatus")

	if session == nil {
		session = m.engine.NewSession()
		defer session.Close()
	}

	var (
		t    = time.Now().Unix()
		sql  string
		vals = []interface{}{""}
	)

	switch status {
	case 1:
		sql = `update request set status=1,status_retry_count=status_retry_count+1,start_utc=0,end_utc=0 where id=? and option_max_retry_count>status_retry_count`
		vals = append(vals, id)
	case 2:
		sql = `update request set status=2,start_utc=?,end_utc=0 where id=? and status=1`
		vals = append(vals, t, id)
	case 3:
		sql = `update request set status=3,end_utc=?,duration=?,is_succeed=?,err_msg=? where id=?`
		vals = append(vals, t, duration, isSucceed, msg, id)
	default:
		return false, pbError.ErrInvalidArgument.New("invalid status")
	}
	vals[0] = sql

	if ret, err := session.Context(ctx).Exec(vals...); err != nil {
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
