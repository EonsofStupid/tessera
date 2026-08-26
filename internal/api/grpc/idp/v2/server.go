package idp

import (
	"net/http"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/shippinAI/nomen/internal/api/authz"
	"github.com/shippinAI/nomen/internal/api/grpc/server"
	"github.com/shippinAI/nomen/internal/command"
	"github.com/shippinAI/nomen/internal/domain"
	"github.com/shippinAI/nomen/internal/query"
	"github.com/shippinAI/nomen/pkg/grpc/idp/v2"
	"github.com/shippinAI/nomen/pkg/grpc/idp/v2/idpconnect"
)

var _ idpconnect.IdentityProviderServiceHandler = (*Server)(nil)

type Server struct {
	command *command.Commands
	query   *query.Queries

	checkPermission domain.PermissionCheck
}

type Config struct{}

func CreateServer(
	command *command.Commands,
	query *query.Queries,
	checkPermission domain.PermissionCheck,
) *Server {
	return &Server{
		command:         command,
		query:           query,
		checkPermission: checkPermission,
	}
}

func (s *Server) RegisterConnectServer(interceptors ...connect.Interceptor) (string, http.Handler) {
	return idpconnect.NewIdentityProviderServiceHandler(s, connect.WithInterceptors(interceptors...))
}

func (s *Server) FileDescriptor() protoreflect.FileDescriptor {
	return idp.File_nomen_idp_v2_idp_service_proto
}

func (s *Server) AppName() string {
	return idp.IdentityProviderService_ServiceDesc.ServiceName
}

func (s *Server) MethodPrefix() string {
	return idp.IdentityProviderService_ServiceDesc.ServiceName
}

func (s *Server) AuthMethods() authz.MethodMapping {
	return idp.IdentityProviderService_AuthMethods
}

func (s *Server) RegisterGateway() server.RegisterGatewayFunc {
	return idp.RegisterIdentityProviderServiceHandler
}
