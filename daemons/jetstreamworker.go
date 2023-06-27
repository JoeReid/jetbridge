package daemons

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/JoeReid/jetbridge/repositories"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
)

type jetstreamWorkerBinding struct {
	binding repositories.JetstreamBinding
	cancel  context.CancelFunc
}

func NewJetstreamWorker(bindings repositories.Bindings, messages repositories.MessageSource, handler repositories.MessageHandler) (*JetstreamWorker, error) {
	const (
		defaultUpdateInterval = 5 * time.Second
	)

	zl, err := zap.NewDevelopment() // TODO: this needs to be managed better
	if err != nil {
		return nil, err
	}

	return &JetstreamWorker{
		logger:         zl,
		bindings:       bindings,
		messages:       messages,
		handler:        handler,
		updateInterval: defaultUpdateInterval,
		mu:             &sync.Mutex{},
		workers:        make(map[string]jetstreamWorkerBinding),
	}, nil
}

type JetstreamWorker struct {
	logger         *zap.Logger
	bindings       repositories.Bindings
	messages       repositories.MessageSource
	handler        repositories.MessageHandler
	updateInterval time.Duration

	mu      *sync.Mutex
	workers map[string]jetstreamWorkerBinding
}

func (j *JetstreamWorker) Run(ctx context.Context, peerID uuid.UUID) error {
	ticker := time.NewTicker(j.updateInterval)
	defer ticker.Stop()

	var lastBindings []repositories.JetstreamBinding
	for {
		select {
		case <-ctx.Done():
			for _, binding := range lastBindings {
				j.removeBinding(ctx, binding)
			}
			return ctx.Err()

		case <-ticker.C:
			bindings, err := j.bindings.ListJetstreamBindings(ctx)
			if err != nil {
				j.logger.Error("error listing bindings", zap.Error(err))
				continue
			}

			var filteredBindings []repositories.JetstreamBinding
			for _, binding := range bindings {
				j.logger.Debug("checking binding", zap.Any("binding", binding))

				if binding.AssignedPeerID != nil && binding.AssignedPeerID.String() == peerID.String() {
					j.logger.Debug("binding matches peer", zap.Any("binding", binding), zap.String("peer_id", peerID.String()))
					filteredBindings = append(filteredBindings, binding)
				} else {
					j.logger.Debug("binding does not match peer", zap.Any("binding", binding), zap.String("peer_id", peerID.String()))
				}
			}

			// remove bindings that are no longer assigned to this peer
			for _, binding := range lastBindings {
				if !slices.ContainsFunc(filteredBindings, func(b repositories.JetstreamBinding) bool { return b.ID.String() == binding.ID.String() }) {
					j.removeBinding(ctx, binding)
				}
			}

			// add bindings that are now assigned to this peer
			for _, binding := range filteredBindings {
				if !slices.ContainsFunc(lastBindings, func(b repositories.JetstreamBinding) bool { return b.ID.String() == binding.ID.String() }) {
					j.addBinding(ctx, binding)
				}
			}

			lastBindings = filteredBindings
		}
	}
}

func (j *JetstreamWorker) runBinding(ctx context.Context, binding repositories.JetstreamBinding) {
	j.logger.Info(
		"running binding",
		zap.String("binding_id", binding.ID.String()),
	)

	for {
		select {
		case <-ctx.Done():
			return

		default:
			j.logger.Info(
				"fetching messages for binding",
				zap.String("binding_id", binding.ID.String()),
			)

			messages, err := j.messages.FetchJetstreamMessages(ctx, binding)
			if err != nil {
				log.Println("error fetching messages", err)
			}

			j.logger.Info(
				"handling messages for binding",
				zap.String("binding_id", binding.ID.String()),
				zap.Int("messages", len(messages)),
			)

			if err := j.handler.HandleJetstreamMessages(ctx, binding, messages); err != nil {
				log.Println("error handling messages", err)
			}
		}
	}
}

func (j *JetstreamWorker) addBinding(ctx context.Context, binding repositories.JetstreamBinding) {
	j.mu.Lock()
	defer j.mu.Unlock()

	if _, ok := j.workers[binding.ID.String()]; ok {
		panic("binding already exists")
	}

	ctx, cancel := context.WithCancel(ctx)
	j.workers[binding.ID.String()] = jetstreamWorkerBinding{
		binding: binding,
		cancel:  cancel,
	}

	go j.runBinding(ctx, binding)
}

func (j *JetstreamWorker) removeBinding(ctx context.Context, binding repositories.JetstreamBinding) {
	j.mu.Lock()
	defer j.mu.Unlock()

	if b, ok := j.workers[binding.ID.String()]; ok {
		b.cancel()
		delete(j.workers, binding.ID.String())
		return
	}

	panic("binding does not exist")
}
