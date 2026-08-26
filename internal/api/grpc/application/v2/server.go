package app

import (
	"net/http"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/shippinAI/nomen/internal/api/authz"
	"github.com/shippinAI/nomen/internal/command"
	"github.com/shippinAI/nomen/internal/config/systemdefaults"
	"github.com/shippinAI/nomen/internal/domain"
	"github.com/shippinAI/nomen/internal/query"
	"github.com/shippinAI/nomen/pkg/grpc/application/v2"
	"github.com/shippinAI/nomen/pkg/grpc/application/v2/applicationconnect"
)

var _ applicationconnect.ApplicationServiceHandler = (*Server)(nil)

type Server struct {
	command         *command.Commands
	query           *query.Queries
	systemDefaults  systemdefaults.SystemDefaults
	checkPermission domain.PermissionCheck
}

func CreateServer(
	systemDefaults systemdefaults.SystemDefaults,
	command *command.Commands,
	query *query.Queries,
	checkPermission domain.PermissionCheck,
) *Server {
	return &Server{
		command:         command,
		query:           query,
		checkPermission: checkPermission,
		systemDefaults:  systemDefaults,
	}
}

func (s *Server) RegisterConnectServer(interceptors ...connect.Interceptor) (string, http.Handler) {
	return applicationconnect.NewApplicationServiceHandler(s, connect.WithInterceptors(interceptors...))
}

func (s *Server) FileDescriptor() protoreflect.FileDescriptor {
	return application.File_nomen_application_v2_application_service_proto
}

func (s *Server) AppName() string {
	return application.ApplicationService_ServiceDesc.ServiceName
}

func (s *Server) MethodPrefix() string {
	return application.ApplicationService_ServiceDesc.ServiceName
}

func (s *Server) AuthMethods() authz.MethodMapping {
	return application.ApplicationService_AuthMethods
}
