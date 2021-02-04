package main

import (
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
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	NewApp(c).Run(os.Args)
}
