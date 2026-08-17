package sms

import (
	"github.com/EonsofStupid/tessera/internal/notification/channels/twilio"
	"github.com/EonsofStupid/tessera/internal/notification/channels/webhook"
)

type Config struct {
	ProviderConfig *Provider
	TwilioConfig   *twilio.Config
	WebhookConfig  *webhook.Config
}

type Provider struct {
	ID          string `json:"id,omitempty"`
	Description string `json:"description,omitempty"`
}
