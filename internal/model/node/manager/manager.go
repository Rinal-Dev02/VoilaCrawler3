package manager

import (
	"context"
	"sort"
	"sync"

	"github.com/voiladev/VoilaCrawl/internal/model/node"
	"github.com/voiladev/go-framework/glog"
	pbError "github.com/voiladev/protobuf/protoc-gen-go/errors"
)

// NodeManager
type NodeManager struct {
	nodes  sync.Map
	logger glog.Log
}

func NewNodeManager(logger glog.Log) (*NodeManager, error) {
	m := NodeManager{
		logger: logger.New("NodeManager"),
	}

	return &m, nil
}

func (m *NodeManager) GetById(ctx context.Context, id string) (*node.Node, error) {
	if m == nil {
		return nil, nil
	}

	if val, ok := m.nodes.Load(id); ok {
		return val.(*node.Node), nil
	}
	return nil, nil
}

func (m *NodeManager) List(ctx context.Context) (ret []*node.Node, err error) {
	if m == nil {
		return nil, nil
	}

	m.nodes.Range(func(key, value interface{}) bool {
		ret = append(ret, value.(*node.Node))
		return true
	})

	// sort by online time desc
	sort.SliceStable(ret, func(i, j int) bool {
		return ret[i].GetMetadata().GetOnlineUtc() > ret[j].GetMetadata().OnlineUtc
	})
	return ret, nil
}

func (m *NodeManager) Save(ctx context.Context, node *node.Node) (*node.Node, error) {
	if m == nil {
		return nil, nil
	}

	if err := node.Validate(); err != nil {
		return nil, pbError.ErrInvalidArgument.New(err)
	}

	m.nodes.Store(node.GetId(), node)

	return node, nil
}

func (m *NodeManager) Delete(ctx context.Context, id string) error {
	if m == nil {
		return nil
	}

	m.nodes.Delete(id)
	return nil
}
