package userschema

import (
	"context"

	"google.golang.org/grpc"

	"github.com/shippinAI/nomen/internal/api/authz"
	"github.com/shippinAI/nomen/internal/api/grpc/server"
	"github.com/shippinAI/nomen/internal/command"
	"github.com/shippinAI/nomen/internal/config/systemdefaults"
	"github.com/shippinAI/nomen/internal/query"
	"github.com/shippinAI/nomen/internal/zerrors"
	schema "github.com/shippinAI/nomen/pkg/grpc/resources/userschema/v3alpha"
)

var _ schema.NOMENUserSchemasServer = (*Server)(nil)

type Server struct {
	schema.UnimplementedNOMENUserSchemasServer
	systemDefaults systemdefaults.SystemDefaults
	command        *command.Commands
	query          *query.Queries
}

type Config struct{}

func CreateServer(
	systemDefaults systemdefaults.SystemDefaults,
	command *command.Commands,
	query *query.Queries,
) *Server {
	return &Server{
		systemDefaults: systemDefaults,
		command:        command,
		query:          query,
	}
}

func (s *Server) RegisterServer(grpcServer *grpc.Server) {
	schema.RegisterNOMENUserSchemasServer(grpcServer, s)
}

func (s *Server) AppName() string {
	return schema.NOMENUserSchemas_ServiceDesc.ServiceName
}

func (s *Server) MethodPrefix() string {
	return schema.NOMENUserSchemas_ServiceDesc.ServiceName
}

func (s *Server) AuthMethods() authz.MethodMapping {
	return schema.NOMENUserSchemas_AuthMethods
}

func (s *Server) RegisterGateway() server.RegisterGatewayFunc {
	return schema.RegisterNOMENUserSchemasHandler
}

func checkUserSchemaEnabled(ctx context.Context) error {
	if authz.GetInstance(ctx).Features().UserSchema {
		return nil
	}
	return zerrors.ThrowPreconditionFailed(nil, "SCHEMA-SFjk3", "Errors.UserSchema.NotEnabled")
}
