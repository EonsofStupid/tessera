package internal_permission

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
	internal_permission "github.com/shippinAI/nomen/pkg/grpc/internal_permission/v2beta"
	"github.com/shippinAI/nomen/pkg/grpc/internal_permission/v2beta/internal_permissionconnect"
)

var _ internal_permissionconnect.InternalPermissionServiceHandler = (*Server)(nil)

type Server struct {
	systemDefaults  systemdefaults.SystemDefaults
	command         *command.Commands
	query           *query.Queries
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
	return internal_permissionconnect.NewInternalPermissionServiceHandler(s, connect.WithInterceptors(interceptors...))
}

func (s *Server) FileDescriptor() protoreflect.FileDescriptor {
	return internal_permission.File_nomen_internal_permission_v2beta_internal_permission_service_proto
}

func (s *Server) AppName() string {
	return internal_permission.InternalPermissionService_ServiceDesc.ServiceName
}

func (s *Server) MethodPrefix() string {
	return internal_permission.InternalPermissionService_ServiceDesc.ServiceName
}

func (s *Server) AuthMethods() authz.MethodMapping {
	return internal_permission.InternalPermissionService_AuthMethods
}

func (s *Server) RegisterGateway() server.RegisterGatewayFunc {
	return internal_permission.RegisterInternalPermissionServiceHandler
}
