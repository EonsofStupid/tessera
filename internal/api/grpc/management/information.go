package management

import (
	"context"

	"github.com/shippinAI/nomen/oidc/v3/pkg/oidc"

	"github.com/shippinAI/nomen/internal/api/http"
	mgmt_pb "github.com/shippinAI/nomen/pkg/grpc/management"
)

func (s *Server) Healthz(context.Context, *mgmt_pb.HealthzRequest) (*mgmt_pb.HealthzResponse, error) {
	return &mgmt_pb.HealthzResponse{}, nil
}

func (s *Server) GetOIDCInformation(ctx context.Context, _ *mgmt_pb.GetOIDCInformationRequest) (*mgmt_pb.GetOIDCInformationResponse, error) {
	issuer := http.DomainContext(ctx).Origin()
	return &mgmt_pb.GetOIDCInformationResponse{
		Issuer:            issuer,
		DiscoveryEndpoint: issuer + oidc.DiscoveryEndpoint,
	}, nil
}
