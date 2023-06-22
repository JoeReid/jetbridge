package dynamo

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"github.com/ory/dockertest/v3"
)

func testingDynamoDB(t *testing.T) *dynamo.DB {
	t.Helper()

	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("Could construct docker pool: %s", err)
	}

	if err := pool.Client.Ping(); err != nil {
		t.Fatalf("Could not connect to Docker: %s", err)
	}

	dynamoContainer, err := pool.Run("amazon/dynamodb-local", "latest", nil)
	if err != nil {
		t.Fatalf("Could not start dynamo container: %s", err)
	}

	s, err := session.NewSession()
	if err != nil {
		t.Fatalf("Could not create aws session: %s", err)
	}

	c := &aws.Config{
		Region:      aws.String("us-east-1"),
		Endpoint:    aws.String("http://localhost:" + dynamoContainer.GetPort("8000/tcp")),
		Credentials: credentials.NewStaticCredentials("test", "test", ""),
	}

	if err := pool.Retry(func() error {
		db := dynamo.New(s, c)

		_, err := db.ListTables().All()
		return err
	}); err != nil {
		t.Fatalf("Could not connect to dynamo container: %s", err)
	}

	t.Cleanup(func() {
		if err := pool.Purge(dynamoContainer); err != nil {
			t.Fatalf("Could not purge dynamo container: %s", err)
		}
	})

	return dynamo.New(s, c)
}
