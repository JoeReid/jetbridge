package dynamo

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/JoeReid/jetbridge/repositories"
	"github.com/google/uuid"
	"github.com/guregu/dynamo"
)

var _ repositories.Peers = (*Peers)(nil)

type Peers struct {
	db        *dynamo.DB
	tableName string
	peerTTL   time.Duration
}

func (s *Peers) JoinPeers(ctx context.Context) (*repositories.Peer, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	record := &peerRecord{
		PK:          &peerPK{},
		ID:          uuid.New(),
		Name:        hostname,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		DeleteAfter: time.Now().Add(s.peerTTL),
	}

	query := s.db.
		Table(s.tableName).
		Put(record).
		If("attribute_not_exists(pk)")

	if err := query.RunWithContext(ctx); err != nil {
		return nil, err
	}

	return &repositories.Peer{
		ID:             record.ID,
		Hostname:       record.Name,
		JoinedAt:       record.CreatedAt,
		LastSeenAt:     record.UpdatedAt,
		HeartbeatDueBy: record.DeleteAfter,
	}, nil
}

func (s *Peers) SendHeartbeat(ctx context.Context, id uuid.UUID) (*repositories.Peer, error) {
	query := s.db.
		Table(s.tableName).
		Update("pk", &peerPK{}).
		Range("sk", id).
		Set("updated_at", time.Now()).
		Set("delete_after", time.Now().Add(s.peerTTL)).
		If("attribute_exists(pk)")

	var row peerRecord
	if err := query.ValueWithContext(ctx, &row); err != nil {
		return nil, err
	}

	return &repositories.Peer{
		ID:             row.ID,
		Hostname:       row.Name,
		JoinedAt:       row.CreatedAt,
		LastSeenAt:     row.UpdatedAt,
		HeartbeatDueBy: row.DeleteAfter,
	}, nil
}

func (s *Peers) LeavePeers(ctx context.Context, id uuid.UUID) error {
	query := s.db.
		Table(s.tableName).
		Update("pk", &peerPK{}).
		Range("sk", id).
		Set("updated_at", time.Now()).
		Set("delete_after", time.Now()).
		If("attribute_exists(pk)")

	err := query.RunWithContext(ctx)
	return err
}

func (s *Peers) ListPeers(ctx context.Context) ([]repositories.Peer, error) {
	query := s.db.
		Table(s.tableName).
		Get("pk", &peerPK{}).
		Filter("delete_after > ?", time.Now())

	var rows []peerRecord
	if err := query.AllWithContext(ctx, &rows); err != nil {
		return nil, err
	}

	var peers []repositories.Peer
	for _, row := range rows {
		peers = append(peers, repositories.Peer{
			ID:             row.ID,
			Hostname:       row.Name,
			JoinedAt:       row.CreatedAt,
			LastSeenAt:     row.UpdatedAt,
			HeartbeatDueBy: row.DeleteAfter,
		})
	}

	return peers, nil
}

func NewPeers(db *dynamo.DB, tableName string) (*Peers, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := db.Table(tableName).WaitWithContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to wait for table %s: %w", tableName, err)
	}

	return &Peers{
		db:        db,
		tableName: tableName,
		peerTTL:   5 * time.Second, // TODO: Make configurable? or make it an application wide setting/constant?
	}, nil
}
