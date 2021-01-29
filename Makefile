.PHONY: all api crawlet

all: api crawlet plugins

api:
	go vet ./cmd/crawl-api
	go build -ldflags "-X main.buildTime=`date +%Y%m%d.%H:%M:%S` -X main.buildCommit=`git rev-parse --short=12 HEAD` -X main.buildBranch=`git branch --show-current`" -o ./releases/crawl-api ./cmd/crawl-api

crawlet:
	go vet ./cmd/crawlet
	go build -ldflags "-X main.buildTime=`date +%Y%m%d.%H:%M:%S` -X main.buildCommit=`git rev-parse --short=12 HEAD` -X main.buildBranch=`git branch --show-current`" -o ./releases/crawlet ./cmd/crawlet

plugins:
	make -C ./spiders
