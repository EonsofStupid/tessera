package user

import (
	"google.golang.org/grpc"

	"github.com/shippinAI/nomen/internal/api/authz"
	"github.com/shippinAI/nomen/internal/api/grpc/server"
	"github.com/shippinAI/nomen/internal/command"
	user "github.com/shippinAI/nomen/pkg/grpc/resources/user/v3alpha"
)

var _ user.NOMENUsersServer = (*Server)(nil)

type Server struct {
	user.UnimplementedNOMENUsersServer
	command *command.Commands
}

type Config struct{}

func CreateServer(
	command *command.Commands,
) *Server {
	return &Server{
		command: command,
	}
}

func (s *Server) RegisterServer(grpcServer *grpc.Server) {
	user.RegisterNOMENUsersServer(grpcServer, s)
}

func (s *Server) AppName() string {
	return user.NOMENUsers_ServiceDesc.ServiceName
}

func (s *Server) MethodPrefix() string {
	return user.NOMENUsers_ServiceDesc.ServiceName
}

func (s *Server) AuthMethods() authz.MethodMapping {
	return user.NOMENUsers_AuthMethods
}

func (s *Server) RegisterGateway() server.RegisterGatewayFunc {
	return user.RegisterNOMENUsersHandler
}
