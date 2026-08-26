package settings

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/shippinAI/nomen/internal/api/assets"
	"github.com/shippinAI/nomen/internal/api/authz"
	"github.com/shippinAI/nomen/internal/api/grpc/server"
	"github.com/shippinAI/nomen/internal/command"
	"github.com/shippinAI/nomen/internal/config/systemdefaults"
	"github.com/shippinAI/nomen/internal/domain"
	"github.com/shippinAI/nomen/internal/query"
	settings "github.com/shippinAI/nomen/pkg/grpc/settings/v2beta"
	"github.com/shippinAI/nomen/pkg/grpc/settings/v2beta/settingsconnect"
)

var _ settingsconnect.SettingsServiceHandler = (*Server)(nil)

type Server struct {
	systemDefaults systemdefaults.SystemDefaults
	command        *command.Commands
	query          *query.Queries

	checkPermission domain.PermissionCheck
	assetsAPIDomain func(context.Context) string
}

type Config struct{}

func CreateServer(
	systemDefaults systemdefaults.SystemDefaults,
	command *command.Commands,
	query *query.Queries,
	checkPermission domain.PermissionCheck,
) *Server {
	return &Server{
		systemDefaults:  systemDefaults,
		command:         command,
		query:           query,
		checkPermission: checkPermission,
		assetsAPIDomain: assets.AssetAPI(),
	}
}

func (s *Server) RegisterConnectServer(interceptors ...connect.Interceptor) (string, http.Handler) {
	return settingsconnect.NewSettingsServiceHandler(s, connect.WithInterceptors(interceptors...))
}

func (s *Server) FileDescriptor() protoreflect.FileDescriptor {
	return settings.File_nomen_settings_v2beta_settings_service_proto
}

func (s *Server) AppName() string {
	return settings.SettingsService_ServiceDesc.ServiceName
}

func (s *Server) MethodPrefix() string {
	return settings.SettingsService_ServiceDesc.ServiceName
}

func (s *Server) AuthMethods() authz.MethodMapping {
	return settings.SettingsService_AuthMethods
}

func (s *Server) RegisterGateway() server.RegisterGatewayFunc {
	return settings.RegisterSettingsServiceHandler
}
