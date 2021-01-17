package manager

import (
	"context"

	"github.com/voiladev/VoilaCrawl/internal/model/request"
	"github.com/voiladev/go-framework/glog"
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
	return nil, nil
}

func (m *RequestManager) Create(ctx context.Context, session *xorm.Session, req *request.Request) (*request.Request, error) {
	if m == nil {
		return nil, nil
	}
	return nil, nil
}

func (m *RequestManager) UpdateStatus(ctx context.Context, session *xorm.Session, id string, isSucceed bool, msg string) error {
	if m == nil {
		return nil
	}
	return nil
}

func (m *RequestManager) Delete(ctx context.Context, id string) error {
	if m == nil {
		return nil
	}
	return nil
}
