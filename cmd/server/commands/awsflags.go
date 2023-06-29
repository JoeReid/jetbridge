package commands

import "github.com/urfave/cli/v2"

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
