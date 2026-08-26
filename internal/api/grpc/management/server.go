package management

import (
	"context"

	"google.golang.org/grpc"

	"github.com/shippinAI/nomen/internal/api/assets"
	"github.com/shippinAI/nomen/internal/api/authz"
	"github.com/shippinAI/nomen/internal/api/grpc/server"
	"github.com/shippinAI/nomen/internal/command"
	"github.com/shippinAI/nomen/internal/config/systemdefaults"
	"github.com/shippinAI/nomen/internal/crypto"
	"github.com/shippinAI/nomen/internal/query"
	"github.com/shippinAI/nomen/pkg/grpc/management"
)

const (
	mgmtName = "Management-API"
)

var _ management.ManagementServiceServer = (*Server)(nil)

type Server struct {
	management.UnimplementedManagementServiceServer
	command        *command.Commands
	query          *query.Queries
	systemDefaults systemdefaults.SystemDefaults
	assetAPIPrefix func(context.Context) string
	userCodeAlg    crypto.EncryptionAlgorithm
}

func CreateServer(
	command *command.Commands,
	query *query.Queries,
	sd systemdefaults.SystemDefaults,
	userCodeAlg crypto.EncryptionAlgorithm,
) *Server {
	return &Server{
		command:        command,
		query:          query,
		systemDefaults: sd,
		assetAPIPrefix: assets.AssetAPI(),
		userCodeAlg:    userCodeAlg,
	}
}

func (s *Server) RegisterServer(grpcServer *grpc.Server) {
	management.RegisterManagementServiceServer(grpcServer, s)
}

func (s *Server) AppName() string {
	return mgmtName
}

func (s *Server) MethodPrefix() string {
	return management.ManagementService_ServiceDesc.ServiceName
}

func (s *Server) AuthMethods() authz.MethodMapping {
	return management.ManagementService_AuthMethods
}

func (s *Server) RegisterGateway() server.RegisterGatewayFunc {
	return management.RegisterManagementServiceHandler
}

func (s *Server) GatewayPathPrefix() string {
	return GatewayPathPrefix()
}

func GatewayPathPrefix() string {
	return "/management/v1"
}
