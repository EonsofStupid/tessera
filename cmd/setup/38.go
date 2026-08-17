package setup

import (
	"context"
	_ "embed"

	"github.com/EonsofStupid/tessera/internal/database"
	"github.com/EonsofStupid/tessera/internal/eventstore"
)

var (
	//go:embed 38.sql
	backChannelLogoutCurrentState string
)

type BackChannelLogoutNotificationStart struct {
	dbClient *database.DB
}

func (mig *BackChannelLogoutNotificationStart) Execute(ctx context.Context, e eventstore.Event) error {
	_, err := mig.dbClient.ExecContext(ctx, backChannelLogoutCurrentState, e.Sequence(), e.CreatedAt(), e.Position())
	return err
}

func (mig *BackChannelLogoutNotificationStart) String() string {
	return "38_back_channel_logout_notification_start_"
}
