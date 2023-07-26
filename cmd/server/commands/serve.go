package commands

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/JoeReid/jetbridge/daemons"
	"github.com/JoeReid/jetbridge/proto/gen/go/jetbridge/v1/v1connect"
	dynamorepo "github.com/JoeReid/jetbridge/repositories/dynamo"
	lambdarepo "github.com/JoeReid/jetbridge/repositories/lambda"
	natsrepo "github.com/JoeReid/jetbridge/repositories/nats"
	"github.com/JoeReid/jetbridge/server"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/bufbuild/connect-go"
	grpchealth "github.com/bufbuild/connect-grpchealth-go"
	"github.com/google/uuid"
	"github.com/guregu/dynamo"
	"github.com/nats-io/nats.go"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

var (
	natsUrl  string
	httpPort int
)

var ServeCommand = &cli.Command{
	Name:  "serve",
	Usage: "Start the server",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "nats-url",
			EnvVars:     []string{"NATS_URL"},
			Usage:       "A URL indicating the NATS host to connect to",
			Value:       "nats://localhost:4222",
			Destination: &natsUrl,
		},
		dynamoEndpointFlag,
		dynamoTableFlag,
		lambdaEndpointFlag,
		&cli.IntFlag{
			Name:        "http-port",
			EnvVars:     []string{"HTTP_PORT"},
			Usage:       "The port to run the HTTP API server on",
			Value:       8080,
			Destination: &httpPort,
		},
	},
	Action: func(c *cli.Context) error {
		nc, err := nats.Connect(natsUrl)
		if err != nil {
			return err
		}
		defer nc.Close()

		js, err := nc.JetStream()
		if err != nil {
			return err
		}

		awsSession, err := session.NewSession()
		if err != nil {
			return err
		}

		dynamoSvc := dynamo.New(awsSession, aws.NewConfig().WithEndpoint(dynamoEndpoint))

		bindings, err := dynamorepo.NewBindings(dynamoSvc, dynamoTable)
		if err != nil {
			return err
		}

		peers, err := dynamorepo.NewPeers(dynamoSvc, dynamoTable)
		if err != nil {
			return err
		}

		lambdaSvc := lambda.New(awsSession, aws.NewConfig().WithEndpoint(lambdaEndpoint))

		eg, ctx := errgroup.WithContext(c.Context)

		eg.Go(func() error {
			mux := http.NewServeMux()

			mux.Handle(v1connect.NewJetbridgeServiceHandler(&server.V1{
				Bindings: bindings,
				Peers:    peers,
			}, connect.WithInterceptors(connect.UnaryInterceptorFunc(server.LoggingInterceptor))))

			mux.Handle(grpchealth.NewHandler(grpchealth.NewStaticChecker(v1connect.JetbridgeServiceName)))

			server := &http.Server{
				Addr:    fmt.Sprintf(":%d", httpPort),
				Handler: mux,
			}

			eg.Go(func() error {
				<-ctx.Done()

				ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				defer cancel()

				return server.Shutdown(ctx)
			})

			return server.ListenAndServe()
		})

		eg.Go(func() error {
			membership, ctx := daemons.NewPeerMembership(ctx, peers)

			membership.Go(func(peerID uuid.UUID) error {
				source, err := natsrepo.NewMessageSource(js)
				if err != nil {
					return err
				}

				handler, err := lambdarepo.NewMessageHandler(lambdaSvc)
				if err != nil {
					return err
				}

				jsw, err := daemons.NewJetstreamWorker(bindings, source, handler)
				if err != nil {
					return err
				}

				return jsw.Run(ctx, peerID)
			})

			return membership.Wait()
		})

		return eg.Wait()
	},
}
