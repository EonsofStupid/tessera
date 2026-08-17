package email

import (
	"github.com/EonsofStupid/tessera/internal/notification/channels/smtp"
	"github.com/EonsofStupid/tessera/internal/notification/channels/webhook"
)

type Config struct {
	ProviderConfig *Provider
	SMTPConfig     *smtp.Config
	WebhookConfig  *webhook.Config
}

type Provider struct {
	ID          string `json:"id,omitempty"`
	Description string `json:"description,omitempty"`
}
