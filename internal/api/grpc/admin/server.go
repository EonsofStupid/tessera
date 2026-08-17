package admin

import (
	"context"
	"time"

	"google.golang.org/grpc"

	"github.com/EonsofStupid/tessera/internal/admin/repository/eventsourcing"
	"github.com/EonsofStupid/tessera/internal/api/assets"
	"github.com/EonsofStupid/tessera/internal/api/authz"
	"github.com/EonsofStupid/tessera/internal/api/grpc/server"
	"github.com/EonsofStupid/tessera/internal/command"
	"github.com/EonsofStupid/tessera/internal/crypto"
	"github.com/EonsofStupid/tessera/internal/query"
	"github.com/EonsofStupid/tessera/pkg/grpc/admin"
)

const (
	adminName = "Admin-API"
)

var _ admin.AdminServiceServer = (*Server)(nil)

type Server struct {
	admin.UnimplementedAdminServiceServer
	database          string
	command           *command.Commands
	query             *query.Queries
	assetsAPIDomain   func(context.Context) string
	userCodeAlg       crypto.EncryptionAlgorithm
	auditLogRetention time.Duration
}

type Config struct {
	Repository eventsourcing.Config
}

func CreateServer(
	database string,
	command *command.Commands,
	query *query.Queries,
	userCodeAlg crypto.EncryptionAlgorithm,
	auditLogRetention time.Duration,
) *Server {
	return &Server{
		database:          database,
		command:           command,
		query:             query,
		assetsAPIDomain:   assets.AssetAPI(),
		userCodeAlg:       userCodeAlg,
		auditLogRetention: auditLogRetention,
	}
}

func (s *Server) RegisterServer(grpcServer *grpc.Server) {
	admin.RegisterAdminServiceServer(grpcServer, s)
}

func (s *Server) AppName() string {
	return adminName
}

func (s *Server) MethodPrefix() string {
	return admin.AdminService_ServiceDesc.ServiceName
}

func (s *Server) AuthMethods() authz.MethodMapping {
	return admin.AdminService_AuthMethods
}

func (s *Server) RegisterGateway() server.RegisterGatewayFunc {
	return admin.RegisterAdminServiceHandler
}

func (s *Server) GatewayPathPrefix() string {
	return GatewayPathPrefix()
}

func GatewayPathPrefix() string {
	return "/admin/v1"
}
