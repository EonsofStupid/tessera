package instance

import (
	"net/http"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/shippinAI/nomen/internal/api/authz"
	"github.com/shippinAI/nomen/internal/api/grpc/server"
	"github.com/shippinAI/nomen/internal/command"
	"github.com/shippinAI/nomen/internal/config/systemdefaults"
	"github.com/shippinAI/nomen/internal/query"
	instance "github.com/shippinAI/nomen/pkg/grpc/instance/v2beta"
	"github.com/shippinAI/nomen/pkg/grpc/instance/v2beta/instanceconnect"
)

var _ instanceconnect.InstanceServiceHandler = (*Server)(nil)

type Server struct {
	command         *command.Commands
	query           *query.Queries
	systemDefaults  systemdefaults.SystemDefaults
	defaultInstance command.InstanceSetup
	externalDomain  string
}

type Config struct{}

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
		defaultInstance: defaultInstance,
		externalDomain:  externalDomain,
	}
}

func (s *Server) RegisterConnectServer(interceptors ...connect.Interceptor) (string, http.Handler) {
	return instanceconnect.NewInstanceServiceHandler(s, connect.WithInterceptors(interceptors...))
}

func (s *Server) FileDescriptor() protoreflect.FileDescriptor {
	return instance.File_nomen_instance_v2beta_instance_service_proto
}

func (s *Server) AppName() string {
	return instance.InstanceService_ServiceDesc.ServiceName
}

func (s *Server) MethodPrefix() string {
	return instance.InstanceService_ServiceDesc.ServiceName
}

func (s *Server) AuthMethods() authz.MethodMapping {
	return instance.InstanceService_AuthMethods
}

func (s *Server) RegisterGateway() server.RegisterGatewayFunc {
	return instance.RegisterInstanceServiceHandler
}
