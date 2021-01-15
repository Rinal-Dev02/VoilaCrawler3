package node

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	pbCrawl "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl"
	pbError "github.com/voiladev/protobuf/protoc-gen-go/errors"
)

type Node struct {
	pbCrawl.Node
	mutex sync.RWMutex
}

// New
func New(i *pbCrawl.Join_Ping_Node) *Node {
	n := Node{}
	n.Id = i.GetId()
	n.Host = i.GetHost()
	n.MaxConcurrency = i.GetMaxConcurrency()
	n.IdleConcurrency = i.GetIdleConcurrency()
	n.Status = pbCrawl.NodeStatus_Online
	n.Metadata = &pbCrawl.Node_Metadata{
		OnlineUtc: time.Now().UnixNano(),
	}
	return &n
}

func (n *Node) GetId() string {
	if n == nil {
		return ""
	}
	n.mutex.Lock()
	defer n.mutex.RUnlock()

	return n.Node.GetId()
}

func (n *Node) GetHost() string {
	if n == nil {
		return ""
	}
	n.mutex.Lock()
	defer n.mutex.RUnlock()

	return n.Node.GetHost()
}

func (n *Node) SetStatus(status pbCrawl.NodeStatus) error {
	if n == nil {
		return errors.New("nil node")
	}
	if _, ok := pbCrawl.NodeStatus_name[int32(status)]; !ok || status == pbCrawl.NodeStatus_NodeStatusUnknown {
		return pbError.ErrInvalidArgument.New("invalid status")
	}
	n.mutex.Lock()
	defer n.mutex.Unlock()

	n.Status = status

	return nil
}

func (n *Node) SetMaxConcurrency(val int32) error {
	if n == nil {
		return errors.New("nil node")
	}

	n.mutex.Lock()
	defer n.mutex.Unlock()

	if val < 0 {
		return pbError.ErrInvalidArgument.New(fmt.Sprintf("invalid val"))
	}
	atomic.StoreInt32(&n.MaxConcurrency, val)
	return nil
}

func (n *Node) SetIdleConcurrency(val int32) error {
	if n == nil {
		return errors.New("nil node")
	}

	n.mutex.Lock()
	defer n.mutex.Unlock()

	if val < 0 {
		return pbError.ErrInvalidArgument.New(fmt.Sprintf("invalid val"))
	}
	atomic.StoreInt32(&n.IdleConcurrency, val)
	return nil
}

func (n *Node) IncrIdleConcurrency(delta int32) (int32, error) {
	if n == nil {
		return 0, errors.New("nil node")
	}

	n.mutex.Lock()
	defer n.mutex.Unlock()

	if n.Node.GetIdleConcurrency()+delta < 0 {
		return 0, pbError.ErrInvalidArgument.New(fmt.Sprintf("invalid delta %v", delta))
	}
	return atomic.AddInt32(&n.IdleConcurrency, delta), nil
}

func (n *Node) SetHeartbetaUtc(t int64) error {
	if n == nil {
		return errors.New("nil node")
	}

	n.mutex.Lock()
	n.mutex.Unlock()

	n.Metadata.LastHeartbeatUtc = time.Now().UnixNano()

	return nil
}

func (n *Node) Validate() error {
	if n == nil {
		return errors.New("nil node")
	}

	n.mutex.RLock()
	defer n.mutex.RUnlock()

	if n.Id == "" {
		return errors.New("invalid node uuid")
	}
	if n.Host == "" {
		return errors.New("invalid node host")
	}
	if n.MaxConcurrency <= 0 {
		return errors.New("invalid max concurrency")
	}
	if n.IdleConcurrency < 0 {
		return errors.New("invalid idle concurrency")
	}
	if _, ok := pbCrawl.NodeStatus_name[int32(n.Status)]; !ok || n.Status == pbCrawl.NodeStatus_NodeStatusUnknown {
		return errors.New("invalid status")
	}
	return nil
}

func (n *Node) Unmarshal(ret interface{}) error {
	if n == nil {
		return errors.New("nil Node")
	}
	if ret == nil {
		return nil
	}

	return nil
}
