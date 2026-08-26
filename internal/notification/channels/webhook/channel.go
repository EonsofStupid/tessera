package webhook

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/shippinAI/nomen/logging"

	"github.com/shippinAI/nomen/internal/notification/channels"
	"github.com/shippinAI/nomen/internal/notification/messages"
	"github.com/shippinAI/nomen/internal/zerrors"
	"github.com/shippinAI/nomen/pkg/actions"
)

const (
	SigningHeader = "NOMEN-Signature"
)

func InitChannel(ctx context.Context, cfg Config) (channels.NotificationChannel, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	logging.Debug("successfully initialized webhook json channel")
	return channels.HandleMessageFunc(func(message channels.Message) error {
		requestCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		msg, ok := message.(*messages.JSON)
		if !ok {
			return zerrors.ThrowInternal(nil, "WEBH-K686U", "message is not JSON")
		}
		payload, err := msg.GetContent()
		if err != nil {
			return err
		}
		req, err := http.NewRequestWithContext(requestCtx, cfg.Method, cfg.CallURL, strings.NewReader(payload))
		if err != nil {
			return err
		}
		if cfg.Headers != nil {
			req.Header = cfg.Headers
		}
		req.Header.Set("Content-Type", "application/json")
		if cfg.SigningKey != "" {
			req.Header.Set(SigningHeader, actions.ComputeSignatureHeader(time.Now(), []byte(payload), cfg.SigningKey))
		}

		resp, err := cfg.Client.Do(req)
		if err != nil {
			return err
		}
		if err = resp.Body.Close(); err != nil {
			return err
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return zerrors.ThrowUnknown(fmt.Errorf("calling url %s returned %s", cfg.CallURL, resp.Status), "WEBH-LBxU0", "webhook didn't return a success status")
		}
		logging.WithFields("calling_url", cfg.CallURL, "method", cfg.Method).Debug("webhook called")
		return nil
	}), nil
}
