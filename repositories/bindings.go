package repositories

import (
	"context"
)

//go:generate go run github.com/golang/mock/mockgen -destination=./mocks/mock_bindings.go -package=mocks . Bindings

type Bindings interface {
	ListJetstreamBindings(ctx context.Context) ([]JetstreamBinding, error)
}
