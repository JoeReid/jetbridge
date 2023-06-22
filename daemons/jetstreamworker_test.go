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
)

func TestJetstreamWorker(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	id := uuid.New()

	bindings := mocks.NewMockBindings(ctrl)
	bindings.EXPECT().ListJetstreamBindings(gomock.Any()).Return([]repositories.JetstreamBinding{
		{
			ID:                 uuid.New(),
			NatsStream:         "test-stream",
			NatsConsumer:       "test-consumer",
			NatsSubjectPattern: "test-stream.*",
			LambdaARN:          "test-arn",
			Batching: &repositories.JetstreamBindingBatching{
				MaxMessages: 10,
				MaxLatency:  time.Second,
			},
			AssignedPeerID: &id,
		},
	}, nil).Times(4)

	msg := mocks.NewMockJetstreamMessage(ctrl)

	source := mocks.NewMockMessageSource(ctrl)
	source.EXPECT().FetchJetstreamMessages(gomock.Any(), gomock.Any()).Return([]repositories.JetstreamMessage{msg}, nil).AnyTimes()

	handler := mocks.NewMockMessageHandler(ctrl)
	handler.EXPECT().HandleJetstreamMessages(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	candidate := &JetstreamWorker{
		Bindings:       bindings,
		Messages:       source,
		Handler:        handler,
		UpdateInterval: time.Second,
	}

	ctx, cancel := context.WithTimeout(context.TODO(), time.Millisecond*4500) // 4.5 seconds, not enough time to run a full 5 times
	defer cancel()

	err := candidate.Run(ctx, id)
	assert.ErrorContains(t, err, "context deadline exceeded")
}
