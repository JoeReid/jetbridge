package dynamo

import (
	"context"
	"fmt"
	"time"

	"github.com/JoeReid/jetbridge/repositories"
	"github.com/google/uuid"
	"github.com/guregu/dynamo"
)

var _ repositories.Bindings = (*Bindings)(nil)

type Bindings struct {
	db        *dynamo.DB
	tableName string
}

func (b *Bindings) CreateJetstreamBinding(ctx context.Context, create *repositories.CreateJetstreamBinding) (*repositories.JetstreamBinding, error) {
	record, err := newJetstreamBinding(create)
	if err != nil {
		return nil, err
	}

	createQuery := b.db.Table(b.tableName).
		Put(record).
		If("attribute_not_exists(pk)")

	peerQuery := b.db.Table(b.tableName).
		Get("pk", &peerPK{})

	if err := createQuery.RunWithContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to create jetstream binding: %w", err)
	}

	var peers []peerRecord
	if err := peerQuery.AllWithContext(ctx, &peers); err != nil {
		return nil, err
	}

	return record.toJetstreamBinding(peers), nil
}

func (b *Bindings) GetJetstreamBinding(ctx context.Context, id uuid.UUID) (*repositories.JetstreamBinding, error) {
	peerQuery := b.db.Table(b.tableName).
		Get("pk", &peerPK{})

	bindingQuery := b.db.Table(b.tableName).
		Get("pk", &jetstreamBindingPK{}).
		Range("sk", dynamo.Equal, id.String())

	var peers []peerRecord
	if err := peerQuery.AllWithContext(ctx, &peers); err != nil {
		return nil, err
	}

	var binding jetstreamBindingRecord
	if err := bindingQuery.OneWithContext(ctx, &binding); err != nil {
		return nil, err
	}

	return binding.toJetstreamBinding(peers), nil
}

func (b *Bindings) ListJetstreamBindings(ctx context.Context) ([]repositories.JetstreamBinding, error) {
	peerQuery := b.db.Table(b.tableName).
		Get("pk", &peerPK{})

	bindingQuery := b.db.Table(b.tableName).
		Get("pk", &jetstreamBindingPK{}).
		Index("created_at-index")

	var peers []peerRecord
	if err := peerQuery.AllWithContext(ctx, &peers); err != nil {
		return nil, err
	}

	var bindings jetstreamBindingRecords
	if err := bindingQuery.AllWithContext(ctx, &bindings); err != nil {
		return nil, err
	}

	return bindings.toJetstreamBindings(peers), nil
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
