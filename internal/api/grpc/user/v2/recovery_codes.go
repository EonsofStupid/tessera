package user

import (
	"context"

	"connectrpc.com/connect"

	"github.com/shippinAI/nomen/internal/api/grpc/object/v2"
	user "github.com/shippinAI/nomen/pkg/grpc/user/v2"
)

func (s *Server) GenerateRecoveryCodes(ctx context.Context, req *connect.Request[user.GenerateRecoveryCodesRequest]) (*connect.Response[user.GenerateRecoveryCodesResponse], error) {
	details, err := s.command.GenerateRecoveryCodes(ctx, req.Msg.GetUserId(), int(req.Msg.GetCount()), "", nil)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&user.GenerateRecoveryCodesResponse{
		Details:       object.DomainToDetailsPb(details.ObjectDetails),
		RecoveryCodes: details.RawCodes,
	}), nil
}

func (s *Server) RemoveRecoveryCodes(ctx context.Context, req *connect.Request[user.RemoveRecoveryCodesRequest]) (*connect.Response[user.RemoveRecoveryCodesResponse], error) {
	objectDetails, err := s.command.RemoveRecoveryCodes(ctx, req.Msg.GetUserId(), "", nil)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&user.RemoveRecoveryCodesResponse{Details: object.DomainToDetailsPb(objectDetails)}), nil
}
