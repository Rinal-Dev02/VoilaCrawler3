package main

import (
	"crypto/md5"
	"errors"
	"fmt"
	"net/http"
	"plugin"
	"strings"

	"github.com/voiladev/VoilaCrawl/pkg/crawler"
	"github.com/voiladev/go-framework/glog"
)

type Crawler struct {
	crawler.Crawler

	gid  string
	path string
}

func NewCrawler(path string, logger glog.Log) (*Crawler, error) {
	p, err := plugin.Open(path)
	if err != nil {
		return nil, err
	}
	funcVal, err := p.Lookup("New")
	if err != nil {
		return nil, err
	}

	newFunc, ok := funcVal.(func(logger glog.Log) (crawler.Crawler, error))
	if !ok {
		return nil, fmt.Errorf("plugin %s %s", path, crawler.ErrNotImplementNewType)
	}

	crawler := Crawler{path: path}
	crawler.Crawler, err = newFunc(logger) // TODO: added more args
	if err != nil {
		return nil, err
	}
	crawler.gid = fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s-%s-%v", hostname, crawler.ID(), crawler.Version()))))

	return &crawler, nil
}

func (e *Crawler) GlobalID() string {
	if e == nil {
		return ""
	}
	return e.gid
}

func (e *Crawler) SetHeader(r *http.Request) *http.Request {
	if e == nil || r == nil {
		return nil
	}

	options := e.CrawlOptions()
	for key := range options.MustHeader {
		r.Header.Set(key, options.MustHeader.Get(key))
	}
	for _, item := range options.MustCookies {
		if item.Path != "" && !strings.HasPrefix(r.URL.Path, item.Path) {
			continue
		}
		v := fmt.Sprintf("%s=%s", item.Name, item.Value)
		if r.Header.Get("Cookie") == "" {
			r.Header.Set("Cookie", v)
		} else {
			r.Header.Set("Cookie", r.Header.Get("Cookie")+"; "+v)
		}
	}
	return r
}

func (e *Crawler) Unmarshal(ret interface{}) error {
	if e == nil {
		return errors.New("nil")
	}

	return nil
}
