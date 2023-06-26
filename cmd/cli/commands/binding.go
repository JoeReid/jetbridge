package commands

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/JoeReid/jetbridge/cmd/cli/prettyprint"
	v1 "github.com/JoeReid/jetbridge/proto/gen/go/jetbridge/v1"
	"github.com/JoeReid/jetbridge/proto/gen/go/jetbridge/v1/v1connect"
	"github.com/bufbuild/connect-go"
	"github.com/urfave/cli/v2"
	"google.golang.org/protobuf/types/known/durationpb"
)

var Binding = &cli.Command{
	Name:    "binding",
	Aliases: []string{"b"},
	Usage:   "subcommands for managing bindings",
	Subcommands: []*cli.Command{
		BindingCreate,
		BindingList,
		BindingDelete,
	},
}

var (
	lambdaARN       string
	stream          string
	subject         string
	maxBatchSize    int
	maxBatchLatency time.Duration
)

var BindingCreate = &cli.Command{
	Name:    "create",
	Aliases: []string{"c"},
	Usage:   "create a new binding",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "lambda",
			Usage:       "the ARN of the lambda to invoke",
			Required:    true,
			Destination: &lambdaARN,
		},
		&cli.StringFlag{
			Name:        "stream",
			Usage:       "the stream to read from",
			Required:    true,
			Destination: &stream,
		},
		&cli.StringFlag{
			Name:        "subject",
			Usage:       "the subject pattern to apply to the consumer",
			Required:    true,
			Destination: &subject,
		},
		&cli.IntFlag{
			Name:        "max-batch-size",
			Usage:       "the maximum number of messages to batch together before sending to the lambda",
			Required:    false,
			Destination: &maxBatchSize,
		},
		&cli.DurationFlag{
			Name:        "max-batch-latency",
			Usage:       "the maximum amount of time to delay messages while waiting for the batch to fill up",
			Required:    false,
			Destination: &maxBatchLatency,
		},
	},
	Action: func(c *cli.Context) error {
		client := v1connect.NewJetbridgeServiceClient(http.DefaultClient, ServerURL)

		ctx, cancel := context.WithTimeout(c.Context, time.Minute)
		defer cancel()

		var batching *v1.BindingBatching
		switch {
		case maxBatchSize > 0 && maxBatchLatency > 0:
			batching = &v1.BindingBatching{
				MaxMessages: int64(maxBatchSize),
				MaxLatency:  durationpb.New(maxBatchLatency),
			}

		case maxBatchSize > 0 && maxBatchLatency <= 0 || maxBatchSize <= 0 && maxBatchLatency > 0:
			return errors.New("both batch size and latency must be set")
		}

		resp, err := client.CreateBinding(ctx, connect.NewRequest(&v1.CreateBindingRequest{
			LambdaArn:      lambdaARN,
			Stream:         stream,
			SubjectPattern: subject,
			Batching:       batching,
		}))
		if err != nil {
			return err
		}

		prettyprint.Binding(resp.Msg.Binding)
		return nil
	},
}

var BindingList = &cli.Command{
	Name:    "list",
	Aliases: []string{"l"},
	Usage:   "list all bindings",
	Action: func(c *cli.Context) error {
		client := v1connect.NewJetbridgeServiceClient(http.DefaultClient, ServerURL)

		ctx, cancel := context.WithTimeout(c.Context, time.Minute)
		defer cancel()

		resp, err := client.ListBindings(ctx, connect.NewRequest(&v1.ListBindingsRequest{}))
		if err != nil {
			return err
		}

		prettyprint.Bindings(resp.Msg.Bindings)
		return nil
	},
}

var BindingDelete = &cli.Command{
	Name:    "delete",
	Aliases: []string{"d"},
	Usage:   "delete a binding",
	Action: func(c *cli.Context) error {
		return errors.New("not implemented")
	},
}
