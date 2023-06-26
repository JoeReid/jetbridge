package commands

import (
	"context"
	"net/http"
	"time"

	"github.com/JoeReid/jetbridge/cmd/cli/prettyprint"
	v1 "github.com/JoeReid/jetbridge/proto/gen/go/jetbridge/v1"
	"github.com/JoeReid/jetbridge/proto/gen/go/jetbridge/v1/v1connect"
	"github.com/bufbuild/connect-go"
	"github.com/urfave/cli/v2"
)

var Peer = &cli.Command{
	Name:    "peer",
	Aliases: []string{"p"},
	Usage:   "subcommands for managing peers",
	Subcommands: []*cli.Command{
		PeerList,
	},
}

var PeerList = &cli.Command{
	Name:    "list",
	Aliases: []string{"l"},
	Usage:   "list all peers",
	Action: func(c *cli.Context) error {
		client := v1connect.NewJetbridgeServiceClient(http.DefaultClient, ServerURL)

		ctx, cancel := context.WithTimeout(c.Context, time.Minute)
		defer cancel()

		resp, err := client.ListPeers(ctx, connect.NewRequest(&v1.ListPeersRequest{}))
		if err != nil {
			return err
		}

		prettyprint.Peers(resp.Msg.Peers)
		return nil
	},
}
