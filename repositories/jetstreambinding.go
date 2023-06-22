package repositories

import (
	"time"

	"github.com/google/uuid"
)

type JetstreamBinding struct {
	ID                 uuid.UUID
	NatsStream         string
	NatsConsumer       string
	NatsSubjectPattern string
	LambdaARN          string
	Batching           *JetstreamBindingBatching
	AssignedPeerID     *uuid.UUID
}

type JetstreamBindingBatching struct {
	MaxMessages int
	MaxLatency  time.Duration
}
