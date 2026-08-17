package system

import (
	"context"

	object_pb "github.com/EonsofStupid/tessera/internal/api/grpc/object"
	"github.com/EonsofStupid/tessera/internal/command"
	"github.com/EonsofStupid/tessera/internal/domain"
	"github.com/EonsofStupid/tessera/internal/zerrors"
	system_pb "github.com/EonsofStupid/tessera/pkg/grpc/system"
)

func (s *Server) SetInstanceFeature(ctx context.Context, req *system_pb.SetInstanceFeatureRequest) (*system_pb.SetInstanceFeatureResponse, error) {
	details, err := s.setInstanceFeature(ctx, req)
	if err != nil {
		return nil, err
	}
	return &system_pb.SetInstanceFeatureResponse{
		Details: object_pb.DomainToChangeDetailsPb(details),
	}, nil

}

func (s *Server) setInstanceFeature(ctx context.Context, req *system_pb.SetInstanceFeatureRequest) (*domain.ObjectDetails, error) {
	feat := domain.Feature(req.FeatureId)
	if feat != domain.FeatureLoginDefaultOrg {
		return nil, zerrors.ThrowInvalidArgument(nil, "SYST-SGV45", "Errors.Feature.NotExisting")
	}
	switch t := req.Value.(type) {
	case *system_pb.SetInstanceFeatureRequest_Bool:
		return s.command.SetInstanceFeatures(ctx, &command.InstanceFeatures{
			LoginDefaultOrg: &t.Bool,
		})
	default:
		return nil, zerrors.ThrowInvalidArgument(nil, "SYST-dag5g", "Errors.Feature.TypeNotSupported")
	}
}
