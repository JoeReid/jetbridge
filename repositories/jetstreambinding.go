package repositories

import (
	"time"

	"github.com/google/uuid"
)

type JetstreamBinding struct {
	ID             uuid.UUID
	LambdaARN      string
	Consumer       JetstreamConsumer
	Batching       *BindingBatching
	AssignedPeerID *uuid.UUID
}

type JetstreamConsumer struct {
	Stream  string
	Name    string
	Subject string
}

type BindingBatching struct {
	MaxMessages int
	MaxLatency  time.Duration
}
