package manager

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/voiladev/VoilaCrawl/internal/model/request"
	"github.com/voiladev/VoilaCrawl/pkg/types"
	pbSearch "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/search"
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

type ListRequest struct {
	Limit  int32
	Cursor string

	StoreId  string
	JobId    string
	StartUtc int64
	EndUtc   int64
	Order    pbSearch.SortOrder
}

type ListResponse struct {
	Cursor string
	Data   []*types.RequestHistory
}

type _Cursor struct {
	Limit         int32              `json:"l"`
	SortField     string             `json:"sf"`
	SortOrder     pbSearch.SortOrder `json:"so"`
	LastTimestamp int64              `json:"lt"`
}

func loadCursor(raw string) (*_Cursor, error) {
	data, err := hex.DecodeString(raw)
	if err != nil {
		return nil, errors.New("invalid cursor")
	}

	var c _Cursor
	if err = json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *_Cursor) String() string {
	if c == nil || c.LastTimestamp == 0 {
		return ""
	}

	data, _ := json.Marshal(c)
	return hex.EncodeToString(data)
}

func (m *HistoryManager) List(ctx context.Context, req ListRequest) (*ListResponse, error) {
	if m == nil {
		return nil, nil
	}
	logger := m.logger.New("List")

	logger.Infof("list logs %+v", &req)

	var (
		err           error
		c             *_Cursor
		lastTimestamp int64
	)
	if req.Cursor != "" {
		if c, err = loadCursor(req.Cursor); err != nil {
			return nil, pbError.ErrInvalidArgument.New(err)
		}
		if c.Limit > 0 && c.LastTimestamp > 0 && c.SortField == "timestamp" {
			req.Limit = c.Limit
			req.Order = c.SortOrder
			lastTimestamp = c.LastTimestamp
		}
	}
	if req.Order == pbSearch.SortOrder_Unknown {
		req.Order = pbSearch.SortOrder_DESC
	}

	handler := m.engine.Context(ctx)
	if req.StoreId != "" {
		handler = handler.And("store_id=?", req.StoreId)
	}
	if req.JobId != "" {
		handler = handler.And("job_id=?", req.JobId)
	}
	if req.StartUtc > 0 {
		handler = handler.And("timestamp>=?", req.StartUtc)
	}
	if req.EndUtc > 0 {
		handler = handler.And("timestamp<=?", req.EndUtc)
	}
	if lastTimestamp > 0 {
		if req.Order == pbSearch.SortOrder_ASC {
			handler = handler.And("timestamp>?", lastTimestamp)
		} else {
			handler = handler.And("timestamp<?", lastTimestamp)
		}
	}

	orderBy := "timestamp desc"
	if req.Order == pbSearch.SortOrder_ASC {
		orderBy = "timestamp asc"
	}
	var items []*types.RequestHistory
	if err := handler.OrderBy(orderBy).Limit(int(req.Limit)).Find(&items); err != nil {
		logger.Error(err)
		return nil, err
	}
	if c == nil {
		c = &_Cursor{}
	}
	if len(items) > 0 {
		c.LastTimestamp = items[len(items)-1].Timestamp
		c.Limit = req.Limit
		c.SortField = "timestamp"
		c.SortOrder = req.Order
	}
	return &ListResponse{Data: items, Cursor: c.String()}, nil
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
