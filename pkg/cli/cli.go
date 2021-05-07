package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/urfave/cli/v2"
	"github.com/voiladev/go-crawler/pkg/crawler"
	"github.com/voiladev/go-crawler/pkg/net/http"
	"github.com/voiladev/go-crawler/pkg/net/http/cookiejar"
	"github.com/voiladev/go-crawler/pkg/proxy"
	pbCrawl "github.com/voiladev/go-crawler/protoc-gen-go/chameleon/smelter/v1/crawl"
	pbSession "github.com/voiladev/go-crawler/protoc-gen-go/chameleon/smelter/v1/crawl/session"
	"github.com/voiladev/go-framework/glog"
	"github.com/voiladev/go-framework/grpcutil"
	"github.com/voiladev/go-framework/invocation"
	pbDesc "github.com/voiladev/protobuf/protoc-gen-go/protobuf"
	"go.uber.org/fx"
	grpc "google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/anypb"
)

var (
	c = make(chan os.Signal, 1)
)

func init() {
	signal.Notify(c, os.Interrupt)
}

var (
	buildName   string
	buildBranch string
	buildCommit string
	buildTime   string

	// BuildVersion
	Version = fmt.Sprintf("Branch [%s] Commit [%s] Build Time [%s]", buildBranch, buildCommit, buildTime)
)

var _ServiceDescs = map[string]*pbDesc.ServiceDesc{}

func init() {
	rand.Seed(time.Now().UnixNano())

	for _, desc := range pbCrawl.ServiceDescs {
		_ServiceDescs[desc.GetFullname()] = desc
	}
}

func getOnePort() int {
	return func() int {
		for port := rand.Int()%30000 + 3000; port < 65535; port++ {
			listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
			if err != nil {
				continue
			}
			listener.Close()
			return port
		}
		return 0
	}()
}

// App
type App struct {
	cliApp     *cli.App
	ctx        context.Context
	cancelFunc context.CancelFunc

	version   string
	newFunc   crawler.New
	servePort int
}

func NewApp(newFunc crawler.New) *App {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-c
		cancel()
	}()

	app := App{
		cliApp:     cli.NewApp(),
		ctx:        ctx,
		cancelFunc: cancel,
		version:    Version,
		newFunc:    newFunc,
	}
	app.cliApp.Name = "crawler"
	if buildName != "" {
		app.cliApp.Name = buildName
	}
	app.cliApp.Usage = "crawler node"
	app.cliApp.Version = app.version
	app.cliApp.Commands = []*cli.Command{
		localCommand(ctx, &app, newFunc),
	}
	return &app
}

func (app *App) Run(args []string) error {
	cmd := cli.Command{
		Name:  "serve",
		Usage: "run crawler server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "host",
				Value: "0.0.0.0",
				Usage: "The bind host",
			},
			&cli.IntFlag{
				Name:  "port",
				Usage: "The bind (grpc) port, if not specified, will randomly choise one",
			},
			&cli.StringFlag{
				Name:    "proxy-addr",
				Usage:   "proxy server address",
				EnvVars: []string{"VOILA_PROXY_URL"},
			},
			&cli.StringFlag{
				Name:  "session-addr",
				Usage: "session server grpc address, if not provided, use local cookiejar",
			},
			&cli.StringFlag{
				Name:  "crawlet-addr",
				Usage: "crawlet server grpc address",
			},
			&cli.BoolFlag{
				Name:    "debug",
				Usage:   "Enable debug",
				EnvVars: []string{"DEBUG"},
			},
		},
		Action: func(c *cli.Context) error {
			logger := glog.New(glog.LogLevelInfo)
			if c.Bool("debug") {
				logger.SetLevel(glog.LogLevelDebug)
				os.Setenv("DEBUG", "1")
			}

			options := []fx.Option{
				fx.Provide(func() context.Context {
					return app.ctx
				}),
				fx.Provide(func() glog.Log { return logger }),
				fx.Logger(logger),
			}

			options = append(options,
				// grpc server
				fx.Provide(app.newGrpcServer(c)),
				fx.Provide(func(logger glog.Log) (http.CookieJar, error) {
					sessionAddr := c.String("session-addr")
					if sessionAddr == "" {
						logger.Warnf("served with local session manager")
						return cookiejar.New(), nil
					} else {
						conn, err := grpc.DialContext(app.ctx, sessionAddr, grpc.WithInsecure(), grpc.WithBlock())
						if err != nil {
							return nil, err
						}
						return cookiejar.NewRemoteJar(pbSession.NewSessionManagerClient(conn), logger)
					}
				}),
				fx.Provide(func(jar http.CookieJar, logger glog.Log) (http.Client, error) {
					proxyAddr := c.String("proxy-addr")
					if proxyAddr == "" {
						return nil, errors.New("proxy address not specified")
					}
					return proxy.NewProxyClient(proxyAddr, jar, logger)
				}),
				fx.Provide(app.newFunc),

				// Register services
				fx.Provide(NewCrawlerServer),
				// Register grpc handler
				fx.Invoke(pbCrawl.RegisterCrawlerNodeServer),
				fx.Invoke(func(crawler crawler.Crawler, logger glog.Log) error {
					// register to crawlet
					crawletAddr := c.String("crawlet-addr")
					if crawletAddr == "" {
						return errors.New("invalid crawlet server address")
					}
					conn, err := grpc.DialContext(app.ctx, crawletAddr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(time.Second*10))
					if err != nil {
						if err != context.Canceled {
							logger.Errorf("connect %s failed, error=%s", crawletAddr, err)
						}
						return err
					}

					go func() {
						defer conn.Close()

						for {
							func() error {
								registerClient := pbCrawl.NewCrawlerRegisterClient(conn)
								client, err := registerClient.Connect(app.ctx)
								if err != nil {
									logger.Errorf("connect to crawlet failed, error=%s", err)
									return err
								}
								data, _ := anypb.New(&pbCrawl.ConnectRequest_Ping{
									Timestamp:      time.Now().Unix(),
									Id:             crawler.ID(),
									StoreId:        crawler.ID(),
									Version:        crawler.Version(),
									AllowedDomains: crawler.AllowedDomains(),
									ServePort:      int32(app.servePort),
								})
								if err := client.Send(data); err != nil {
									logger.Errorf("register info to crawlet failed, error=%s", err)
									return err
								}
								logger.Infof("connected to crawlet")

								ticker := time.NewTicker(time.Second * 5)
								for {
									select {
									case <-app.ctx.Done():
										return app.ctx.Err()
									case <-ticker.C:
										data, _ := anypb.New(&pbCrawl.ConnectRequest_Heartbeat{Timestamp: time.Now().Unix()})
										if err := client.Send(data); err != nil {
											if err == io.EOF {
												logger.Errorf("connection closed")
												return err
											}
											logger.Errorf("send heartbeat failed, error=%s", err)
											return err
										}
									}
								}
							}()

							select {
							case <-app.ctx.Done():
								return
							default:
							}
							logger.Infof("reconnect after 10 seconds")
							time.Sleep(time.Second * 5)

							select {
							case <-app.ctx.Done():
								return
							default:
							}
						}
					}()
					return nil
				}),
			)

			depInj := fx.New(options...)
			if err := depInj.Start(app.ctx); err != nil {
				return cli.NewExitError(err, 1)
			}

			<-app.ctx.Done()
			depInj.Stop(app.ctx)
			return nil
		},
	}
	app.cliApp.Commands = append(app.cliApp.Commands, &cmd)

	return app.cliApp.Run(args)
}

func (app *App) Exit() {
	if app == nil || app.cancelFunc == nil {
		return
	}
	app.cancelFunc()
}

func (app *App) newGrpcServer(c *cli.Context) func(fx.Lifecycle, glog.Log) (grpc.ServiceRegistrar, error) {
	port := c.Int("port")
	if port == 0 {
		port = getOnePort()
	}
	app.servePort = port
	addr := fmt.Sprintf("%s:%d", c.String("host"), port)

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
