package admin

import (
	"context"

	"github.com/EonsofStupid/tessera/internal/api/authz"
	"github.com/EonsofStupid/tessera/internal/api/grpc/object"
	policy_grpc "github.com/EonsofStupid/tessera/internal/api/grpc/policy"
	admin_pb "github.com/EonsofStupid/tessera/pkg/grpc/admin"
)

func (s *Server) AddNotificationPolicy(ctx context.Context, req *admin_pb.AddNotificationPolicyRequest) (*admin_pb.AddNotificationPolicyResponse, error) {
	result, err := s.command.AddDefaultNotificationPolicy(ctx, authz.GetInstance(ctx).InstanceID(), req.GetPasswordChange())
	if err != nil {
		return nil, err
	}
	return &admin_pb.AddNotificationPolicyResponse{
		Details: object.AddToDetailsPb(
			result.Sequence,
			result.EventDate,
			result.ResourceOwner,
		),
	}, nil
}

func (s *Server) GetNotificationPolicy(ctx context.Context, _ *admin_pb.GetNotificationPolicyRequest) (*admin_pb.GetNotificationPolicyResponse, error) {
	policy, err := s.query.DefaultNotificationPolicy(ctx, true)
	if err != nil {
		return nil, err
	}
	return &admin_pb.GetNotificationPolicyResponse{Policy: policy_grpc.ModelNotificationPolicyToPb(policy)}, nil
}

func (s *Server) UpdateNotificationPolicy(ctx context.Context, req *admin_pb.UpdateNotificationPolicyRequest) (*admin_pb.UpdateNotificationPolicyResponse, error) {
	result, err := s.command.ChangeDefaultNotificationPolicy(ctx, authz.GetInstance(ctx).InstanceID(), req.GetPasswordChange())
	if err != nil {
		return nil, err
	}
	return &admin_pb.UpdateNotificationPolicyResponse{
		Details: object.ChangeToDetailsPb(
			result.Sequence,
			result.EventDate,
			result.ResourceOwner,
		),
	}, nil
}
