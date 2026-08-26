package user

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/shippinAI/nomen/internal/api/authz"
	"github.com/shippinAI/nomen/internal/api/grpc/server"
	"github.com/shippinAI/nomen/internal/command"
	"github.com/shippinAI/nomen/internal/crypto"
	"github.com/shippinAI/nomen/internal/domain"
	"github.com/shippinAI/nomen/internal/query"
	user "github.com/shippinAI/nomen/pkg/grpc/user/v2beta"
	"github.com/shippinAI/nomen/pkg/grpc/user/v2beta/userconnect"
)

var _ userconnect.UserServiceHandler = (*Server)(nil)

type Server struct {
	command     *command.Commands
	query       *query.Queries
	userCodeAlg crypto.EncryptionAlgorithm
	idpAlg      crypto.EncryptionAlgorithm
	idpCallback func(ctx context.Context) string
	samlRootURL func(ctx context.Context, idpID string) string

	assetAPIPrefix func(context.Context) string

	checkPermission domain.PermissionCheck
}

type Config struct{}

func CreateServer(
	command *command.Commands,
	query *query.Queries,
	userCodeAlg crypto.EncryptionAlgorithm,
	idpAlg crypto.EncryptionAlgorithm,
	idpCallback func(ctx context.Context) string,
	samlRootURL func(ctx context.Context, idpID string) string,
	assetAPIPrefix func(ctx context.Context) string,
	checkPermission domain.PermissionCheck,
) *Server {
	return &Server{
		command:         command,
		query:           query,
		userCodeAlg:     userCodeAlg,
		idpAlg:          idpAlg,
		idpCallback:     idpCallback,
		samlRootURL:     samlRootURL,
		assetAPIPrefix:  assetAPIPrefix,
		checkPermission: checkPermission,
	}
}

func (s *Server) RegisterConnectServer(interceptors ...connect.Interceptor) (string, http.Handler) {
	return userconnect.NewUserServiceHandler(s, connect.WithInterceptors(interceptors...))
}

func (s *Server) FileDescriptor() protoreflect.FileDescriptor {
	return user.File_nomen_user_v2beta_user_service_proto
}

func (s *Server) AppName() string {
	return user.UserService_ServiceDesc.ServiceName
}

func (s *Server) MethodPrefix() string {
	return user.UserService_ServiceDesc.ServiceName
}

func (s *Server) AuthMethods() authz.MethodMapping {
	return user.UserService_AuthMethods
}

func (s *Server) RegisterGateway() server.RegisterGatewayFunc {
	return user.RegisterUserServiceHandler
}
