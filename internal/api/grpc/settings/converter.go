package settings

import (
	obj_pb "github.com/shippinAI/nomen/internal/api/grpc/object"
	"github.com/shippinAI/nomen/internal/query"
	settings_pb "github.com/shippinAI/nomen/pkg/grpc/settings"
)

func NotificationProviderToPb(provider *query.DebugNotificationProvider) *settings_pb.DebugNotificationProvider {
	mapped := &settings_pb.DebugNotificationProvider{
		Compact: provider.Compact,
		Details: obj_pb.ToViewDetailsPb(provider.Sequence, provider.CreationDate, provider.ChangeDate, provider.AggregateID),
	}
	return mapped
}
