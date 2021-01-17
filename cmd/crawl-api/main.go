package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
)

var (
	buildBranch string
	buildCommit string
	buildTime   string

	// Version The version string
	Version = fmt.Sprintf("Branch [%s] Commit [%s] Build Time [%s]", buildBranch, buildCommit, buildTime)
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	NewApp(ctx, c).Run(os.Args)
}
