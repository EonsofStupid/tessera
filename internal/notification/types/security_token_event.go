package types

import (
	"context"

	"github.com/EonsofStupid/tessera/internal/eventstore"
	"github.com/EonsofStupid/tessera/internal/notification/channels/set"
	"github.com/EonsofStupid/tessera/internal/notification/messages"
)

func handleSecurityTokenEvent(
	ctx context.Context,
	setConfig set.Config,
	channels ChannelChains,
	token any,
	triggeringEventType eventstore.EventType,
) error {
	message := &messages.Form{
		Serializable:        token,
		TriggeringEventType: triggeringEventType,
	}
	setChannels, err := channels.SecurityTokenEvent(ctx, setConfig)
	if err != nil {
		return err
	}
	return setChannels.HandleMessage(message)
}
