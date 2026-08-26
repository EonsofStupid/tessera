package types

import (
	"context"

	"github.com/shippinAI/nomen/internal/eventstore"
	"github.com/shippinAI/nomen/internal/notification/channels/webhook"
	"github.com/shippinAI/nomen/internal/notification/messages"
)

func handleWebhook(
	ctx context.Context,
	webhookConfig webhook.Config,
	channels ChannelChains,
	serializable interface{},
	triggeringEventType eventstore.EventType,
) error {
	message := &messages.JSON{
		Serializable:        serializable,
		TriggeringEventType: triggeringEventType,
	}
	webhookChannels, err := channels.Webhook(ctx, webhookConfig)
	if err != nil {
		return err
	}
	return webhookChannels.HandleMessage(message)
}
