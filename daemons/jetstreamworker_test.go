package daemons

import (
	"context"
	"testing"
	"time"

	"github.com/JoeReid/jetbridge/repositories"
	"github.com/JoeReid/jetbridge/repositories/mocks"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJetstreamWorker(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var (
		peerID    = uuid.New()
		bindingID = uuid.New()
	)

	bindings := mocks.NewMockBindings(ctrl)
	bindings.EXPECT().ListJetstreamBindings(gomock.Any()).Return([]repositories.JetstreamBinding{
		{
			ID:             bindingID,
			LambdaARN:      "test-arn",
			Stream:         "test-stream",
			Consumer:       bindingID,
			Subject:        "test-stream.*",
			MaxMessages:    10,
			MaxLatency:     time.Second,
			AssignedPeerID: &peerID,
		},
	}, nil).Times(4)

	msg := mocks.NewMockJetstreamMessage(ctrl)

	source := mocks.NewMockMessageSource(ctrl)
	source.EXPECT().FetchJetstreamMessages(gomock.Any(), gomock.Any()).Return([]repositories.JetstreamMessage{msg}, nil).AnyTimes()

	handler := mocks.NewMockMessageHandler(ctrl)
	handler.EXPECT().HandleJetstreamMessages(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	candidate, err := NewJetstreamWorker(bindings, source, handler)
	require.NoError(t, err)
	candidate.updateInterval = time.Second

	ctx, cancel := context.WithTimeout(context.TODO(), time.Millisecond*4500) // 4.5 seconds, not enough time to run a full 5 times
	defer cancel()

	err = candidate.Run(ctx, peerID)
	assert.ErrorContains(t, err, "context deadline exceeded")
}
