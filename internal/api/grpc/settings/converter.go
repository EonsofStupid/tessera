package settings

import (
	obj_pb "github.com/EonsofStupid/tessera/internal/api/grpc/object"
	"github.com/EonsofStupid/tessera/internal/query"
	settings_pb "github.com/EonsofStupid/tessera/pkg/grpc/settings"
)

func NotificationProviderToPb(provider *query.DebugNotificationProvider) *settings_pb.DebugNotificationProvider {
	mapped := &settings_pb.DebugNotificationProvider{
		Compact: provider.Compact,
		Details: obj_pb.ToViewDetailsPb(provider.Sequence, provider.CreationDate, provider.ChangeDate, provider.AggregateID),
	}
	return mapped
}
