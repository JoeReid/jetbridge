package commands

import "github.com/urfave/cli/v2"

var Root = &cli.App{
	Name:  "jetbridge-server",
	Usage: "The (un-official) bridge between NATS and AWS Lambda",
	Commands: []*cli.Command{
		ServeCommand,
		CreateTableCommand,
	},
}
