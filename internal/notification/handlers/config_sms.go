package handlers

import (
	"context"
	"net/http"

	"github.com/EonsofStupid/tessera/internal/api/authz"
	"github.com/EonsofStupid/tessera/internal/crypto"
	"github.com/EonsofStupid/tessera/internal/notification/channels/sms"
	"github.com/EonsofStupid/tessera/internal/notification/channels/twilio"
	"github.com/EonsofStupid/tessera/internal/notification/channels/webhook"
	"github.com/EonsofStupid/tessera/internal/zerrors"
)

// GetActiveSMSConfig reads the active iam sms provider config
func (n *NotificationQueries) GetActiveSMSConfig(ctx context.Context) (*sms.Config, error) {
	config, err := n.SMSProviderConfigActive(ctx, authz.GetInstance(ctx).InstanceID())
	if err != nil {
		return nil, err
	}

	provider := &sms.Provider{
		ID:          config.ID,
		Description: config.Description,
	}
	if config.TwilioConfig != nil {
		if config.TwilioConfig.Token == nil {
			return nil, zerrors.ThrowNotFound(err, "QUERY-SFefsd", "Errors.SMS.Twilio.NotFound")
		}
		token, err := crypto.DecryptString(config.TwilioConfig.Token, n.SMSTokenCrypto)
		if err != nil {
			return nil, err
		}
		return &sms.Config{
			ProviderConfig: provider,
			TwilioConfig: &twilio.Config{
				SID:              config.TwilioConfig.SID,
				Token:            token,
				SenderNumber:     config.TwilioConfig.SenderNumber,
				VerifyServiceSID: config.TwilioConfig.VerifyServiceSID,
			},
		}, nil
	}
	if config.HTTPConfig != nil {
		return &sms.Config{
			ProviderConfig: provider,
			WebhookConfig: &webhook.Config{
				CallURL:    config.HTTPConfig.Endpoint,
				Method:     http.MethodPost,
				Headers:    nil,
				SigningKey: config.HTTPConfig.SigningKey,
				Client:     n.httpClient,
			},
		}, nil
	}

	return nil, zerrors.ThrowNotFound(nil, "HANDLER-8nfow", "Errors.SMS.Twilio.NotFound")
}
