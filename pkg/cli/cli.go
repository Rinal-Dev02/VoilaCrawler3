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
	pbCrawl "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/smelter/v1/crawl"
	pbDesc "github.com/voiladev/protobuf/protoc-gen-go/protobuf"
)

type (
	Context = cli.Context
)

var (
	c = make(chan os.Signal, 1)
)

// type (
// 	New        = func(http.Client, glog.Log) (interface{}, error)
// 	NewWithApp = func(*cli.Context, http.Client, glog.Log) (interface{}, error)
// )

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
	servePort int
}

func NewApp(crawler crawler.NewCrawler, flags ...cli.Flag) *App {
	if c == nil {
		panic("Require crawler instance")
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
	}

	app.cliApp.Name = "crawler"
	if buildName != "" {
		app.cliApp.Name = buildName
	}
	app.cliApp.Usage = "crawler node"
	app.cliApp.Version = app.version
	app.cliApp.Commands = []*cli.Command{
		serveCommand(ctx, &app, crawler, flags),
		localCommand(ctx, &app, crawler, flags),
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
