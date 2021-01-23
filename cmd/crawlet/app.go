package main

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/urfave/cli/v2"
	"github.com/voiladev/VoilaCrawl/pkg/net/http/proxycrawl"
	"github.com/voiladev/go-framework/glog"
)

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
	cliApp.Name = "crawlet"
	cliApp.Version = Version
	cliApp.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "crawlet-id",
			Usage: "node unique id",
			Value: NodeId(),
		},
		&cli.StringFlag{
			Name:  "crawl-addr",
			Usage: "Crawl server grpc address",
		},
		&cli.StringFlag{
			Name:  "account-addr",
			Usage: "(TODO) account server grpc addr, used to get website auth info includes cookie...",
		},
		&cli.IntFlag{
			Name:  "max-currency",
			Usage: "max goroutines in currency",
			Value: runtime.NumCPU(),
		},
		&cli.StringFlag{
			Name:  "plugins",
			Usage: "the dir of plugins",
			Value: "./plugins",
		},
		&cli.StringFlag{
			Name:  "proxy-api-token",
			Usage: "proxy api token",
			Value: "C1hwEn7zzYhHptBUoZFisQ",
		},
		&cli.StringFlag{
			Name:  "proxy-js-token",
			Usage: "proxy js api token",
			Value: "YOhYOQ6Ppd17eK9ACA54cw",
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

		httpClient, err := proxycrawl.NewProxyCrawlClient(
			proxycrawl.WithAPITokenOption(c.String("proxy-api-token")),
			proxycrawl.WithJSTokenOption(c.String("proxy-js-token")),
		)
		if err != nil {
			logger.Error(err)
			return cli.NewExitError(err, 1)
		}

		crawlerManager, err := NewCrawlerManager(logger)
		if err != nil {
			logger.Error(err)
			return cli.NewExitError(err, 1)
		}
		// load plugins
		var loadedPlugintCount int
		if err := filepath.Walk(c.String("plugins"), func(p string, info os.FileInfo, err error) error {
			if info == nil || info.IsDir() || filepath.Ext(p) != ".so" {
				return nil
			}

			if cl, err := NewCrawler(httpClient, p, logger); err != nil {
				logger.Errorf("load plugin %s failed, error=%s", p, err)
				return err
			} else {
				crawlerManager.Save(app.ctx, cl)
				logger.Infof("loaded plugin %s", cl.ID())
				loadedPlugintCount += 1
			}
			return nil
		}); err != nil {
			return cli.NewExitError(err, 1)
		}
		if loadedPlugintCount == 0 {
			return cli.NewExitError("no usable plugins", 1)
		}

		conn, err := NewConnection(app.ctx, c.String("crawl-addr"), logger)
		if err != nil {
			logger.Error(err)
			return cli.NewExitError(err, 1)
		}

		crawlerCtrl, err := NewCrawlerController(
			app.ctx,
			crawlerManager,
			httpClient,
			conn,
			&CrawlerControllerOptions{MaxConcurrency: int32(c.Int("max-currency"))},
			logger,
		)
		if err != nil {
			logger.Error(err)
			return cli.NewExitError(err, 1)
		}
		go crawlerCtrl.Run(app.ctx)

		<-app.exitChan
		return nil
	}
	cliApp.Run(args)
}
