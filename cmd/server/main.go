package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/JoeReid/jetbridge/cmd/server/app"
	"github.com/urfave/cli/v2"
)

func main() {
	root := &cli.App{
		Name:  "jetbridge-server",
		Usage: "The (un-official) bridge between NATS and AWS Lambda",
		Commands: []*cli.Command{
			app.ServeCommand,
			app.CreateTableCommand,
		},
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := root.RunContext(ctx, os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}
