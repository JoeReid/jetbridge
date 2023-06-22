package dynamo

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/google/uuid"
)

type jetstreamBindingRecord struct {
	PK                 *jetstreamBindingPK `dynamo:"pk,hash"`
	ID                 uuid.UUID           `dynamo:"sk,range"`
	NatsStream         string              `dynamo:"nats_stream"`
	NatsConsumer       string              `dynamo:"nats_consumer"`
	NatsSubjectPattern string              `dynamo:"nats_subject_pattern"`
	LambdaARN          string              `dynamo:"lambda_arn"`
	Batching           *bindingBatching    `dynamo:"batching"`
	CreatedAt          time.Time           `dynamo:"created_at" localIndex:"created_at-index"`
	UpdatedAt          time.Time           `dynamo:"updated_at" localIndex:"updated_at-index"`
}

type jetstreamBindingPK struct{}

func (p *jetstreamBindingPK) MarshalDynamo() (*dynamodb.AttributeValue, error) {
	return &dynamodb.AttributeValue{
		S: aws.String("BINDING"),
	}, nil
}

func (p *jetstreamBindingPK) UnmarshalDynamo(av *dynamodb.AttributeValue) error {
	if av == nil || av.S == nil || *av.S == "" || *av.S != "BINDING" {
		return fmt.Errorf("invalid bindingPK: %v", av)
	}

	return nil
}

type bindingBatching struct {
	MaxMessages int           `json:"max_messages"`
	MaxLatency  time.Duration `json:"max_latency"`
}

func (b *bindingBatching) MarshalDynamo() (*dynamodb.AttributeValue, error) {
	data, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}

	return &dynamodb.AttributeValue{
		S: aws.String(string(data)),
	}, nil
}

func (b *bindingBatching) UnmarshalDynamo(av *dynamodb.AttributeValue) error {
	if av == nil || av.S == nil || *av.S == "" {
		return fmt.Errorf("invalid bindingBatching: %v", av)
	}

	return json.Unmarshal([]byte(*av.S), b)
}
