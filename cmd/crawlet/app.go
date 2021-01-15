package main

import (
	"context"
	"io"
	"os"

	"github.com/urfave/cli/v2"
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
			Name: "",
		},
		&cli.StringSliceFlag{
			Name:  "nsqlookupd-http-addr",
			Usage: "nsqlookupd http address",
			Value: cli.NewStringSlice("voiladev.com:4161"),
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

		<-app.exitChan
		return nil
	}
	cliApp.Run(args)
}
