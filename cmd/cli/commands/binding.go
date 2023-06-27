package commands

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/JoeReid/jetbridge/cmd/cli/prettyprint"
	v1 "github.com/JoeReid/jetbridge/proto/gen/go/jetbridge/v1"
	"github.com/JoeReid/jetbridge/proto/gen/go/jetbridge/v1/v1connect"
	"github.com/bufbuild/connect-go"
	"github.com/urfave/cli/v2"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
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
	startFrom       string
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
		&cli.StringFlag{
			Name:        "start-from",
			Usage:       "Where to begin reading messages. Either an integer sequence number, a timestamp or the special values 'all', 'last', 'last-per-subject' or 'new'",
			Required:    false,
			Value:       "all",
			Destination: &startFrom,
		},
	},
	Action: func(c *cli.Context) error {
		client := v1connect.NewJetbridgeServiceClient(http.DefaultClient, ServerURL)

		ctx, cancel := context.WithTimeout(c.Context, time.Minute)
		defer cancel()

		req := &v1.CreateBindingRequest{
			LambdaArn:       lambdaARN,
			Stream:          stream,
			SubjectPattern:  subject,
			MaxBatchSize:    int64(maxBatchSize),
			MaxBatchLatency: durationpb.New(maxBatchLatency),
		}

		switch startFrom {
		case "all", "last", "last-per-subject", "new":
			req.DeliveryPolicy = &v1.CreateBindingRequest_Policy{
				Policy: startFrom,
			}

		default:
			t, tErr := time.Parse(time.RFC3339, startFrom)
			if tErr == nil {
				req.DeliveryPolicy = &v1.CreateBindingRequest_StartTime{StartTime: timestamppb.New(t)}
			}

			i, iErr := strconv.ParseUint(startFrom, 10, 64)
			if iErr == nil {
				req.DeliveryPolicy = &v1.CreateBindingRequest_StartSequence{StartSequence: i}
			}

			if tErr != nil && iErr != nil {
				return errors.New("start-from must be either an integer sequence number, a timestamp or the special values 'all', 'last', 'last-per-subject' or 'new'")
			}
		}

		resp, err := client.CreateBinding(ctx, connect.NewRequest(req))
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
	Name:      "delete",
	Aliases:   []string{"d"},
	ArgsUsage: `ID of the binding to delete.`,
	Usage:     "delete a binding",
	Action: func(c *cli.Context) error {
		client := v1connect.NewJetbridgeServiceClient(http.DefaultClient, ServerURL)

		ctx, cancel := context.WithTimeout(c.Context, time.Minute)
		defer cancel()

		_, err := client.DeleteBinding(ctx, connect.NewRequest(&v1.DeleteBindingRequest{Id: c.Args().First()}))
		if err != nil {
			return err
		}

		return nil
	},
}
