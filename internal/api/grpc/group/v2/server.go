package group

import (
	"net/http"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/shippinAI/nomen/internal/api/authz"
	"github.com/shippinAI/nomen/internal/command"
	"github.com/shippinAI/nomen/internal/config/systemdefaults"
	"github.com/shippinAI/nomen/internal/domain"
	"github.com/shippinAI/nomen/internal/query"
	group_v2 "github.com/shippinAI/nomen/pkg/grpc/group/v2"
	"github.com/shippinAI/nomen/pkg/grpc/group/v2/groupconnect"
)

var _ groupconnect.GroupServiceHandler = (*Server)(nil)

type Server struct {
	systemDefaults systemdefaults.SystemDefaults
	command        *command.Commands
	query          *query.Queries

	checkPermission domain.PermissionCheck
}

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
	}
}

func (s *Server) RegisterConnectServer(interceptors ...connect.Interceptor) (string, http.Handler) {
	return groupconnect.NewGroupServiceHandler(s, connect.WithInterceptors(interceptors...))
}

func (s *Server) FileDescriptor() protoreflect.FileDescriptor {
	return group_v2.File_nomen_group_v2_group_service_proto
}

func (s *Server) AppName() string {
	return group_v2.GroupService_ServiceDesc.ServiceName
}

func (s *Server) MethodPrefix() string {
	return group_v2.GroupService_ServiceDesc.ServiceName
}

func (s *Server) AuthMethods() authz.MethodMapping {
	return group_v2.GroupService_AuthMethods
}
