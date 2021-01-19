package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/urfave/cli/v2"
	svcGateway "github.com/voiladev/VoilaCrawl/internal/api/gateway/v1"
	crawlerCtrl "github.com/voiladev/VoilaCrawl/internal/controller/crawler"
	nodeCtrl "github.com/voiladev/VoilaCrawl/internal/controller/node"
	reqCtrl "github.com/voiladev/VoilaCrawl/internal/controller/request"
	crawlerManager "github.com/voiladev/VoilaCrawl/internal/model/crawler/manager"
	nodeManager "github.com/voiladev/VoilaCrawl/internal/model/node/manager"
	reqManager "github.com/voiladev/VoilaCrawl/internal/model/request/manager"
	"github.com/voiladev/VoilaCrawl/pkg/types"
	pbCrawl "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/grpcutil"
	"github.com/voiladev/go-framework/invocation"
	"github.com/voiladev/go-framework/mysql"
	pbDesc "github.com/voiladev/protobuf/protoc-gen-go/protobuf"
	"go.uber.org/fx"
	grpc "google.golang.org/grpc"
)

var _ServiceDescs = map[string]*pbDesc.ServiceDesc{}

func init() {
	for _, desc := range pbCrawl.ServiceDescs {
		_ServiceDescs[desc.GetFullname()] = desc
	}
}

// App
type App struct {
	ctx        context.Context
	exitChan   <-chan os.Signal
	closeQueue []io.Closer
}

func NewApp(ctx context.Context, exitChan <-chan os.Signal) *App {
	return &App{ctx: ctx, exitChan: exitChan}
}

func (app *App) Context() context.Context {
	return app.ctx
}

func (app *App) Run(args []string) {
	var cliApp = cli.NewApp()
	cliApp.Name = "crawl-api"
	cliApp.Usage = "crawl api server"
	cliApp.Version = Version
	cliApp.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "host",
			Value: "0.0.0.0",
			Usage: "The bind host",
		},
		&cli.IntFlag{
			Name:  "port",
			Usage: "The bind (grpc) port",
			Value: 6000,
		},
		&cli.IntFlag{
			Name:  "http-port",
			Usage: "The bind (http) port",
			Value: 8080,
		},
		&cli.StringFlag{
			Name:  "mysql-dsn",
			Usage: "mysql data source name",
			Value: "root:china123@tcp(voiladev.com:3306)/voila_crawl?charset=utf8mb4&parseTime=True",
		},
		&cli.StringFlag{
			Name:  "nsqd-tcp-addr",
			Usage: "nsqd tcp address",
			Value: "voiladev.com:4150",
		},
		&cli.StringSliceFlag{
			Name:  "nsqlookupd-http-addr",
			Usage: "nsqlookupd http address",
			Value: cli.NewStringSlice("voiladev.com:4161"),
		},
		&cli.BoolFlag{
			Name:  "disable-access-control",
			Usage: "Disable access control",
		},
		&cli.BoolFlag{
			Name:    "debug",
			Usage:   "Enable debug",
			EnvVars: []string{"DEBUG"},
		},
	}
	cliApp.Action = func(c *cli.Context) error {
		logger := glog.New(glog.LogLevelInfo)
		if c.Bool("debug") {
			logger.SetLevel(glog.LogLevelDebug)
			os.Setenv("DEBUG", "1")
		}

		options := []fx.Option{
			fx.Provide(app.Context),
			fx.Provide(func() glog.Log { return logger }),
			fx.Logger(logger),
		}

		if opts, err := app.loadBackends(c); err != nil {
			return cli.NewExitError(err, 1)
		} else {
			options = append(options, opts...)
		}

		options = append(options,
			// grpc server
			fx.Provide(app.newGrpcServer(c)),
			// grpc client
			fx.Provide(app.newGrpcClient(c)),
			// http server
			fx.Provide(app.newHttpServer(c)),

			// Managers
			fx.Provide(crawlerManager.NewCrawlerManager),
			fx.Provide(nodeManager.NewNodeManager),
			fx.Provide(reqManager.NewRequestManager),

			// Controller
			fx.Provide(crawlerCtrl.NewCrawlerController),
			fx.Provide(func() *nodeCtrl.NodeControllerOptions {
				return &nodeCtrl.NodeControllerOptions{
					HeartbeatInternal: 1000,
					NsqdAddr:          c.String("nsqd-tcp-addr"),
				}
			}),
			fx.Provide(nodeCtrl.NewNodeController),
			fx.Provide(func() *reqCtrl.RequestControllerOptions {
				return &reqCtrl.RequestControllerOptions{
					NsqLookupdAddresses: c.StringSlice("nsqlookupd-http-addr"),
				}
			}),
			fx.Provide(func(ctrl *nodeCtrl.NodeController) reqCtrl.Sender {
				return ctrl
			}),

			// Register services
			fx.Provide(svcGateway.NewGatewayServer),

			// Register grpc handler
			fx.Invoke(pbCrawl.RegisterGatewayServer),
			// fx.Invoke(pbCrawl.RegisterCrawlerManagerServer),
			// fx.Invoke(pbCrawl.RegisterNodeManagerServer),

			// Register http handler
			fx.Invoke(pbCrawl.RegisterGatewayHandler),
			fx.Invoke(reqCtrl.NewRequestController),
		)

		depInj := fx.New(options...)
		if err := depInj.Start(app.ctx); err != nil {
			return cli.NewExitError(err, 1)
		}

		<-app.exitChan
		depInj.Stop(app.ctx)
		return nil
	}
	cliApp.Run(args)
}

func (app *App) loadBackends(c *cli.Context) (opts []fx.Option, err error) {
	if app == nil {
		return nil, nil
	}

	if ins, err := mysql.NewMysqlInstaller(
		mysql.WithMysqlInstallerDebugOption(c.Bool("debug")),
		mysql.WithMysqlInstallerDSNOption(c.String("mysql-dsn")),
		mysql.WithMysqlInstallerMaxIdleConnsOption(30),
		mysql.WithMysqlInstallerMaxOpenConnsOption(100),
		mysql.WithMysqlInstallerTablesOption(
			&types.Request{},
		),
	); err != nil {
		return nil, err
	} else {
		opts = append(opts, fx.Provide(ins.Instance))
	}
	return
}

func (app *App) newGrpcServer(c *cli.Context) func(fx.Lifecycle, glog.Log) (grpc.ServiceRegistrar, error) {
	addr := fmt.Sprintf("%s:%d", c.String("host"), c.Int("port"))

	return func(lc fx.Lifecycle, logger glog.Log) (grpc.ServiceRegistrar, error) {
		var interceptor *grpcutil.ServerInterceptor
		if c.Bool("disable-access-control") {
			interceptor = grpcutil.NewServerInterceptor(
				_ServiceDescs, []invocation.NewOption{
					invocation.NewWithAuthController(invocation.NewAdminInsecureAuthController()),
				})
		} else {
			interceptor = grpcutil.NewServerInterceptor(
				_ServiceDescs, []invocation.NewOption{
					invocation.NewWithAuthController(invocation.NewOpenapiProjectionAuthController()),
				})
		}
		server := grpc.NewServer(grpc.UnaryInterceptor(interceptor.UnaryInterceptor), grpc.StreamInterceptor(interceptor.StreamInterceptor))

		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				logger.Infof("grpc listen on %s", addr)
				listener, err := net.Listen("tcp", addr)
				if err != nil {
					return err
				}

				go server.Serve(listener)
				return nil
			},
			OnStop: func(ctx context.Context) error {
				server.GracefulStop()
				return nil
			},
		})
		return grpc.ServiceRegistrar(server), nil
	}
}

func (app *App) newGrpcClient(c *cli.Context) func(logger glog.Log) (conn *grpc.ClientConn, err error) {
	grpcAddr := fmt.Sprintf("%s:%d", "127.0.0.1", c.Int("port"))

	return func(logger glog.Log) (conn *grpc.ClientConn, err error) {
		globalTimer := time.NewTimer(time.Second * 10)
		timer := time.NewTimer(time.Millisecond * 100)
		defer timer.Stop()
		defer globalTimer.Stop()

		for {
			select {
			case <-timer.C:
				conn, err = grpc.Dial(grpcAddr, grpc.WithInsecure(), grpc.WithBackoffMaxDelay(time.Second))
				if err != nil {
					logger.Errorf("connect to grpc server failed, try...")
					timer.Reset(time.Millisecond * 100)
				} else {
					return conn, nil
				}
			case <-globalTimer.C:
				logger.Errorf("connect to grpc server timeout")
				return nil, errors.New("timeout to connect to grpc server")
			}
		}
	}
}

func (app *App) newHttpServer(c *cli.Context) func(fx.Lifecycle, *grpc.ClientConn, glog.Log) (*runtime.ServeMux, error) {
	addr := fmt.Sprintf("%s:%d", c.String("host"), c.Int("http-port"))

	return func(lc fx.Lifecycle, conn *grpc.ClientConn, logger glog.Log) (mux *runtime.ServeMux, err error) {
		mux = runtime.NewServeMux(runtime.WithErrorHandler(runtime.ErrorHandlerFunc(grpcutil.HTTPGatewayErrorHandler)))

		server := http.Server{Handler: mux}
		lc.Append(fx.Hook{
			OnStart: func(c context.Context) error {
				logger.Infof("http listen on %s", addr)
				listener, err := net.Listen("tcp", addr)
				if err != nil {
					return err
				}
				go server.Serve(listener)

				return nil
			},
			OnStop: func(c context.Context) error {
				return server.Shutdown(c)
			},
		})
		return mux, nil
	}
}
