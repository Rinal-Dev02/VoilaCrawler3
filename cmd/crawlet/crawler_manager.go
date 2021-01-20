package main

import (
	"context"

	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/types/vermap"
)

// CrawlerManager
type CrawlerManager struct {
	crawlers *vermap.VersionMap
	logger   glog.Log
}

func NewCrawlerManager(logger glog.Log) (*CrawlerManager, error) {
	m := CrawlerManager{
		crawlers: &vermap.VersionMap{},
		logger:   logger.New("CrawlerManager"),
	}
	return &m, nil
}

// GetByGID
func (m *CrawlerManager) GetByGID(ctx context.Context, gid string) (*Crawler, error) {
	if m == nil {
		return nil, nil
	}

	var ret *Crawler
	m.crawlers.Range(func(key string, vals []interface{}) bool {
		for _, val := range vals {
			crawler := val.(*Crawler)
			if crawler.GlobalID() == gid {
				ret = crawler
				return false
			}
		}
		return true
	})
	return ret, nil
}

// GetByID
func (m *CrawlerManager) GetByID(ctx context.Context, id string) ([]*Crawler, error) {
	if m == nil {
		return nil, nil
	}

	var ret []*Crawler
	m.crawlers.Range(func(key string, vals []interface{}) bool {
		for _, val := range vals {
			crawler := val.(*Crawler)
			ret = append(ret, crawler)
		}
		return true
	})
	return ret, nil
}

// GetByDomain
func (m *CrawlerManager) GetByHost(ctx context.Context, host string) ([]*Crawler, error) {
	if m == nil {
		return nil, nil
	}

	var ret []*Crawler
	m.crawlers.Range(func(key string, vals []interface{}) bool {
		m.logger.Debugf("%s count=%d", key, len(vals))
		for _, val := range vals {
			crawler := val.(*Crawler)
			for _, d := range crawler.AllowedDomains() {
				m.logger.Debugf("%s=%s", host, d)
				if d == host {
					ret = append(ret, crawler)
				}
			}
		}
		return true
	})
	return ret, nil
}

// Count
func (m *CrawlerManager) Count(ctx context.Context) (int32, error) {
	if m == nil {
		return 0, nil
	}

	var count int32
	m.crawlers.Range(func(key string, vals []interface{}) bool {
		count += 1
		return true
	})
	return count, nil
}

// List
func (m *CrawlerManager) List(ctx context.Context) ([]*Crawler, error) {
	if m == nil {
		return nil, nil
	}

	var ret []*Crawler
	m.crawlers.Range(func(key string, vals []interface{}) bool {
		for _, val := range vals {
			crawler := val.(*Crawler)
			ret = append(ret, crawler)
		}
		return true
	})
	return ret, nil
}

// Save
func (m *CrawlerManager) Save(ctx context.Context, crawler *Crawler) error {
	if m == nil {
		return nil
	}
	m.crawlers.Set(crawler.ID(), crawler, int(crawler.Version()))
	return nil
}
