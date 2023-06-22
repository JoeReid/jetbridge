package repositories

import (
	"context"
)

//go:generate go run github.com/golang/mock/mockgen -destination=./mocks/mock_messagesource.go -package=mocks . MessageSource

type MessageSource interface {
	FetchJetstreamMessages(ctx context.Context, binding JetstreamBinding) ([]JetstreamMessage, error)
}
