package dynamo

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/google/uuid"
)

type peerRecord struct {
	PK          *peerPK   `dynamo:"pk,hash"`
	ID          uuid.UUID `dynamo:"sk,range"`
	Name        string    `dynamo:"name" localIndex:"name-index"`
	CreatedAt   time.Time `dynamo:"created_at" localIndex:"created_at-index"`
	UpdatedAt   time.Time `dynamo:"updated_at" localIndex:"updated_at-index"`
	DeleteAfter time.Time `dynamo:"delete_after,unixtime"`
}

type peerPK struct{}

func (*peerPK) MarshalDynamo() (*dynamodb.AttributeValue, error) {
	return &dynamodb.AttributeValue{
		S: aws.String("PEER"),
	}, nil
}

func (*peerPK) UnmarshalDynamo(av *dynamodb.AttributeValue) error {
	if av == nil || av.S == nil || *av.S != "PEER" {
		return fmt.Errorf("invalid peerPK: %v", av)
	}

	return nil
}
