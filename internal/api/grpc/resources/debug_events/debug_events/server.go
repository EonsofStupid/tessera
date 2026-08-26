package debug_events

import (
	"google.golang.org/grpc"

	"github.com/shippinAI/nomen/internal/api/authz"
	"github.com/shippinAI/nomen/internal/api/grpc/server"
	"github.com/shippinAI/nomen/internal/command"
	"github.com/shippinAI/nomen/internal/query"
	debug_events "github.com/shippinAI/nomen/pkg/grpc/resources/debug_events/v3alpha"
)

type Server struct {
	debug_events.UnimplementedNOMENDebugEventsServer
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

func (s *Server) RegisterServer(grpcServer *grpc.Server) {
	debug_events.RegisterNOMENDebugEventsServer(grpcServer, s)
}

func (s *Server) AppName() string {
	return debug_events.NOMENDebugEvents_ServiceDesc.ServiceName
}

func (s *Server) MethodPrefix() string {
	return debug_events.NOMENDebugEvents_ServiceDesc.ServiceName
}

func (s *Server) AuthMethods() authz.MethodMapping {
	return debug_events.NOMENDebugEvents_AuthMethods
}

func (s *Server) RegisterGateway() server.RegisterGatewayFunc {
	return debug_events.RegisterNOMENDebugEventsHandler
}
