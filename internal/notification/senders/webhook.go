package senders

import (
	"context"

	"github.com/shippinAI/nomen/logging"

	"github.com/shippinAI/nomen/internal/api/authz"
	"github.com/shippinAI/nomen/internal/notification/channels"
	"github.com/shippinAI/nomen/internal/notification/channels/fs"
	"github.com/shippinAI/nomen/internal/notification/channels/instrumenting"
	"github.com/shippinAI/nomen/internal/notification/channels/log"
	"github.com/shippinAI/nomen/internal/notification/channels/webhook"
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
