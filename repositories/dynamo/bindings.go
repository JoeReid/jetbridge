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

func (b *Bindings) CreateJetstreamBinding(ctx context.Context, lambdaARN, stream, subject string, batching *repositories.BindingBatching) (*repositories.JetstreamBinding, error) {
	eg, ctx := errgroup.WithContext(ctx)

	var (
		peerIDs    []string
		newBinding *jetstreamBindingRecord
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
		var batch *bindingBatching

		if batching != nil {
			batch = &bindingBatching{
				MaxMessages: batching.MaxMessages,
				MaxLatency:  batching.MaxLatency,
			}
		}

		id := uuid.New()

		newBinding = &jetstreamBindingRecord{
			PK:                 &jetstreamBindingPK{},
			ID:                 id,
			LambdaARN:          lambdaARN,
			NatsStream:         stream,
			NatsConsumer:       id.String(),
			NatsSubjectPattern: subject,
			Batching:           batch,
		}

		query := b.db.Table(b.tableName).
			Put(newBinding).
			If("attribute_not_exists(pk)")

		if err := query.RunWithContext(ctx); err != nil {
			return fmt.Errorf("failed to create jetstream binding: %w", err)
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	// Determine the assigned peer using rendezvous hashing
	var assignedPeerID *uuid.UUID
	if len(peerIDs) > 0 {
		h := rendezvous.NewHasher(rendezvous.WithMembers(peerIDs...))

		owner, err := uuid.Parse(h.Owner(newBinding.ID.String()))
		if err != nil {
			return nil, err
		}
		assignedPeerID = &owner
	}

	resp := &repositories.JetstreamBinding{
		ID:        newBinding.ID,
		LambdaARN: newBinding.LambdaARN,
		Consumer: repositories.JetstreamConsumer{
			Stream:  newBinding.NatsStream,
			Name:    newBinding.NatsConsumer,
			Subject: newBinding.NatsSubjectPattern,
		},
		Batching:       nil,
		AssignedPeerID: assignedPeerID,
	}

	if newBinding.Batching != nil {
		resp.Batching = &repositories.BindingBatching{
			MaxMessages: newBinding.Batching.MaxMessages,
			MaxLatency:  newBinding.Batching.MaxLatency,
		}
	}

	return resp, nil
}

func (b *Bindings) GetJetstreamBinding(ctx context.Context, id uuid.UUID) (*repositories.JetstreamBinding, error) {
	eg, ctx := errgroup.WithContext(ctx)

	var (
		peerIDs []string
		binding *repositories.JetstreamBinding
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
			Get("pk", &jetstreamBindingPK{}).
			Range("sk", dynamo.Equal, id)

		var jetstreamBinding jetstreamBindingRecord
		if err := query.OneWithContext(ctx, &jetstreamBinding); err != nil {
			return err
		}

		binding = &repositories.JetstreamBinding{
			ID:        jetstreamBinding.ID,
			LambdaARN: jetstreamBinding.LambdaARN,
			Consumer: repositories.JetstreamConsumer{
				Stream:  jetstreamBinding.NatsStream,
				Name:    jetstreamBinding.NatsConsumer,
				Subject: jetstreamBinding.NatsSubjectPattern,
			},
			Batching: nil,
		}

		if jetstreamBinding.Batching == nil {
			binding.Batching = &repositories.BindingBatching{
				MaxMessages: jetstreamBinding.Batching.MaxMessages,
				MaxLatency:  jetstreamBinding.Batching.MaxLatency,
			}
		}

		return nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	// Determine the assigned peer using rendezvous hashing
	if len(peerIDs) > 0 {
		h := rendezvous.NewHasher(rendezvous.WithMembers(peerIDs...))

		assignedPeerID, err := uuid.Parse(h.Owner(binding.ID.String()))
		if err != nil {
			return nil, err
		}

		binding.AssignedPeerID = &assignedPeerID
	}

	return binding, nil
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
				ID:        binding.ID,
				LambdaARN: binding.LambdaARN,
				Consumer: repositories.JetstreamConsumer{
					Stream:  binding.NatsStream,
					Name:    binding.NatsConsumer,
					Subject: binding.NatsSubjectPattern,
				},
				Batching: &repositories.BindingBatching{
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

func (b *Bindings) DeleteJetstreamBinding(ctx context.Context, id uuid.UUID) error {
	query := b.db.Table(b.tableName).
		Delete("pk", &jetstreamBindingPK{}).
		Range("sk", id)

	return query.RunWithContext(ctx)
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
