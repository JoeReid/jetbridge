package nats

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"sync"
	"time"

	"github.com/JoeReid/jetbridge"
	"github.com/JoeReid/jetbridge/repositories"
	"github.com/nats-io/nats.go"
)

var _ repositories.MessageSource = (*MessageSource)(nil)

func NewMessageSource(js nats.JetStreamContext) *MessageSource {
	return &MessageSource{
		js:            js,
		mu:            &sync.Mutex{},
		subscriptions: make(map[string]*nats.Subscription),
	}
}

type MessageSource struct {
	js nats.JetStreamContext

	mu            *sync.Mutex
	subscriptions map[string]*nats.Subscription
}

func (m *MessageSource) FetchJetstreamMessages(ctx context.Context, binding repositories.JetstreamBinding) ([]repositories.JetstreamMessage, error) {
	var (
		batchSize    = 1
		batchLatency = 30 * time.Second
	)

	if binding.Batching != nil {
		batchSize = binding.Batching.MaxMessages
		batchLatency = binding.Batching.MaxLatency
	}

	sub, err := m.subscription(ctx, binding)
	if err != nil {
		return nil, err
	}

	fetchCtx, cancel := context.WithTimeout(ctx, batchLatency)
	defer cancel()

	msgs, err := sub.Fetch(batchSize, nats.Context(fetchCtx))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %w", err)
	}

	return newMessages(msgs)
}

func (m *MessageSource) subscription(ctx context.Context, binding repositories.JetstreamBinding) (*nats.Subscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if sub, ok := m.subscriptions[binding.ID.String()]; ok {
		return sub, nil
	}

	desiredConfig := &nats.ConsumerConfig{
		Durable:           binding.NatsConsumer,
		Name:              "",
		Description:       fmt.Sprintf("JetBridge Lambda consumer for %s", binding.LambdaARN),
		DeliverPolicy:     nats.DeliverAllPolicy, // TODO: does this need exposing in the binding?
		AckPolicy:         nats.AckAllPolicy,
		AckWait:           time.Minute, // TODO: does this need exposing in the binding? Can we infer it from lambda timeout?
		MaxDeliver:        -1,          // TODO: does this need exposing in the binding?
		FilterSubject:     "",          // TODO: this really needs exposing in the binding
		ReplayPolicy:      nats.ReplayInstantPolicy,
		MaxWaiting:        1, // Only one worker should be processing a message at a time (in most cases), we may as well ask NATS to enforce this
		MaxAckPending:     1,
		FlowControl:       false, // TODO: is this right? what are the implications of this?
		MaxRequestBatch:   1,
		MaxRequestExpires: time.Minute,
	}

	if binding.Batching != nil {
		desiredConfig.MaxAckPending = binding.Batching.MaxMessages
		desiredConfig.MaxRequestBatch = binding.Batching.MaxMessages
		desiredConfig.MaxRequestExpires = binding.Batching.MaxLatency
	}

	infoCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	info, err := m.js.ConsumerInfo(binding.NatsStream, binding.NatsConsumer, nats.Context(infoCtx))
	switch {
	case errors.Is(err, nats.ErrConsumerNotFound):
		createCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if _, err := m.js.AddConsumer(binding.NatsStream, desiredConfig, nats.Context(createCtx)); err != nil {
			return nil, fmt.Errorf("failed to create consumer: %w", err)
		}

	case err != nil:
		return nil, fmt.Errorf("failed to get consumer info: %w", err)

	default:
		if !reflect.DeepEqual(info.Config, *desiredConfig) {
			return nil, errors.New("consumer info does not match binding")
		}
	}

	sub, err := m.js.PullSubscribe(binding.NatsSubjectPattern, binding.NatsConsumer, nats.Bind(binding.NatsStream, binding.NatsConsumer))
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to jetstream binding: %w", err)
	}

	m.subscriptions[binding.ID.String()] = sub
	return sub, nil
}

type Message struct {
	md  *nats.MsgMetadata
	msg *nats.Msg
}

func (m *Message) Payload() jetbridge.JetstreamLambdaPayload {
	return jetbridge.JetstreamLambdaPayload{
		Subject:  m.msg.Subject,
		Header:   m.msg.Header,
		Data:     m.msg.Data,
		Metadata: *m.md,
	}
}

func (m *Message) Ack() error {
	return m.msg.Ack() // TODO: add necessary options
}

func (m *Message) Nak() error {
	return m.msg.Nak() // TODO: add necessary options
}

func newMessages(msgs []*nats.Msg) ([]repositories.JetstreamMessage, error) {
	cleanup := func() {
		for _, msg := range msgs {
			if err := msg.Nak(); err != nil {
				log.Printf("failed to nack message: %v", err)
			}
		}
	}

	var messages []repositories.JetstreamMessage
	for _, msg := range msgs {
		md, err := msg.Metadata()
		if err != nil {
			defer cleanup()
			return nil, fmt.Errorf("failed to get message metadata: %w", err)
		}

		messages = append(messages, &Message{
			md:  md,
			msg: msg,
		})
	}

	return messages, nil
}
