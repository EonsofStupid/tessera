package senders

import (
	"context"

	"github.com/zitadel/logging"

	"github.com/EonsofStupid/tessera/internal/api/authz"
	"github.com/EonsofStupid/tessera/internal/notification/channels"
	"github.com/EonsofStupid/tessera/internal/notification/channels/fs"
	"github.com/EonsofStupid/tessera/internal/notification/channels/instrumenting"
	"github.com/EonsofStupid/tessera/internal/notification/channels/log"
	"github.com/EonsofStupid/tessera/internal/notification/channels/webhook"
)

const webhookSpanName = "webhook.NotificationChannel"

func WebhookChannels(
	ctx context.Context,
	webhookConfig webhook.Config,
	getFileSystemProvider func(ctx context.Context) (*fs.Config, error),
	getLogProvider func(ctx context.Context) (*log.Config, error),
	successMetricName,
	failureMetricName string,
) (*Chain, error) {
	if err := webhookConfig.Validate(); err != nil {
		return nil, err
	}
	channels := make([]channels.NotificationChannel, 0, 3)
	webhookChannel, err := webhook.InitChannel(ctx, webhookConfig)
	logging.WithFields(
		"instance", authz.GetInstance(ctx).InstanceID(),
		"callurl", webhookConfig.CallURL,
	).OnError(err).Debug("initializing JSON channel failed")
	if err == nil {
		channels = append(
			channels,
			instrumenting.Wrap(
				ctx,
				webhookChannel,
				webhookSpanName,
				successMetricName,
				failureMetricName,
			),
		)
	}
	channels = append(channels, debugChannels(ctx, getFileSystemProvider, getLogProvider)...)
	return ChainChannels(channels...), nil
}
