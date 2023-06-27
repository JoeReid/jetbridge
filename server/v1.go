package server

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	v1 "github.com/JoeReid/jetbridge/proto/gen/go/jetbridge/v1"
	"github.com/JoeReid/jetbridge/proto/gen/go/jetbridge/v1/v1connect"
	"github.com/JoeReid/jetbridge/repositories"
	"github.com/bufbuild/connect-go"
	"github.com/google/uuid"
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

	var deliveryPolicy string
	switch req.Msg.DeliveryPolicy.(type) {
	case *v1.CreateBindingRequest_Policy:
		deliveryPolicy = req.Msg.GetPolicy()
	case *v1.CreateBindingRequest_StartTime:
		deliveryPolicy = req.Msg.GetStartTime().AsTime().Format(time.RFC3339)
	case *v1.CreateBindingRequest_StartSequence:
		deliveryPolicy = fmt.Sprintf("%d", req.Msg.GetStartSequence())
	default:
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid delivery policy"))
	}

	binding, err := v.Bindings.CreateJetstreamBinding(ctx, &repositories.CreateJetstreamBinding{
		LambdaARN:      req.Msg.LambdaArn,
		Stream:         req.Msg.Stream,
		Subject:        req.Msg.SubjectPattern,
		MaxMessages:    int(req.Msg.MaxBatchSize),
		MaxLatency:     req.Msg.MaxBatchLatency.AsDuration(),
		DeliveryPolicy: deliveryPolicy,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	v1Binding := &v1.JetstreamBinding{
		Id:              binding.ID.String(),
		LambdaArn:       binding.LambdaARN,
		Stream:          binding.Stream,
		ConsumerName:    binding.Consumer.String(),
		SubjectPattern:  binding.Subject,
		MaxBatchSize:    int64(binding.MaxMessages),
		MaxBatchLatency: durationpb.New(binding.MaxLatency),
	}

	switch req.Msg.DeliveryPolicy.(type) {
	case *v1.CreateBindingRequest_Policy:
		v1Binding.DeliveryPolicy = &v1.JetstreamBinding_Policy{
			Policy: req.Msg.GetPolicy(),
		}
	case *v1.CreateBindingRequest_StartTime:
		v1Binding.DeliveryPolicy = &v1.JetstreamBinding_StartTime{
			StartTime: req.Msg.GetStartTime(),
		}
	case *v1.CreateBindingRequest_StartSequence:
		v1Binding.DeliveryPolicy = &v1.JetstreamBinding_StartSequence{
			StartSequence: req.Msg.GetStartSequence(),
		}
	default:
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid delivery policy"))
	}

	if binding.AssignedPeerID != nil {
		v1Binding.AssignedPeer = binding.AssignedPeerID.String()
	}

	return connect.NewResponse(&v1.CreateBindingResponse{Binding: v1Binding}), nil
}

func (v *V1) GetBinding(ctx context.Context, req *connect.Request[v1.GetBindingRequest]) (*connect.Response[v1.GetBindingResponse], error) {
	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	id, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	binding, err := v.Bindings.GetJetstreamBinding(ctx, id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	v1Binding := &v1.JetstreamBinding{
		Id:              binding.ID.String(),
		LambdaArn:       binding.LambdaARN,
		Stream:          binding.Stream,
		ConsumerName:    binding.Consumer.String(),
		SubjectPattern:  binding.Subject,
		MaxBatchSize:    int64(binding.MaxMessages),
		MaxBatchLatency: durationpb.New(binding.MaxLatency),
	}

	switch binding.DeliveryPolicy {
	case "all", "last", "last-per-subject", "new":
		v1Binding.DeliveryPolicy = &v1.JetstreamBinding_Policy{
			Policy: binding.DeliveryPolicy,
		}

	default:
		if t, err := time.Parse(time.RFC3339, binding.DeliveryPolicy); err == nil {
			v1Binding.DeliveryPolicy = &v1.JetstreamBinding_StartTime{
				StartTime: timestamppb.New(t),
			}
		}

		if i, err := strconv.ParseUint(binding.DeliveryPolicy, 10, 64); err == nil {
			v1Binding.DeliveryPolicy = &v1.JetstreamBinding_StartSequence{
				StartSequence: i,
			}
		}

		return nil, connect.NewError(connect.CodeInternal, errors.New("invalid delivery policy"))
	}

	if binding.AssignedPeerID != nil {
		v1Binding.AssignedPeer = binding.AssignedPeerID.String()
	}

	return connect.NewResponse(&v1.GetBindingResponse{Binding: v1Binding}), nil
}

func (v *V1) ListBindings(ctx context.Context, req *connect.Request[v1.ListBindingsRequest]) (*connect.Response[v1.ListBindingsResponse], error) {
	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	bindings, err := v.Bindings.ListJetstreamBindings(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	var v1Bindings []*v1.JetstreamBinding
	for _, binding := range bindings {
		v1Binding := &v1.JetstreamBinding{
			Id:              binding.ID.String(),
			LambdaArn:       binding.LambdaARN,
			Stream:          binding.Stream,
			ConsumerName:    binding.Consumer.String(),
			SubjectPattern:  binding.Subject,
			MaxBatchSize:    int64(binding.MaxMessages),
			MaxBatchLatency: durationpb.New(binding.MaxLatency),
		}

		switch binding.DeliveryPolicy {
		case "all", "last", "last-per-subject", "new":
			v1Binding.DeliveryPolicy = &v1.JetstreamBinding_Policy{
				Policy: binding.DeliveryPolicy,
			}

		default:
			if t, err := time.Parse(time.RFC3339, binding.DeliveryPolicy); err == nil {
				v1Binding.DeliveryPolicy = &v1.JetstreamBinding_StartTime{
					StartTime: timestamppb.New(t),
				}
			}

			if i, err := strconv.ParseUint(binding.DeliveryPolicy, 10, 64); err == nil {
				v1Binding.DeliveryPolicy = &v1.JetstreamBinding_StartSequence{
					StartSequence: i,
				}
			}

			return nil, connect.NewError(connect.CodeInternal, errors.New("invalid delivery policy"))
		}

		if binding.AssignedPeerID != nil {
			v1Binding.AssignedPeer = binding.AssignedPeerID.String()
		}

		v1Bindings = append(v1Bindings, v1Binding)
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

	id, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	if err := v.Bindings.DeleteJetstreamBinding(ctx, id); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&v1.DeleteBindingResponse{}), nil
}
