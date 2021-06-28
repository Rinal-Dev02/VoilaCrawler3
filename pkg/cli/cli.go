package cli

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/urfave/cli/v2"
	"github.com/voiladev/VoilaCrawler/pkg/crawler"
	"github.com/voiladev/VoilaCrawler/pkg/net/http"
	pbCrawl "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl"
	"github.com/voiladev/go-framework/glog"
	pbDesc "github.com/voiladev/protobuf/protoc-gen-go/protobuf"
)

var (
	c = make(chan os.Signal, 1)
)

type (
	New        = func(http.Client, glog.Log) (crawler.Crawler, error)
	NewWithApp = func(*cli.Context, http.Client, glog.Log) (crawler.Crawler, error)
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
	newFunc   NewWithApp
	servePort int
}

func NewApp(newFunc interface{}, flags ...cli.Flag) *App {
	var f NewWithApp
	if v, ok := newFunc.(NewWithApp); ok {
		f = v
	} else if v, ok := newFunc.(New); ok {
		f = func(c *cli.Context, client http.Client, logger glog.Log) (crawler.Crawler, error) {
			return (v)(client, logger)
		}
	} else {
		panic("unsupported new function")
	}

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
		newFunc:    f,
	}

	app.cliApp.Name = "crawler"
	if buildName != "" {
		app.cliApp.Name = buildName
	}
	app.cliApp.Usage = "crawler node"
	app.cliApp.Version = app.version
	app.cliApp.Commands = []*cli.Command{
		serveCommand(ctx, &app, app.newFunc, flags),
		localCommand(ctx, &app, app.newFunc, flags),
	}
	return &app
}

func (app *App) Run(args []string) error {
	return app.cliApp.Run(args)
}

func (app *App) Exit() {
	if app == nil || app.cancelFunc == nil {
		return
	}
	app.cancelFunc()
}
