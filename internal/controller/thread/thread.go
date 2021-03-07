package thread

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/voiladev/go-framework/glog"
)

type hostConcurrencyStatus struct {
	Host      string
	Count     int32
	Requests  map[string]int64
	Timestamp int64
	Mutex     sync.Mutex
}

// ThreadController
type ThreadController struct {
	ctx           context.Context
	threadPerHost int32
	hostStatus    sync.Map
	logger        glog.Log
}

func NewThreadController(ctx context.Context, theadPerHost int32, logger glog.Log) (*ThreadController, error) {
	if ctx == nil {
		return nil, errors.New("invalid context")
	}
	if theadPerHost < 1 {
		return nil, errors.New("invalid concurrenry")
	}
	if logger == nil {
		return nil, errors.New("invalid logger")
	}

	ctrl := ThreadController{
		ctx:           ctx,
		threadPerHost: theadPerHost,
		logger:        logger.New("ThreadController"),
	}

	go func() {
		var (
			ticker      = time.NewTicker(time.Second)
			invalidReqs []string
		)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				invalidReqs = invalidReqs[0:0]
				ctrl.hostStatus.Range(func(key, val interface{}) bool {
					status := val.(*hostConcurrencyStatus)
					status.Mutex.Lock()
					defer status.Mutex.Unlock()

					t := time.Now().Unix()
					for k, deadline := range status.Requests {
						if deadline <= t {
							invalidReqs = append(invalidReqs, k)
						}
					}
					for _, id := range invalidReqs {
						delete(status.Requests, id)
						atomic.AddInt32(&status.Count, -1)
					}
					ctrl.logger.Debugf("thread %s %d", key, status.Count)
					return true
				})
			}
		}
	}()
	return &ctrl, nil
}

func (ctrl *ThreadController) Lock(ctx context.Context, host string, reqId string, ttl int64) bool {
	if ctrl == nil {
		return false
	}
	if ttl == 0 {
		ttl = 6 * 60
	}

	val, _ := ctrl.hostStatus.LoadOrStore(host, &hostConcurrencyStatus{
		Host:      host,
		Count:     0,
		Requests:  map[string]int64{},
		Timestamp: time.Now().Unix(),
	})
	status := val.(*hostConcurrencyStatus)

	status.Mutex.Lock()
	if status.Count < ctrl.threadPerHost {
		ctrl.logger.Debugf("set lock %s %s %v", host, reqId, ttl)

		atomic.AddInt32(&status.Count, 1)
		status.Requests[reqId] = time.Now().Unix() + ttl
		status.Mutex.Unlock()
		return true
	}
	status.Mutex.Unlock()
	return false
}

func (ctrl *ThreadController) Unlock(ctx context.Context, host string, reqId string) {
	if ctrl == nil {
		return
	}

	ctrl.logger.Debugf("unlock %s, %s", host, reqId)

	if val, ok := ctrl.hostStatus.Load(host); ok {
		status := val.(*hostConcurrencyStatus)

		status.Mutex.Lock()
		if _, ok := status.Requests[reqId]; ok {
			delete(status.Requests, reqId)
			atomic.AddInt32(&status.Count, -1)
		}
		status.Mutex.Unlock()
	}
}
