package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/urfave/cli/v2"
	svcGateway "github.com/voiladev/VoilaCrawl/internal/api/gateway/v1"
	reqCtrl "github.com/voiladev/VoilaCrawl/internal/controller/request"
	"github.com/voiladev/VoilaCrawl/internal/controller/request/history"
	historyCtrl "github.com/voiladev/VoilaCrawl/internal/controller/request/history"
	historyManager "github.com/voiladev/VoilaCrawl/internal/model/request/history/manager"
	reqManager "github.com/voiladev/VoilaCrawl/internal/model/request/manager"
	"github.com/voiladev/VoilaCrawl/pkg/types"
	pbCrawl "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/grpcutil"
	"github.com/voiladev/go-framework/invocation"
	"github.com/voiladev/go-framework/mysql"
	"github.com/voiladev/go-framework/redis"
	pbDesc "github.com/voiladev/protobuf/protoc-gen-go/protobuf"
	"go.uber.org/fx"
	grpc "google.golang.org/grpc"
)

var (
	buildBranch string
	buildCommit string
	buildTime   string

	// Version The version string
	Version = fmt.Sprintf("Branch [%s] Commit [%s] Build Time [%s]", buildBranch, buildCommit, buildTime)
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
	cancel     context.CancelFunc
	exitChan   <-chan os.Signal
	closeQueue []io.Closer
}

func NewApp(exitChan <-chan os.Signal) *App {
	ctx, cancel := context.WithCancel(context.Background())
	return &App{
		ctx:      ctx,
		cancel:   cancel,
		exitChan: exitChan,
	}
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
			Name:  "redis-addr",
			Usage: "redis server address",
			Value: "127.0.0.1:6379",
		},
		&cli.StringSliceFlag{
			Name:  "nsqlookupd-http-addr",
			Usage: "nsqlookupd http address",
			Value: cli.NewStringSlice("voiladev.com:4161"),
		},
		&cli.StringFlag{
			Name:  "nsqd-tcp-addr",
			Usage: "nsqd tcp address",
			Value: "voiladev.com:4150",
		},
		&cli.StringFlag{
			// TODO: added support for crawlet client group
			// currently, only support one grpc
			Name:  "crawlet-addr",
			Usage: "crawlet server grpc address",
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
			fx.Provide(reqManager.NewRequestManager),
			fx.Provide(historyManager.NewHistoryManager),

			// Controller
			fx.Provide(func() (pbCrawl.CrawlerManagerClient, error) {
				crawletAddr := c.String("crawlet-addr")
				if crawletAddr == "" {
					return nil, errors.New("invalid crawlet address")
				}
				conn, err := grpc.DialContext(app.ctx, crawletAddr,
					grpc.WithInsecure(),
					grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(100*1024*1024)),
					grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(100*1024*1024)),
				)
				if err != nil {
					return nil, err
				}
				return pbCrawl.NewCrawlerManagerClient(conn), nil
			}),
			fx.Provide(func() historyCtrl.RequestHistoryControllerOptions {
				return historyCtrl.RequestHistoryControllerOptions{NsqLookupdAddresses: c.StringSlice("nsqlookupd-http-addr")}
			}),
			fx.Provide(history.NewRequestHistoryController),

			fx.Provide(func() reqCtrl.RequestControllerOptions {
				return reqCtrl.RequestControllerOptions{NsqLookupdAddresses: c.StringSlice("nsqlookupd-http-addr")}
			}),
			fx.Provide(reqCtrl.NewRequestController),

			// Register services
			fx.Provide(svcGateway.NewGatewayServer),

			// Register grpc handler
			fx.Invoke(pbCrawl.RegisterGatewayServer),

			// Register http handler
			fx.Invoke(pbCrawl.RegisterGatewayHandler),
		)

		depInj := fx.New(options...)
		if err := depInj.Start(app.ctx); err != nil {
			return cli.NewExitError(err, 1)
		}

		<-app.exitChan
		app.cancel()
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
			&types.RequestHistory{},
		),
	); err != nil {
		return nil, err
	} else {
		opts = append(opts, fx.Provide(ins.Instance))
	}

	if redisClient, err := redis.NewRedisClient(redis.RedisClientOptions{
		URI:          c.String("redis-addr"),
		MaxIdelConns: 10,
	}); err != nil {
		return nil, err
	} else {
		opts = append(opts, fx.Provide(func() *redis.RedisClient {
			return redisClient
		}))
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
		server := grpc.NewServer(
			grpc.UnaryInterceptor(interceptor.UnaryInterceptor),
			grpc.StreamInterceptor(interceptor.StreamInterceptor),
			grpc.MaxRecvMsgSize(100*1024*1024),
			grpc.MaxSendMsgSize(100*1024*1024),
		)

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
				conn, err = grpc.Dial(grpcAddr,
					grpc.WithInsecure(),
					grpc.WithBackoffMaxDelay(time.Second),
					grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(100*1024*1024)),
					grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(100*1024*1024)),
				)
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

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	NewApp(c).Run(os.Args)
}
