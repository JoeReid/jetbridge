package dynamo

import (
	"fmt"
	"time"

	"github.com/JoeReid/go-rendezvous"
	"github.com/JoeReid/jetbridge/repositories"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/google/uuid"
)

type jetstreamBindingRecords []*jetstreamBindingRecord

func (r jetstreamBindingRecords) toJetstreamBindings(peers []peerRecord) []repositories.JetstreamBinding {
	bindings := make([]repositories.JetstreamBinding, len(r))
	for i, record := range r {
		bindings[i] = *record.toJetstreamBinding(peers)
	}
	return bindings
}

type jetstreamBindingRecord struct {
	PK             *jetstreamBindingPK `dynamo:"pk,hash"`
	ID             uuid.UUID           `dynamo:"sk,range"`
	LambdaARN      string              `dynamo:"lambda_arn"`
	Stream         string              `dynamo:"nats_stream"`
	Consumer       uuid.UUID           `dynamo:"nats_consumer"`
	SubjectPattern string              `dynamo:"nats_subject_pattern"`
	MaxMessages    int                 `dynamo:"max_messages"`
	MaxLatency     time.Duration       `dynamo:"max_latency"`
	DeliveryPolicy string              `dynamo:"delivery_policy"`
	CreatedAt      time.Time           `dynamo:"created_at" localIndex:"created_at-index"`
	UpdatedAt      time.Time           `dynamo:"updated_at" localIndex:"updated_at-index"`
}

func (r *jetstreamBindingRecord) assignedPeerID(peerIDs ...uuid.UUID) *uuid.UUID {
	if len(peerIDs) == 0 {
		return nil
	}

	peers := make([]string, len(peerIDs))
	for i, peerID := range peerIDs {
		peers[i] = peerID.String()
	}

	h := rendezvous.NewHasher(rendezvous.WithMembers(peers...))
	owner, err := uuid.Parse(h.Owner(r.ID.String()))
	if err != nil {
		panic(fmt.Sprintf("failed to parse owner: %s", err))
	}

	return &owner
}

func (r *jetstreamBindingRecord) toJetstreamBinding(peers []peerRecord) *repositories.JetstreamBinding {
	peerIDs := make([]uuid.UUID, len(peers))
	for i, peer := range peers {
		peerIDs[i] = peer.ID
	}

	return &repositories.JetstreamBinding{
		ID:             r.ID,
		LambdaARN:      r.LambdaARN,
		Stream:         r.Stream,
		Consumer:       r.Consumer,
		Subject:        r.SubjectPattern,
		MaxMessages:    r.MaxMessages,
		MaxLatency:     r.MaxLatency,
		DeliveryPolicy: r.DeliveryPolicy,
		AssignedPeerID: r.assignedPeerID(peerIDs...),
	}
}

func newJetstreamBinding(create *repositories.CreateJetstreamBinding) (*jetstreamBindingRecord, error) {
	id := uuid.New()

	return &jetstreamBindingRecord{
		PK:             &jetstreamBindingPK{},
		ID:             id,
		LambdaARN:      create.LambdaARN,
		Stream:         create.Stream,
		Consumer:       id,
		SubjectPattern: create.Subject,
		MaxMessages:    create.MaxMessages,
		MaxLatency:     create.MaxLatency,
		DeliveryPolicy: create.DeliveryPolicy,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}, nil
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
