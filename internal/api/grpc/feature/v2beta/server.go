package feature

import (
	"net/http"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/EonsofStupid/tessera/internal/api/authz"
	"github.com/EonsofStupid/tessera/internal/api/grpc/server"
	"github.com/EonsofStupid/tessera/internal/command"
	"github.com/EonsofStupid/tessera/internal/query"
	feature "github.com/EonsofStupid/tessera/pkg/grpc/feature/v2beta"
	"github.com/EonsofStupid/tessera/pkg/grpc/feature/v2beta/featureconnect"
)

var _ featureconnect.FeatureServiceHandler = (*Server)(nil)

type Server struct {
	command *command.Commands
	query   *query.Queries
}

func CreateServer(
	command *command.Commands,
	query *query.Queries,
) *Server {
	return &Server{
		command: command,
		query:   query,
	}
}

func (s *Server) RegisterConnectServer(interceptors ...connect.Interceptor) (string, http.Handler) {
	return featureconnect.NewFeatureServiceHandler(s, connect.WithInterceptors(interceptors...))
}

func (s *Server) FileDescriptor() protoreflect.FileDescriptor {
	return feature.File_zitadel_feature_v2beta_feature_service_proto
}

func (s *Server) AppName() string {
	return feature.FeatureService_ServiceDesc.ServiceName
}

func (s *Server) MethodPrefix() string {
	return feature.FeatureService_ServiceDesc.ServiceName
}

func (s *Server) AuthMethods() authz.MethodMapping {
	return feature.FeatureService_AuthMethods
}

func (s *Server) RegisterGateway() server.RegisterGatewayFunc {
	return feature.RegisterFeatureServiceHandler
}
