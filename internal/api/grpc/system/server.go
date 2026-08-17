package system

import (
	"google.golang.org/grpc"

	"github.com/EonsofStupid/tessera/internal/admin/repository/eventsourcing"
	"github.com/EonsofStupid/tessera/internal/api/authz"
	"github.com/EonsofStupid/tessera/internal/api/grpc/server"
	"github.com/EonsofStupid/tessera/internal/command"
	"github.com/EonsofStupid/tessera/internal/query"
	"github.com/EonsofStupid/tessera/pkg/grpc/system"
)

const (
	systemAPI = "System-API"
)

var _ system.SystemServiceServer = (*Server)(nil)

type Server struct {
	system.UnimplementedSystemServiceServer
	database        string
	command         *command.Commands
	query           *query.Queries
	defaultInstance command.InstanceSetup
	externalDomain  string
}

type Config struct {
	Repository eventsourcing.Config
}

func CreateServer(
	command *command.Commands,
	query *query.Queries,
	database string,
	defaultInstance command.InstanceSetup,
	externalDomain string,
) *Server {
	return &Server{
		command:         command,
		query:           query,
		database:        database,
		defaultInstance: defaultInstance,
		externalDomain:  externalDomain,
	}
}

func (s *Server) RegisterServer(grpcServer *grpc.Server) {
	system.RegisterSystemServiceServer(grpcServer, s)
}

func (s *Server) AppName() string {
	return systemAPI
}

func (s *Server) MethodPrefix() string {
	return system.SystemService_ServiceDesc.ServiceName
}

func (s *Server) AuthMethods() authz.MethodMapping {
	return system.SystemService_AuthMethods
}

func (s *Server) RegisterGateway() server.RegisterGatewayFunc {
	return system.RegisterSystemServiceHandler
}

func (s *Server) GatewayPathPrefix() string {
	return "/system/v1"
}
