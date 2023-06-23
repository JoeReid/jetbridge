package nats

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/JoeReid/jetbridge/repositories"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessageSource_FetchJetstreamMessages(t *testing.T) {
	js := testingNATS(t)

	_, err := js.AddStream(&nats.StreamConfig{
		Name:     "TESTSTREAM",
		Subjects: []string{"TESTSTREAM.*"},
	}, nats.MaxWait(5*time.Second))
	require.NoError(t, err)

	_, err = js.Publish("TESTSTREAM.1", []byte("test message 1"))
	require.NoError(t, err)

	_, err = js.Publish("TESTSTREAM.2", []byte("test message 2"))
	require.NoError(t, err)

	_, err = js.Publish("TESTSTREAM.3", []byte("test message 3"))
	require.NoError(t, err)

	_, err = js.Publish("TESTSTREAM.4", []byte("test message 4"))
	require.NoError(t, err)

	candidate := &MessageSource{
		js:            js,
		mu:            &sync.Mutex{},
		subscriptions: make(map[string]*nats.Subscription),
	}

	id := uuid.New()

	t.Run("consumer not exist", func(t *testing.T) {
		msgs, err := candidate.FetchJetstreamMessages(context.TODO(), repositories.JetstreamBinding{
			ID:        id,
			LambdaARN: "test-arn",
			Consumer: repositories.JetstreamConsumer{
				Stream:  "TESTSTREAM",
				Name:    id.String(),
				Subject: "TESTSTREAM.*",
			},
			Batching: &repositories.BindingBatching{
				MaxMessages: 2,
				MaxLatency:  time.Second,
			},
		})
		assert.NoError(t, err)
		assert.Len(t, msgs, 2)

		for _, msg := range msgs {
			assert.NoError(t, msg.Ack())
		}
	})

	t.Run("consumer exists", func(t *testing.T) {
		msgs, err := candidate.FetchJetstreamMessages(context.TODO(), repositories.JetstreamBinding{
			ID:        id,
			LambdaARN: "test-arn",
			Consumer: repositories.JetstreamConsumer{
				Stream:  "TESTSTREAM",
				Name:    id.String(),
				Subject: "TESTSTREAM.*",
			},
			Batching: &repositories.BindingBatching{
				MaxMessages: 2,
				MaxLatency:  time.Second,
			},
		})
		assert.NoError(t, err)
		assert.Len(t, msgs, 2)

		for _, msg := range msgs {
			assert.NoError(t, msg.Ack())
		}
	})
}

func testingNATS(t *testing.T) nats.JetStreamContext {
	t.Helper()

	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("Could construct docker pool: %s", err)
	}

	if err := pool.Client.Ping(); err != nil {
		t.Fatalf("Could not connect to Docker: %s", err)
	}

	natsContainer, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "nats",
		Tag:        "latest",
		Cmd:        []string{"-js"},
	})
	if err != nil {
		t.Fatalf("Could not start NATS container: %s", err)
	}

	if err := pool.Retry(func() error {
		nc, err := nats.Connect(natsContainer.GetHostPort("4222/tcp"))
		if err != nil {
			return err
		}
		defer nc.Close()

		return nil
	}); err != nil {
		t.Fatalf("Could not connect to NATS: %s", err)
	}

	nc, err := nats.Connect(natsContainer.GetHostPort("4222/tcp"))
	if err != nil {
		t.Fatalf("Could not connect to NATS: %s", err)
	}
	t.Cleanup(nc.Close)

	js, err := nc.JetStream()
	if err != nil {
		t.Fatalf("Could not connect to JetStream: %s", err)
	}

	return js
}
