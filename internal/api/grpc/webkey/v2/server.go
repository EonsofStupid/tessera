package webkey

import (
	"net/http"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/EonsofStupid/tessera/internal/api/authz"
	"github.com/EonsofStupid/tessera/internal/command"
	"github.com/EonsofStupid/tessera/internal/query"
	"github.com/EonsofStupid/tessera/pkg/grpc/webkey/v2"
	"github.com/EonsofStupid/tessera/pkg/grpc/webkey/v2/webkeyconnect"
)

var _ webkeyconnect.WebKeyServiceHandler = (*Server)(nil)

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

func (s *Server) AppName() string {
	return webkey.WebKeyService_ServiceDesc.ServiceName
}

func (s *Server) MethodPrefix() string {
	return webkey.WebKeyService_ServiceDesc.ServiceName
}

func (s *Server) AuthMethods() authz.MethodMapping {
	return webkey.WebKeyService_AuthMethods
}

func (s *Server) RegisterConnectServer(interceptors ...connect.Interceptor) (string, http.Handler) {
	return webkeyconnect.NewWebKeyServiceHandler(s, connect.WithInterceptors(interceptors...))
}

func (s *Server) FileDescriptor() protoreflect.FileDescriptor {
	return webkey.File_zitadel_webkey_v2_webkey_service_proto
}
