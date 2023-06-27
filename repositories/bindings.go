package repositories

import (
	"context"

	"github.com/google/uuid"
)

//go:generate go run github.com/golang/mock/mockgen -destination=./mocks/mock_bindings.go -package=mocks . Bindings

type Bindings interface {
	CreateJetstreamBinding(context.Context, *CreateJetstreamBinding) (*JetstreamBinding, error)
	GetJetstreamBinding(ctx context.Context, id uuid.UUID) (*JetstreamBinding, error)
	ListJetstreamBindings(ctx context.Context) ([]JetstreamBinding, error)
	DeleteJetstreamBinding(ctx context.Context, id uuid.UUID) error
}
