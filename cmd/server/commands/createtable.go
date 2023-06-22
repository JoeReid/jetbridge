package commands

import (
	"context"
	"time"

	dynamorepo "github.com/JoeReid/jetbridge/repositories/dynamo"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"github.com/urfave/cli/v2"
)

var CreateTableCommand = &cli.Command{
	Name:  "create-table",
	Usage: "Create the DynamoDB table used to store internal state",
	Flags: []cli.Flag{
		awsRegionFlag,
		awsAccessKeyIDFlag,
		awsSecretAccessKeyFlag,
		dynamoEndpointFlag,
		dynamoTableFlag,
	},
	Action: func(c *cli.Context) error {
		ctx, cancel := context.WithTimeout(c.Context, time.Minute)
		defer cancel()

		awsSession, err := session.NewSession()
		if err != nil {
			return err
		}

		awsConfig := &aws.Config{
			Region:      &awsRegion,
			Endpoint:    &dynamoEndpoint,
			Credentials: credentials.NewStaticCredentials(awsAccessKeyID, awsSecretAccessKey, ""),
		}

		dynamoSvc := dynamo.New(awsSession, awsConfig)

		return dynamorepo.CreateTable(ctx, dynamoSvc, dynamoTable)
	},
}
