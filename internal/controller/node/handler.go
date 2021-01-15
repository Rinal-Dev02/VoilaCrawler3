package node

import (
	"context"
	"io"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/voiladev/VoilaCrawl/internal/model/node"
	pbCrawl "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/protoutil"
	pbError "github.com/voiladev/protobuf/protoc-gen-go/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
)

// nodeHanadler
type nodeHanadler struct {
	ctx context.Context

	id        string
	ctrl      *NodeController
	conn      pbCrawl.Gateway_ChannelServer
	node      *node.Node
	msgBuffer chan protoreflect.ProtoMessage

	logger glog.Log
}

// Node
func NewNodeHandler(ctx context.Context, ctrl *NodeController, conn pbCrawl.Gateway_ChannelServer, logger glog.Log) (*nodeHanadler, error) {
	h := nodeHanadler{
		ctx:       ctx,
		id:        uuid.NewV4().String(),
		ctrl:      ctrl,
		conn:      conn,
		msgBuffer: make(chan protoreflect.ProtoMessage, 10),
		logger:    logger.New("NodeHandler"),
	}

	go func(ctx context.Context, handler *nodeHanadler, logger glog.Log) {
		defer func() {
			close(handler.msgBuffer)
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-handler.msgBuffer:
				if ok {
					return
				}

				{
					turn := 0
					for handler.node == nil {
						logger.Warnf("node not init yet, waiting...")
						time.Sleep(time.Millisecond * time.Duration((300 * (1 + turn/2))))
						turn += 1
					}
				}

				anyData, err := anypb.New(msg)
				if err != nil {
					logger.Errorf("marshal send data failed, error=%s", err)
					continue
				}
				if err = handler.conn.Send(anyData); err != nil {
					logger.Errorf("send data failed, error=%s", err)
				}
			}
		}
	}(ctx, &h, h.logger)

	return &h, nil
}

func (handler *nodeHanadler) ID() string {
	if handler == nil {
		return ""
	}
	return handler.id
}

func (handler *nodeHanadler) MaxConcurrency() int32 {
	if handler == nil || handler.node == nil {
		return 0
	}
	return handler.node.GetMaxConcurrency()
}

func (handler *nodeHanadler) IdleConcurrency() int32 {
	if handler == nil || handler.node == nil {
		return 0
	}
	return handler.node.GetIdleConcurrency()
}

func (handler *nodeHanadler) Send(ctx context.Context, cmd proto.Message) error {
	if handler == nil {
		return nil
	}
	if cmd == nil {
		return pbError.ErrInternal.New("invalid cmd")
	}
	if _, ok := cmd.(*anypb.Any); ok {
		return pbError.ErrInvalidArgument.New("invalid cmd type, cmd should not be Any type")
	}

	return nil
}

func isConnectionClosed(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

var (
	joinPingTypeUrl      = protoutil.GetTypeUrl(&pbCrawl.Join_Ping{})
	heartbetaPingTypeUrl = protoutil.GetTypeUrl(&pbCrawl.Heartbeat_Ping{})
)

func (handler *nodeHanadler) Run() error {
	if handler == nil {
		return nil
	}
	logger := handler.logger.New("Run")

	for {
		if isConnectionClosed(handler.ctx) {
			return handler.ctx.Err()
		}

		anyData, err := handler.conn.Recv()
		if err == io.EOF {
			return nil
		} else if err != nil {
			handler.logger.Errorf("read from connection failed, error=%s", err)
			return err
		}

		now := time.Now()
		switch anyData.GetTypeUrl() {
		case joinPingTypeUrl:
			var (
				err    error
				packet pbCrawl.Join_Ping
			)
			if err = anypb.UnmarshalTo(anyData, &packet, proto.UnmarshalOptions{}); err != nil {
				logger.Errorf("unmarshal Join Ping failed, error=%s", err)
				return pbError.ErrInternal.New(err)
			}

			node := node.New(packet.Node)
			if node, err = handler.ctrl.nodeManager.Save(handler.ctx, node); err != nil {
				logger.Errorf("save node failed, error=%s", err)
				return err
			} else {
				handler.node = node
			}

			// TODO: register crawlers

			delay := now.UnixNano() - packet.Timestamp
			if err = handler.Send(handler.ctx, &pbCrawl.Join_Pong{
				Timestamp:         time.Now().UnixNano(),
				NodeId:            node.GetId(),
				NetworkDelay:      delay,
				HeartbeatInterval: handler.ctrl.options.HeartbeatInternal,
			}); err != nil {
				logger.Errorf("send command failed, error=%s", err)
				return pbError.ErrInternal.New(err)
			}
		case heartbetaPingTypeUrl:
			var (
				err    error
				packet pbCrawl.Heartbeat_Ping
			)
			if err = anypb.UnmarshalTo(anyData, &packet, proto.UnmarshalOptions{}); err != nil {
				logger.Errorf("unmarshal heartbeat Ping failed, error=%s", err)
				return pbError.ErrInternal.New(err)
			}
			handler.node.SetIdleConcurrency(packet.GetIdleConcurrency())
			handler.node.SetMaxConcurrency(packet.GetMaxConcurrency())

			delay := now.UnixNano() - packet.Timestamp
			if err = handler.Send(handler.ctx, &pbCrawl.Heartbeat_Pong{
				Timestamp:    time.Now().UnixNano(),
				NodeId:       packet.GetNodeId(),
				NetworkDelay: delay,
			}); err != nil {
				logger.Errorf("send command failed, error=%s", err)
				return pbError.ErrInternal.New(err)
			}
		default:
			return pbError.ErrUnavailable.New(
				handler.logger.Errorf("unsupported command %s", anyData.GetTypeUrl()).ToError(),
			)
		}
	}
}
