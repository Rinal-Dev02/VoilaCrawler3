package v1

import (
	"context"
	"errors"

	nodeCtrl "github.com/voiladev/VoilaCrawl/internal/controller/node"
	"github.com/voiladev/VoilaCrawl/internal/model/request"
	reqManager "github.com/voiladev/VoilaCrawl/internal/model/request/manager"
	pbCrawl "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl"
	"github.com/voiladev/go-framework/glog"
	pbError "github.com/voiladev/protobuf/protoc-gen-go/errors"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GatewayServer struct {
	pbCrawl.UnimplementedGatewayServer

	ctx            context.Context
	nodeCtrl       *nodeCtrl.NodeController
	requestManager *reqManager.RequestManager
	logger         glog.Log
}

func NewGatewayServer(
	ctx context.Context,
	nodeCtrl *nodeCtrl.NodeController,
	requestManager *reqManager.RequestManager,
	logger glog.Log,
) (pbCrawl.GatewayServer, error) {
	if nodeCtrl == nil {
		return nil, errors.New("invalid node controller")
	}
	if requestManager == nil {
		return nil, errors.New("invalid request manager")
	}
	s := GatewayServer{
		ctx:            ctx,
		nodeCtrl:       nodeCtrl,
		requestManager: requestManager,
		logger:         logger.New("GatewayServer"),
	}
	return &s, nil
}

func (s *GatewayServer) Channel(cs pbCrawl.Gateway_ChannelServer) error {
	if s == nil {
		return nil
	}
	logger := s.logger.New("Channel")

	handler, err := s.nodeCtrl.Register(cs.Context(), cs)
	if err != nil {
		logger.Error(err)
		return err
	}

	defer func() {
		s.nodeCtrl.Unregister(cs.Context(), handler.ID())
	}()

	if err := handler.Run(); err != nil {
		logger.Error(err)
		return err
	}
	return nil
}

func (s *GatewayServer) Fetch(ctx context.Context, req *pbCrawl.FetchRequest) (*emptypb.Empty, error) {
	if s == nil {
		return nil, nil
	}
	logger := s.logger.New("Fetch")

	r, err := request.NewRequest(req)
	if err != nil {
		logger.Errorf("load request failed, error=%s", err)
		return nil, pbError.ErrInvalidArgument.New(err)
	}
	if r, err = s.requestManager.Create(ctx, nil, r); err != nil {
		logger.Errorf("save request failed, error=%s", err)
		return nil, err
	}
	if err = s.nodeCtrl.PublishRequest(ctx, r); err != nil {
		logger.Error(err)
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
