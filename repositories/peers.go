package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
)

//go:generate go run github.com/golang/mock/mockgen -destination=./mocks/mock_peers.go -package=mocks github.com/JoeReid/jetbridge/repositories Peers

type Peers interface {
	JoinPeers(ctx context.Context) (*Peer, error)
	SendHeartbeat(ctx context.Context, id uuid.UUID) (*Peer, error)
	LeavePeers(ctx context.Context, id uuid.UUID) error

	ListPeers(ctx context.Context) ([]Peer, error)
}

type Peer struct {
	ID             uuid.UUID
	Hostname       string
	JoinedAt       time.Time
	LastSeenAt     time.Time
	HeartbeatDueBy time.Time
}
