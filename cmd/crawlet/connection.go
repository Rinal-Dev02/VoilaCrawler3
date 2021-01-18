package main

import (
	"context"
	"io"
	"sync/atomic"
	"time"

	pbCrawl "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/protoutil"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
)

type Connection struct {
	ctx           context.Context
	conn          *grpc.ClientConn
	gatewayClient pbCrawl.GatewayClient
	channelClient pbCrawl.Gateway_ChannelClient
	msgBuffer     chan *anypb.Any
	logger        glog.Log
}

func NewConnection(ctx context.Context, addr string, logger glog.Log) (*Connection, error) {
	c := Connection{
		ctx:       ctx,
		msgBuffer: make(chan *anypb.Any, 10),
		logger:    logger.New("Connection"),
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	var err error
	c.conn, err = grpc.DialContext(timeoutCtx, addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	c.gatewayClient = pbCrawl.NewGatewayClient(c.conn)

	return &c, nil
}

func (conn *Connection) NewChannelHandler(ctx context.Context, ctrl *CrawlerController) (*ChannelHandler, error) {
	handler := ChannelHandler{
		conn:            conn,
		ctrl:            ctrl,
		heartbeatTicker: time.NewTicker(time.Hour),
		logger:          conn.logger.New("ChannelHandler"),
	}

	var err error
	handler.client, err = conn.gatewayClient.Channel(ctx)
	if err != nil {
		conn.logger.Error("connect channel failed, error=%s", err)
		return nil, err
	}
	return &handler, nil
}

type ChannelHandler struct {
	conn *Connection
	ctrl *CrawlerController

	client            pbCrawl.Gateway_ChannelClient
	heartbeatTicker   *time.Ticker
	heartbeatInterval int64
	isRegistered      bool
	logger            glog.Log
}

func (handler *ChannelHandler) Send(ctx context.Context, msg protoreflect.ProtoMessage) error {
	if handler == nil || msg == nil {
		return nil
	}

	anydata, err := anypb.New(msg)
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case handler.conn.msgBuffer <- anydata:
	}
	return nil
}

func (handler *ChannelHandler) Watch(ctx context.Context, callback func(context.Context, *pbCrawl.Command_Request)) error {
	if handler == nil || callback == nil {
		return nil
	}
	logger := handler.logger.New("Watch")

	nctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		err := func() error {
			for {
				anydata, err := handler.client.Recv()
				if err == io.EOF {
					return nil
				} else if err != nil {
					logger.Errorf("receive err failed, error=%s", err)
					return err
				}

				switch anydata.GetTypeUrl() {
				case protoutil.GetTypeUrl(&pbCrawl.Join_Pong{}):
					var packet pbCrawl.Join_Pong
					if err = anypb.UnmarshalTo(anydata, &packet, proto.UnmarshalOptions{}); err != nil {
						logger.Errorf("unmarshal pong message failed, error=%s", err)
						return err
					}
					if packet.GetHeartbeatInterval() > 0 {
						atomic.StoreInt64(&handler.heartbeatInterval, packet.GetHeartbeatInterval())
						handler.isRegistered = true
						handler.heartbeatTicker.Reset(time.Duration(packet.GetHeartbeatInterval()) * time.Millisecond)
					}
					logger.Infof("network delay %v", packet.NetworkDelay)
				case protoutil.GetTypeUrl(&pbCrawl.Heartbeat_Pong{}):
					var packet pbCrawl.Heartbeat_Pong
					if err = anypb.UnmarshalTo(anydata, &packet, proto.UnmarshalOptions{}); err != nil {
						logger.Errorf("unmarshal heartbeat pong message failed, error=%s", err)
						return err
					}
					logger.Infof("network delay %v", packet.NetworkDelay)
				case protoutil.GetTypeUrl(&pbCrawl.Command{}):
					var packet pbCrawl.Command
					if err = anypb.UnmarshalTo(anydata, &packet, proto.UnmarshalOptions{}); err != nil {
						logger.Errorf("unmarshal command message failed, error=%s", err)
						return err
					}

					switch packet.GetData().GetTypeUrl() {
					case protoutil.GetTypeUrl(&pbCrawl.Command_Request{}):
						var req pbCrawl.Command_Request
						if err = anypb.UnmarshalTo(packet.GetData(), &req, proto.UnmarshalOptions{}); err != nil {
							logger.Errorf("unmarshal request command message failed, error=%s", err)
							return err
						}
						go callback(nctx, &req)
					default:
						handler.logger.Errorf("unsupported cmd data type %s", packet.GetData().GetTypeUrl())
					}
				default:
					handler.logger.Errorf("unsupported cmd type %s", anydata.GetTypeUrl())
				}
			}
		}()
		if err != nil {
			cancel()
		}
	}()

	for {
		select {
		case <-nctx.Done():
			return nctx.Err()
		case <-handler.heartbeatTicker.C:
			msg := pbCrawl.Heartbeat_Ping{
				Timestamp:       time.Now().UnixNano(),
				NodeId:          NodeId(),
				MaxConcurrency:  handler.ctrl.gpool.MaxConcurrency(),
				IdleConcurrency: handler.ctrl.gpool.MaxConcurrency() - handler.ctrl.gpool.CurrentConcurrency(),
			}
			anydata, _ := anypb.New(&msg)
			if err := handler.client.Send(anydata); err != nil {
				handler.logger.Errorf("send heartbeat failed, error=%s", err)
				return err
			}
			handler.logger.Debugf("send heartbeta max: %d, idle: %d", msg.GetMaxConcurrency(), msg.GetIdleConcurrency())
		case msg, ok := <-handler.conn.msgBuffer:
			if !ok {
				return nil
			}
			if err := handler.client.Send(msg); err != nil {
				handler.logger.Errorf("send msg failed, error=%s", err)
				return err
			}
		}
	}
}
