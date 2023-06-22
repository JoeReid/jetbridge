package app

import "github.com/urfave/cli/v2"

var (
	awsRegion string

	awsRegionFlag = &cli.StringFlag{
		Name:        "aws-region",
		EnvVars:     []string{"AWS_REGION"},
		Usage:       "The AWS region to use for Lambda",
		Value:       "us-east-1",
		Destination: &awsRegion,
	}
)

var (
	awsAccessKeyID string

	awsAccessKeyIDFlag = &cli.StringFlag{
		Name:        "aws-access-key-id",
		EnvVars:     []string{"AWS_ACCESS_KEY_ID"},
		Usage:       "The AWS access key ID to use for AWS",
		Required:    true,
		Destination: &awsAccessKeyID,
	}
)

var (
	awsSecretAccessKey string

	awsSecretAccessKeyFlag = &cli.StringFlag{
		Name:        "aws-secret-access-key",
		EnvVars:     []string{"AWS_SECRET_ACCESS_KEY"},
		Usage:       "The AWS secret access key to use for AWS",
		Required:    true,
		Destination: &awsSecretAccessKey,
	}
)

var (
	dynamoEndpoint string

	dynamoEndpointFlag = &cli.StringFlag{
		Name:        "dynamo-endpoint",
		EnvVars:     []string{"DYNAMO_ENDPOINT"},
		Usage:       "The endpoint to use for DynamoDB, used for local development and testing",
		Destination: &dynamoEndpoint,
	}
)

var (
	dynamoTable string

	dynamoTableFlag = &cli.StringFlag{
		Name:        "table-name",
		EnvVars:     []string{"TABLE_NAME"},
		Usage:       "The name of the DynamoDB table to use to store internal state",
		Value:       "jetbridge",
		Destination: &dynamoTable,
	}
)

var (
	lambdaEndpoint string

	lambdaEndpointFlag = &cli.StringFlag{
		Name:        "lambda-endpoint",
		EnvVars:     []string{"LAMBDA_ENDPOINT"},
		Usage:       "The endpoint to use for Lambda, used for local development and testing",
		Destination: &lambdaEndpoint,
	}
)
