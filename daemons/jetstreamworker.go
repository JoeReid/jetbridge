package daemons

import (
	"context"
	"log"
	"time"

	"github.com/JoeReid/jetbridge/repositories"
	"github.com/google/uuid"
	"golang.org/x/exp/slices"
	"golang.org/x/sync/errgroup"
)

type JetstreamWorker struct {
	Bindings       repositories.Bindings
	Messages       repositories.MessageSource
	Handler        repositories.MessageHandler
	UpdateInterval time.Duration
}

func (j *JetstreamWorker) Run(ctx context.Context, peerID uuid.UUID) error {
	eg, ctx := errgroup.WithContext(ctx)

	ch := make(chan []repositories.JetstreamBinding)

	eg.Go(func() error {
		return j.pollBindings(ctx, peerID, ch)
	})

	eg.Go(func() error {
		return j.scheduleBindings(ctx, ch)
	})

	return eg.Wait()
}

func (j *JetstreamWorker) pollBindings(ctx context.Context, peerID uuid.UUID, ch chan []repositories.JetstreamBinding) error {
	defer close(ch)

	ticker := time.NewTicker(j.UpdateInterval)
	defer ticker.Stop()

	var lastBindings []repositories.JetstreamBinding
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ticker.C:
			allBindings, err := j.Bindings.ListJetstreamBindings(ctx)
			if err != nil {
				return err
			}

			var filteredBindings []repositories.JetstreamBinding
			for _, binding := range allBindings {
				if binding.AssignedPeerID != nil && binding.AssignedPeerID.String() == peerID.String() {
					filteredBindings = append(filteredBindings, binding)
				}
			}

			if !slices.Equal(lastBindings, filteredBindings) {
				lastBindings = filteredBindings
				ch <- lastBindings
			}
		}
	}
}

func (j *JetstreamWorker) scheduleBindings(ctx context.Context, ch chan []repositories.JetstreamBinding) error {
	var (
		workerCtx context.Context
		cancel    context.CancelFunc
	)

	for {
		select {
		case <-ctx.Done():
			defer func() {
				if cancel != nil {
					cancel()
				}
			}()

			return ctx.Err()

		case bindings := <-ch:
			if cancel != nil {
				cancel()
			}

			workerCtx, cancel = context.WithCancel(ctx)
			for _, binding := range bindings {
				go j.runBinding(workerCtx, binding)
			}
		}
	}
}

func (j *JetstreamWorker) runBinding(ctx context.Context, binding repositories.JetstreamBinding) {
	for {
		select {
		case <-ctx.Done():
			return

		default:
			messages, err := j.Messages.FetchJetstreamMessages(ctx, binding)
			if err != nil {
				log.Println("error fetching messages", err)
			}

			if err := j.Handler.HandleJetstreamMessages(ctx, binding, messages); err != nil {
				log.Println("error handling messages", err)
			}
		}
	}
}
