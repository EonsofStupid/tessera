package command

import (
	"github.com/EonsofStupid/tessera/internal/domain"
	"github.com/EonsofStupid/tessera/internal/eventstore"
	"github.com/EonsofStupid/tessera/internal/repository/settings"
)

type DebugNotificationWriteModel struct {
	eventstore.WriteModel

	Compact bool
	State   domain.NotificationProviderState
}

func (wm *DebugNotificationWriteModel) Reduce() error {
	for _, event := range wm.Events {
		switch e := event.(type) {
		case *settings.DebugNotificationProviderAddedEvent:
			wm.Compact = e.Compact
			wm.State = domain.NotificationProviderStateActive
		case *settings.DebugNotificationProviderChangedEvent:
			if e.Compact != nil {
				wm.Compact = *e.Compact
			}
		case *settings.DebugNotificationProviderRemovedEvent:
			wm.State = domain.NotificationProviderStateRemoved
		}
	}
	return wm.WriteModel.Reduce()
}
