package server

import (
	"context"
	"errors"

	v1 "github.com/JoeReid/jetbridge/proto/gen/go/jetbridge/v1"
	"github.com/JoeReid/jetbridge/proto/gen/go/jetbridge/v1/v1connect"
	"github.com/JoeReid/jetbridge/repositories"
	"github.com/bufbuild/connect-go"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ v1connect.JetbridgeServiceHandler = (*V1)(nil)

type V1 struct {
	v1connect.UnimplementedJetbridgeServiceHandler

	Bindings repositories.Bindings
	Peers    repositories.Peers
}

func (v *V1) ListPeers(ctx context.Context, req *connect.Request[v1.ListPeersRequest]) (*connect.Response[v1.ListPeersResponse], error) {
	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	peers, err := v.Peers.ListPeers(ctx)
	if err != nil {
		return nil, err // TODO: provide a better error
	}

	var v1Peers []*v1.Peer
	for _, peer := range peers {
		v1Peers = append(v1Peers, &v1.Peer{
			Id:           peer.ID.String(),
			Hostname:     peer.Hostname,
			Joined:       timestamppb.New(peer.JoinedAt),
			LastSeen:     timestamppb.New(peer.LastSeenAt),
			HeartbeatDue: timestamppb.New(peer.HeartbeatDueBy),
		})
	}

	resp := connect.NewResponse(&v1.ListPeersResponse{Peers: v1Peers})
	if err := resp.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return resp, nil
}

func (v *V1) CreateBinding(ctx context.Context, req *connect.Request[v1.CreateBindingRequest]) (*connect.Response[v1.CreateBindingResponse], error) {
	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("not implemented"))
}

func (v *V1) GetBinding(ctx context.Context, req *connect.Request[v1.GetBindingRequest]) (*connect.Response[v1.GetBindingResponse], error) {
	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("not implemented"))
}

func (v *V1) ListBindings(ctx context.Context, req *connect.Request[v1.ListBindingsRequest]) (*connect.Response[v1.ListBindingsResponse], error) {
	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	bindings, err := v.Bindings.ListJetstreamBindings(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	var v1Bindings []*v1.Binding
	for _, binding := range bindings {
		var batching *v1.BindingBatching
		if binding.Batching != nil {
			batching = &v1.BindingBatching{
				MaxMessages: int64(binding.Batching.MaxMessages),
				MaxLatency:  durationpb.New(binding.Batching.MaxLatency),
			}
		}

		v1Bindings = append(v1Bindings, &v1.Binding{
			Id:             binding.ID.String(),
			Stream:         binding.NatsStream,
			Consumer:       binding.NatsConsumer,
			SubjectPattern: binding.NatsSubjectPattern,
			LambdaArn:      binding.LambdaARN,
			Batching:       batching,
		})
	}

	resp := connect.NewResponse(&v1.ListBindingsResponse{Bindings: v1Bindings})
	if err := resp.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return resp, nil
}

func (v *V1) DeleteBinding(ctx context.Context, req *connect.Request[v1.DeleteBindingRequest]) (*connect.Response[v1.DeleteBindingResponse], error) {
	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("not implemented"))
}
