package repositories

import (
	"context"
)

//go:generate go run github.com/golang/mock/mockgen -destination=./mocks/mock_messagehandler.go -package=mocks . MessageHandler

// MessageHandler defines the interface for a repository that can
// execute lambdas for jetstream messages.
//
// Implementations of this interface should fully handle the lifecycle of the
// message, including ACK-ing and NACK-ing the message as appropriate.
//
// Implementations of this interface must implement methods in a blocking
// manner, and should not return until the message has been fully processed.
//
// Implementations of this interface should be safe for concurrent use by
// multiple goroutines.
type MessageHandler interface {
	HandleJetstreamMessages(ctx context.Context, binding JetstreamBinding, messages []JetstreamMessage) error
}
