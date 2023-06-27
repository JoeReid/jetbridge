package repositories

import (
	"time"

	"github.com/google/uuid"
)

type JetstreamBinding struct {
	ID             uuid.UUID
	LambdaARN      string
	Stream         string
	Consumer       uuid.UUID
	Subject        string
	MaxMessages    int
	MaxLatency     time.Duration
	DeliveryPolicy string
	AssignedPeerID *uuid.UUID
}

type CreateJetstreamBinding struct {
	LambdaARN      string
	Stream         string
	Subject        string
	MaxMessages    int
	MaxLatency     time.Duration
	DeliveryPolicy string
}
