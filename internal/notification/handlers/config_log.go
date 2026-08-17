package handlers

import (
	"context"

	"github.com/EonsofStupid/tessera/internal/api/authz"
	"github.com/EonsofStupid/tessera/internal/domain"
	"github.com/EonsofStupid/tessera/internal/notification/channels/log"
)

// GetLogProvider reads the iam log provider config
func (n *NotificationQueries) GetLogProvider(ctx context.Context) (*log.Config, error) {
	config, err := n.NotificationProviderByIDAndType(ctx, authz.GetInstance(ctx).InstanceID(), domain.NotificationProviderTypeLog)
	if err != nil {
		return nil, err
	}
	return &log.Config{
		Compact: config.Compact,
	}, nil
}
