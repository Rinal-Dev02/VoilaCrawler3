package v1

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	crawlerCtrl "github.com/voiladev/VoilaCrawl/internal/controller/crawler"
	reqCtrl "github.com/voiladev/VoilaCrawl/internal/controller/request"
	"github.com/voiladev/VoilaCrawl/internal/model/crawler"
	crawlerManager "github.com/voiladev/VoilaCrawl/internal/model/crawler/manager"
	"github.com/voiladev/VoilaCrawl/internal/model/request"
	reqManager "github.com/voiladev/VoilaCrawl/internal/model/request/manager"
	pbCrawl "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/protoutil"
	"github.com/voiladev/go-framework/redis"
	pbError "github.com/voiladev/protobuf/protoc-gen-go/errors"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/proto"
	"xorm.io/xorm"
)

type GatewayServer struct {
	pbCrawl.UnimplementedGatewayServer

	ctx            context.Context
	crawlerCtrl    *crawlerCtrl.CrawlerController
	crawlerManager *crawlerManager.CrawlerManager
	logger         glog.Log
}

func NewGatewayServer(
	ctx context.Context,
	crawlerCtrl *crawlerCtrl.CrawlerController,
	crawlerManager *crawlerManager.CrawlerManager,
	logger glog.Log,
) (pbCrawl.GatewayServer, error) {
	if crawlerCtrl == nil {
		return nil, errors.New("invalid crawler controller")
	}
	if crawlerManager == nil {
		return nil, errors.New("invalid crawler manager")
	}
	if logger == nil {
		return nil, errors.New("invalid logger")
	}

	s := GatewayServer{
		ctx:            ctx,
		crawlerCtrl:    crawlerCtrl,
		crawlerManager: crawlerManager,
		logger:         logger.New("GatewayServer"),
	}
	return &s, nil
}

func (s *GatewayServer) Connect(srv pbCrawl.Gateway_ConnectServer) (err error) {
	if s == nil {
		return nil
	}
	logger := s.logger.New("Connect")

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%s", e)
		}
		// TODO: remove crawler
	}()

	var (
		ip           string
		isRegistered bool
		ctx          = srv.Context()
	)
	if peer, _ := peer.FromContext(srv.Context()); peer != nil {
		ip, _, _ = net.SplitHostPort(peer.Addr.String())
	} else {
		return fmt.Errorf("get peer info failed")
	}

	for {
		anyData, err := srv.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		switch anyData.GetTypeUrl() {
		case protoutil.GetTypeUrl(&pbCrawl.ConnectRequest_Ping{}):
			if isRegistered {
				logger.Errorf("dumplicate register request")
				continue
			}

			var data pbCrawl.ConnectRequest_Ping
			if err := proto.Unmarshal(anyData.GetValue(), &data); err != nil {
				logger.Error(err)
				return err
			}

			cw, err := crawler.NewCrawler(&data)
			if err != nil {
				logger.Error(err)
				return err
			}
			cw.ServeIP = ip

			s.crawlerManager.Delete()

			isRegistered = true
		case protoutil.GetTypeUrl(&pbCrawl.ConnectRequest_Heartbeat{}):
			var data pbCrawl.ConnectRequest_Heartbeat
			if err := proto.Unmarshal(anyData.GetValue(), &data); err != nil {
				logger.Error(err)
				return err
			}
		default:
			logger.Errorf("unsupported type %s", anyData.GetTypeUrl())
			continue
		}
	}
	return nil
}
