package auth

import (
	"context"

	"github.com/shippinAI/nomen/internal/api/authz"
	"github.com/shippinAI/nomen/internal/api/grpc/object"
	user_grpc "github.com/shippinAI/nomen/internal/api/grpc/user"
	"github.com/shippinAI/nomen/internal/query"
	auth_pb "github.com/shippinAI/nomen/pkg/grpc/auth"
)

func (s *Server) ListMyNomenPermissions(ctx context.Context, _ *auth_pb.ListMyNomenPermissionsRequest) (*auth_pb.ListMyNomenPermissionsResponse, error) {
	perms, err := s.query.MyNomenPermissions(ctx, authz.GetCtxData(ctx).OrgID, authz.GetCtxData(ctx).UserID)
	if err != nil {
		return nil, err
	}
	return &auth_pb.ListMyNomenPermissionsResponse{
		Result: perms.Permissions,
	}, nil
}

func (s *Server) ListMyProjectPermissions(ctx context.Context, _ *auth_pb.ListMyProjectPermissionsRequest) (*auth_pb.ListMyProjectPermissionsResponse, error) {
	ctxData := authz.GetCtxData(ctx)
	userGrantOrgID, err := query.NewUserGrantResourceOwnerSearchQuery(ctxData.OrgID)
	if err != nil {
		return nil, err
	}
	userGrantProjectID, err := query.NewUserGrantProjectIDSearchQuery(ctxData.ProjectID)
	if err != nil {
		return nil, err
	}
	userGrantUserID, err := query.NewUserGrantUserIDSearchQuery(ctxData.UserID)
	if err != nil {
		return nil, err
	}
	userGrant, err := s.query.UserGrant(ctx, true, userGrantOrgID, userGrantProjectID, userGrantUserID)
	if err != nil {
		return nil, err
	}
	return &auth_pb.ListMyProjectPermissionsResponse{
		Result: userGrant.Roles,
	}, nil
}

func (s *Server) ListMyMemberships(ctx context.Context, req *auth_pb.ListMyMembershipsRequest) (*auth_pb.ListMyMembershipsResponse, error) {
	request, err := ListMyMembershipsRequestToModel(ctx, req)
	if err != nil {
		return nil, err
	}
	response, err := s.query.Memberships(ctx, request, false)
	if err != nil {
		return nil, err
	}
	return &auth_pb.ListMyMembershipsResponse{
		Result:  user_grpc.MembershipsToMembershipsPb(response.Memberships),
		Details: object.ToListDetails(response.Count, response.Sequence, response.LastRun),
	}, nil
}
