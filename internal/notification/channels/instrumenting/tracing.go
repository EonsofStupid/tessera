package instrumenting

import (
	"context"

	"github.com/shippinAI/nomen/internal/notification/channels"
	"github.com/shippinAI/nomen/internal/telemetry/tracing"
)

func traceMessages(ctx context.Context, channel channels.NotificationChannel, spanName string) channels.NotificationChannel {
	return channels.HandleMessageFunc(func(message channels.Message) (err error) {
		_, span := tracing.NewNamedSpan(ctx, spanName)
		defer func() { span.EndWithError(err) }()
		return channel.HandleMessage(message)
	})
}
