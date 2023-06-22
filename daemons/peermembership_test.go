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

func TestPeerMembership_join(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	id := uuid.New()

	peers := mocks.NewMockPeers(ctrl)
	peers.EXPECT().JoinPeers(gomock.Any()).Return(&repositories.Peer{
		ID:             id,
		Hostname:       "test",
		JoinedAt:       time.Now(),
		LastSeenAt:     time.Now(),
		HeartbeatDueBy: time.Now().Add(5 * time.Second),
	}, nil)
	peers.EXPECT().LeavePeers(gomock.Any(), id).Return(nil)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	candidate, _ := NewPeerMembership(ctx, peers)

	assert.ErrorContains(t, candidate.Wait(), "context deadline exceeded")
}

func TestPeerMembership_joinFailure(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	peers := mocks.NewMockPeers(ctrl)
	peers.EXPECT().JoinPeers(gomock.Any()).Return(nil, assert.AnError)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	candidate, _ := NewPeerMembership(ctx, peers)

	assert.ErrorIs(t, candidate.Wait(), assert.AnError)
}

func TestPeerMembership_heartbeat(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	id := uuid.New()

	peers := mocks.NewMockPeers(ctrl)
	peers.EXPECT().JoinPeers(gomock.Any()).Return(&repositories.Peer{
		ID:             id,
		Hostname:       "test",
		JoinedAt:       time.Now(),
		LastSeenAt:     time.Now(),
		HeartbeatDueBy: time.Now().Add(50 * time.Millisecond),
	}, nil)
	peers.EXPECT().SendHeartbeat(gomock.Any(), id).Return(&repositories.Peer{
		ID:             id,
		Hostname:       "test",
		JoinedAt:       time.Now(),
		LastSeenAt:     time.Now(),
		HeartbeatDueBy: time.Now().Add(100 * time.Millisecond),
	}, nil)
	peers.EXPECT().SendHeartbeat(gomock.Any(), id).Return(&repositories.Peer{
		ID:             id,
		Hostname:       "test",
		JoinedAt:       time.Now(),
		LastSeenAt:     time.Now(),
		HeartbeatDueBy: time.Now().Add(5 * time.Second),
	}, nil)
	peers.EXPECT().LeavePeers(gomock.Any(), id).Return(nil)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	candidate, _ := NewPeerMembership(ctx, peers)

	assert.ErrorContains(t, candidate.Wait(), "context deadline exceeded")
}

func TestPeerMembership_heartbeatFailure(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	id := uuid.New()

	peers := mocks.NewMockPeers(ctrl)
	peers.EXPECT().JoinPeers(gomock.Any()).Return(&repositories.Peer{
		ID:             id,
		Hostname:       "test",
		JoinedAt:       time.Now(),
		LastSeenAt:     time.Now(),
		HeartbeatDueBy: time.Now().Add(50 * time.Millisecond),
	}, nil)
	peers.EXPECT().SendHeartbeat(gomock.Any(), id).Return(nil, assert.AnError)
	peers.EXPECT().LeavePeers(gomock.Any(), id).Return(nil)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	candidate, _ := NewPeerMembership(ctx, peers)

	assert.ErrorIs(t, candidate.Wait(), assert.AnError)
}

func TestPeerMembership_workers(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	id := uuid.New()

	peers := mocks.NewMockPeers(ctrl)
	peers.EXPECT().JoinPeers(gomock.Any()).Return(&repositories.Peer{
		ID:             id,
		Hostname:       "test",
		JoinedAt:       time.Now(),
		LastSeenAt:     time.Now(),
		HeartbeatDueBy: time.Now().Add(5 * time.Second),
	}, nil)
	peers.EXPECT().LeavePeers(gomock.Any(), id).Return(nil)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	candidate, ctx := NewPeerMembership(ctx, peers)

	var (
		called int
		exited int
	)

	go func() {
		time.Sleep(50 * time.Millisecond)

		for i := 0; i < 3; i++ {
			candidate.Go(func(peerID uuid.UUID) error {
				called++
				defer func() { exited++ }()

				assert.Equal(t, id, peerID)

				<-ctx.Done()
				return nil
			})
		}
	}()

	assert.ErrorContains(t, candidate.Wait(), "context deadline exceeded")
	assert.Equal(t, 3, called)
	assert.Equal(t, 3, exited)
}

func TestPeerMembership_workerFailure(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	id := uuid.New()

	peers := mocks.NewMockPeers(ctrl)
	peers.EXPECT().JoinPeers(gomock.Any()).Return(&repositories.Peer{
		ID:             id,
		Hostname:       "test",
		JoinedAt:       time.Now(),
		LastSeenAt:     time.Now(),
		HeartbeatDueBy: time.Now().Add(5 * time.Second),
	}, nil)
	peers.EXPECT().LeavePeers(gomock.Any(), id).Return(nil)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	candidate, ctx := NewPeerMembership(ctx, peers)

	var (
		called int
		exited int
	)

	go func() {
		time.Sleep(50 * time.Millisecond)

		for i := 0; i < 3; i++ {
			candidate.Go(func(peerID uuid.UUID) error {
				called++
				defer func() { exited++ }()

				assert.Equal(t, id, peerID)

				<-ctx.Done()
				return nil
			})
		}

		candidate.Go(func(peerID uuid.UUID) error {
			called++
			defer func() { exited++ }()

			assert.Equal(t, id, peerID)

			return assert.AnError
		})
	}()

	assert.ErrorIs(t, candidate.Wait(), assert.AnError)
	assert.Equal(t, 4, called)
	assert.Equal(t, 4, exited)
}
