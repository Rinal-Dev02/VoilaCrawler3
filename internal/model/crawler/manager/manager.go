package manager

import "github.com/voiladev/go-framework/glog"

type CrawlerManager struct {
	logger glog.Log
}

func NewCrawlerManager(logger glog.Log) (*CrawlerManager, error) {
	m := CrawlerManager{
		logger: logger,
	}
	return &m, nil
}
