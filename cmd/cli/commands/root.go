package commands

import (
	"github.com/urfave/cli/v2"
)

var Root = &cli.App{
	Name:  "jetbridge",
	Usage: "A bridge between NATS and AWS Lambda",
	Flags: []cli.Flag{
		serverURLFlag,
	},
	Commands: []*cli.Command{
		Peer,
		Binding,
	},
}
