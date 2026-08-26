package saml

import (
	"net/http"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/shippinAI/nomen/internal/api/authz"
	"github.com/shippinAI/nomen/internal/api/grpc/server"
	"github.com/shippinAI/nomen/internal/api/saml"
	"github.com/shippinAI/nomen/internal/command"
	"github.com/shippinAI/nomen/internal/query"
	saml_pb "github.com/shippinAI/nomen/pkg/grpc/saml/v2"
	"github.com/shippinAI/nomen/pkg/grpc/saml/v2/samlconnect"
)

var _ samlconnect.SAMLServiceHandler = (*Server)(nil)

type Server struct {
	saml_pb.UnimplementedSAMLServiceServer
	command *command.Commands
	query   *query.Queries

	idp            *saml.Provider
	externalSecure bool
}

type Config struct{}

func CreateServer(
	command *command.Commands,
	query *query.Queries,
	idp *saml.Provider,
	externalSecure bool,
) *Server {
	return &Server{
		command:        command,
		query:          query,
		idp:            idp,
		externalSecure: externalSecure,
	}
}

func (s *Server) RegisterConnectServer(interceptors ...connect.Interceptor) (string, http.Handler) {
	return samlconnect.NewSAMLServiceHandler(s, connect.WithInterceptors(interceptors...))
}

func (s *Server) FileDescriptor() protoreflect.FileDescriptor {
	return saml_pb.File_nomen_saml_v2_saml_service_proto
}

func (s *Server) AppName() string {
	return saml_pb.SAMLService_ServiceDesc.ServiceName
}

func (s *Server) MethodPrefix() string {
	return saml_pb.SAMLService_ServiceDesc.ServiceName
}

func (s *Server) AuthMethods() authz.MethodMapping {
	return saml_pb.SAMLService_AuthMethods
}

func (s *Server) RegisterGateway() server.RegisterGatewayFunc {
	return saml_pb.RegisterSAMLServiceHandler
}
