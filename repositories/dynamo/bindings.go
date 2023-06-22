package dynamo

import (
	"context"
	"fmt"
	"time"

	"github.com/JoeReid/go-rendezvous"
	"github.com/JoeReid/jetbridge/repositories"
	"github.com/google/uuid"
	"github.com/guregu/dynamo"
	"golang.org/x/sync/errgroup"
)

var _ repositories.Bindings = (*Bindings)(nil)

type Bindings struct {
	db        *dynamo.DB
	tableName string
}

func (b *Bindings) ListJetstreamBindings(ctx context.Context) ([]repositories.JetstreamBinding, error) {
	eg, ctx := errgroup.WithContext(ctx)

	var (
		peerIDs  []string
		bindings []repositories.JetstreamBinding
	)

	eg.Go(func() error {
		query := b.db.Table(b.tableName).
			Get("pk", &peerPK{})

		var peers []peerRecord
		if err := query.AllWithContext(ctx, &peers); err != nil {
			return err
		}

		for _, peer := range peers {
			peerIDs = append(peerIDs, peer.ID.String())
		}
		return nil
	})

	eg.Go(func() error {
		query := b.db.Table(b.tableName).
			Get("pk", &jetstreamBindingPK{})

		var jetstreamBindings []jetstreamBindingRecord
		if err := query.AllWithContext(ctx, &jetstreamBindings); err != nil {
			return err
		}

		for _, binding := range jetstreamBindings {
			bindings = append(bindings, repositories.JetstreamBinding{
				ID:                 binding.ID,
				NatsStream:         binding.NatsStream,
				NatsConsumer:       binding.NatsConsumer,
				NatsSubjectPattern: binding.NatsSubjectPattern,
				LambdaARN:          binding.LambdaARN,
				Batching: &repositories.JetstreamBindingBatching{
					MaxMessages: binding.Batching.MaxMessages,
					MaxLatency:  binding.Batching.MaxLatency,
				},
			})
		}

		return nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	// Determine the assigned peer using rendezvous hashing
	if len(peerIDs) > 0 {
		h := rendezvous.NewHasher(rendezvous.WithMembers(peerIDs...))

		for _, binding := range bindings {
			assignedPeerID, err := uuid.Parse(h.Owner(binding.ID.String()))
			if err != nil {
				return nil, err
			}

			binding.AssignedPeerID = &assignedPeerID
		}
	}

	return bindings, nil
}

func NewBindings(db *dynamo.DB, tableName string) (*Bindings, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := db.Table(tableName).WaitWithContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to wait for table %s: %w", tableName, err)
	}

	return &Bindings{
		db:        db,
		tableName: tableName,
	}, nil
}
