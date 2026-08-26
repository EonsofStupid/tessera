package authorization

import (
	"net/http"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/shippinAI/nomen/internal/api/authz"
	"github.com/shippinAI/nomen/internal/api/grpc/server"
	"github.com/shippinAI/nomen/internal/command"
	"github.com/shippinAI/nomen/internal/config/systemdefaults"
	"github.com/shippinAI/nomen/internal/domain"
	"github.com/shippinAI/nomen/internal/query"
	authorization "github.com/shippinAI/nomen/pkg/grpc/authorization/v2beta"
	"github.com/shippinAI/nomen/pkg/grpc/authorization/v2beta/authorizationconnect"
)

var _ authorizationconnect.AuthorizationServiceHandler = (*Server)(nil)

type Server struct {
	systemDefaults systemdefaults.SystemDefaults
	command        *command.Commands
	query          *query.Queries

	checkPermission domain.PermissionCheck
}

type Config struct{}

func CreateServer(
	systemDefaults systemdefaults.SystemDefaults,
	command *command.Commands,
	query *query.Queries,
	checkPermission domain.PermissionCheck,
) *Server {
	return &Server{
		systemDefaults:  systemDefaults,
		command:         command,
		query:           query,
		checkPermission: checkPermission,
	}
}

func (s *Server) RegisterConnectServer(interceptors ...connect.Interceptor) (string, http.Handler) {
	return authorizationconnect.NewAuthorizationServiceHandler(s, connect.WithInterceptors(interceptors...))
}

func (s *Server) FileDescriptor() protoreflect.FileDescriptor {
	return authorization.File_nomen_authorization_v2beta_authorization_service_proto
}

func (s *Server) AppName() string {
	return authorization.AuthorizationService_ServiceDesc.ServiceName
}

func (s *Server) MethodPrefix() string {
	return authorization.AuthorizationService_ServiceDesc.ServiceName
}

func (s *Server) AuthMethods() authz.MethodMapping {
	return authorization.AuthorizationService_AuthMethods
}

func (s *Server) RegisterGateway() server.RegisterGatewayFunc {
	return authorization.RegisterAuthorizationServiceHandler
}
