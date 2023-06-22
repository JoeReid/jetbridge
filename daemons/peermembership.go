package daemons

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/JoeReid/jetbridge/repositories"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

type PeerMembership struct {
	id      *uuid.UUID
	eg      *errgroup.Group
	joinErr error
}

func (p *PeerMembership) Go(f func(peerID uuid.UUID) error) {
	p.eg.Go(func() error {
		if p.id == nil {
			return fmt.Errorf("peerID is nil")
		}

		return f(*p.id)
	})
}

func (p *PeerMembership) Wait() error {
	if err := p.eg.Wait(); p.joinErr != nil {
		return p.joinErr
	} else {
		return err
	}
}

func NewPeerMembership(parent context.Context, peers repositories.Peers) (*PeerMembership, context.Context) {
	// Join the cluster
	peer, err := peers.JoinPeers(parent)
	if err != nil {
		ctx, cancel := context.WithCancel(parent)
		eg, ctx := errgroup.WithContext(ctx)
		cancel()

		return &PeerMembership{
			id:      nil,
			eg:      eg,
			joinErr: err,
		}, ctx
	}

	// Create a new context and errgroup managing child goroutines
	eg, ctx := errgroup.WithContext(parent)

	// Maintain our membership in background using the errgroup.
	// This means that if the membership fails, the whole process fails
	// and cleans up any child goroutines that have been added.
	eg.Go(func() error {
		dueBy := peer.HeartbeatDueBy

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()

			case <-time.After(time.Until(dueBy) / 2):
				updatedPeer, err := peers.SendHeartbeat(ctx, peer.ID)
				if err != nil {
					return err
				}
				dueBy = updatedPeer.HeartbeatDueBy
			}
		}
	})

	// Register a function to leave the cluster when the context is cancelled
	eg.Go(func() error {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		err := peers.LeavePeers(ctx, peer.ID)
		if err != nil {
			log.Printf("failed to leave cluster (peerID=%s): %v", peer.ID.String(), err)
		}
		return err
	})

	return &PeerMembership{
		id:      &peer.ID,
		eg:      eg,
		joinErr: nil,
	}, ctx
}
