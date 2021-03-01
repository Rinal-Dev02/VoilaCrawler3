package main

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/voiladev/go-framework/glog"
)

type GPool struct {
	ctx                context.Context
	maxConcurrency     int32
	currentConcurrency int32
	jobQueue           chan func()
	logger             glog.Log
}

func NewGPool(ctx context.Context, cap int32, logger glog.Log) (*GPool, error) {
	p := GPool{
		ctx:            ctx,
		maxConcurrency: cap,
		jobQueue:       make(chan func(), 10),
		logger:         logger.New("GPool"),
	}

	go func() {
		for {
			if p.CurrentConcurrency() >= p.MaxConcurrency() {
				time.Sleep(300 * time.Millisecond)
				continue
			}

			select {
			case <-p.ctx.Done():
				return
			case jobFunc, ok := <-p.jobQueue:
				if !ok {
					return
				}

				go func() {
					atomic.AddInt32(&p.currentConcurrency, 1)
					defer func() {
						if r := recover(); r != nil {
							p.logger.Error(r)
						}
						atomic.AddInt32(&p.currentConcurrency, -1)
					}()
					jobFunc()
				}()
			}
		}
	}()

	return &p, nil
}

func (p *GPool) MaxConcurrency() int32 {
	if p == nil {
		return 0
	}
	return atomic.LoadInt32(&p.maxConcurrency)
}

func (p *GPool) CurrentConcurrency() int32 {
	if p == nil {
		return 0
	}
	return atomic.LoadInt32(&p.currentConcurrency)
}

func (p *GPool) DoJob(jobFunc func()) error {
	if p == nil || jobFunc == nil {
		return nil
	}

	select {
	case <-p.ctx.Done():
		return p.ctx.Err()
	case p.jobQueue <- jobFunc:
	}
	return nil
}
