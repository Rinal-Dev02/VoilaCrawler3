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
	sessionSvc "github.com/voiladev/VoilaCrawl/internal/api/session/v1"
	cookieManager "github.com/voiladev/VoilaCrawl/internal/model/cookie/manager"
	pbSession "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/smelter/v1/crawl/session"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/grpcutil"
	"github.com/voiladev/go-framework/invocation"
	"github.com/voiladev/go-framework/redis"
	pbDesc "github.com/voiladev/protobuf/protoc-gen-go/protobuf"
	"go.uber.org/fx"
	grpc "google.golang.org/grpc"
)

var _ServiceDescs = map[string]*pbDesc.ServiceDesc{}

func init() {
	for _, desc := range pbSession.ServiceDescs {
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
	cliApp.Name = "session-api"
	cliApp.Usage = "session api server"
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
			Name:  "redis-addr",
			Usage: "redis server address",
			Value: "127.0.0.1:6379",
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
			fx.Provide(cookieManager.NewCookieManager),

			// Register services
			fx.Provide(sessionSvc.NewSessionServer),

			// Register grpc handler
			fx.Invoke(pbSession.RegisterSessionManagerServer),

			// Register http handler
			fx.Invoke(pbSession.RegisterSessionManagerHandler),
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
