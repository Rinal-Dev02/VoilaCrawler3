package thread

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/voiladev/VoilaCrawl/internal/pkg/config"
	"github.com/voiladev/go-framework/glog"
)

type hostConcurrencyStatus struct {
	Host                 string
	Count                int32
	MaxCount             int32
	Requests             map[string]int64
	ContinusSucceesCount int32
	ContinusErrorCount   int32
	Timestamp            int64
	Mutex                sync.Mutex
}

// ThreadController
type ThreadController struct {
	ctx           context.Context
	threadPerHost int32
	hostStatus    sync.Map
	logger        glog.Log
}

func NewThreadController(ctx context.Context, threadPerHost int32, logger glog.Log) (*ThreadController, error) {
	if ctx == nil {
		return nil, errors.New("invalid context")
	}
	if threadPerHost < 1 || threadPerHost > 5 {
		return nil, errors.New("invalid concurrenry")
	}
	if logger == nil {
		return nil, errors.New("invalid logger")
	}

	ctrl := ThreadController{
		ctx:           ctx,
		threadPerHost: threadPerHost,
		logger:        logger.New("ThreadController"),
	}

	go func() {
		const speedCheckInterval = 5 * 60 // 5mins
		var (
			ticker          = time.NewTicker(time.Second * 3)
			speedCheckTimer = time.NewTicker(time.Second * speedCheckInterval)
			invalidReqs     []string
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
						if status.Count > 0 {
							atomic.AddInt32(&status.Count, -1)
						}
					}
					ctrl.logger.Debugf("thread %s max: %d, current: %d", key, status.MaxCount, status.Count)

					return true
				})
			case <-speedCheckTimer.C:
				ctrl.hostStatus.Range(func(key, val interface{}) bool {
					status := val.(*hostConcurrencyStatus)
					status.Mutex.Lock()
					defer status.Mutex.Unlock()

					total := status.ContinusSucceesCount + status.ContinusErrorCount
					succeedRate := float64(status.ContinusSucceesCount) / float64(total)
					status.ContinusErrorCount = 0
					status.ContinusSucceesCount = 0
					if time.Now().Unix()-status.Timestamp < speedCheckInterval {
						return true
					}

					if total <= 2*ctrl.threadPerHost {
						status.MaxCount = ctrl.threadPerHost
						return true
					}
					switch {
					case succeedRate > 0.9:
						if status.MaxCount < 20 {
							status.MaxCount += 1
						}
					case succeedRate == 0:
						// down to only one thread
						for i := 0; i < 2 && status.MaxCount > 1; i++ {
							status.MaxCount -= 1
						}
					case succeedRate < 0.4:
						// down to default thread count
						if status.MaxCount > ctrl.threadPerHost {
							status.MaxCount -= 1
						}
					}
					ctrl.logger.Debugf("thread %s max: %d, current: %d", key, status.MaxCount, status.Count)
					return true
				})
				speedCheckTimer.Reset(time.Second * speedCheckInterval)
			}
		}
	}()
	return &ctrl, nil
}

func (ctrl *ThreadController) TryLock(host string) bool {
	if ctrl == nil {
		return false
	}

	val, _ := ctrl.hostStatus.LoadOrStore(host, &hostConcurrencyStatus{
		Host:      host,
		Count:     0,
		MaxCount:  ctrl.threadPerHost,
		Requests:  map[string]int64{},
		Timestamp: time.Now().Unix(),
	})
	status := val.(*hostConcurrencyStatus)

	status.Mutex.Lock()
	flag := status.Count < status.MaxCount
	status.Mutex.Unlock()

	return flag
}

func (ctrl *ThreadController) Lock(host string, reqId string, ttl int32) bool {
	if ctrl == nil {
		return false
	}
	if ttl <= 0 {
		ttl = config.DefaultTtlPerRequest
	}

	val, _ := ctrl.hostStatus.LoadOrStore(host, &hostConcurrencyStatus{
		Host:      host,
		Count:     0,
		MaxCount:  ctrl.threadPerHost,
		Requests:  map[string]int64{},
		Timestamp: time.Now().Unix(),
	})
	status := val.(*hostConcurrencyStatus)

	status.Mutex.Lock()
	if status.Count < status.MaxCount {
		ctrl.logger.Debugf("set lock %s %s", host, reqId)

		atomic.AddInt32(&status.Count, 1)
		status.Requests[reqId] = time.Now().Unix() + int64(ttl)
		status.Mutex.Unlock()
		return true
	}
	status.Mutex.Unlock()
	return false
}

func (ctrl *ThreadController) Unlock(host string, reqId string, ret error) {
	if ctrl == nil {
		return
	}

	ctrl.logger.Debugf("unlock %s, %s", host, reqId)

	if val, ok := ctrl.hostStatus.Load(host); ok {
		status := val.(*hostConcurrencyStatus)

		status.Mutex.Lock()
		if _, ok := status.Requests[reqId]; ok {
			delete(status.Requests, reqId)
			if status.Count > 0 {
				atomic.AddInt32(&status.Count, -1)
			}
			if ret == nil {
				status.ContinusSucceesCount += 1
			} else {
				status.ContinusErrorCount += 1
			}
		}
		status.Mutex.Unlock()
	}
}
