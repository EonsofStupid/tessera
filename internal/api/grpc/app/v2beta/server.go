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
	app "github.com/shippinAI/nomen/pkg/grpc/app/v2beta"
	"github.com/shippinAI/nomen/pkg/grpc/app/v2beta/appconnect"
)

var _ appconnect.AppServiceHandler = (*Server)(nil)

type Server struct {
	command         *command.Commands
	query           *query.Queries
	systemDefaults  systemdefaults.SystemDefaults
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
	return appconnect.NewAppServiceHandler(s, connect.WithInterceptors(interceptors...))
}

func (s *Server) FileDescriptor() protoreflect.FileDescriptor {
	return app.File_nomen_app_v2beta_app_service_proto
}

func (s *Server) AppName() string {
	return app.AppService_ServiceDesc.ServiceName
}

func (s *Server) MethodPrefix() string {
	return app.AppService_ServiceDesc.ServiceName
}

func (s *Server) AuthMethods() authz.MethodMapping {
	return app.AppService_AuthMethods
}
