package main

import (
	"fmt"
	"github.com/voiladev/VoilaCrawler/pkg/cli"
	"github.com/voiladev/VoilaCrawler/pkg/crawler"
	"github.com/voiladev/VoilaCrawler/pkg/net/http"
	"github.com/voiladev/go-framework/glog"
)

type (
	New        = func(*http.Client, glog.Log) (crawler.Crawler, error)
	NewWithApp = func(*cli.App, *http.Client, glog.Log) (crawler.Crawler, error)
)

func main() {
	var a interface{} = func(*http.Client, glog.Log) (crawler.Crawler, error) {
		fmt.Println("ok")
		return nil, nil
	}

	if f, ok := a.(New); ok {
		New(f)(nil, nil)
	}

	if f, ok := a.(func(*http.Client, glog.Log) (crawler.Crawler, error)); ok {
		New(f)(nil, nil)
	}
}
