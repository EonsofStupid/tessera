package notification

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/shippinAI/nomen/logging"

	"github.com/shippinAI/nomen/internal/api/authz"
	"github.com/shippinAI/nomen/internal/command"
	"github.com/shippinAI/nomen/internal/crypto"
	"github.com/shippinAI/nomen/internal/eventstore"
	"github.com/shippinAI/nomen/internal/eventstore/handler/v2"
	"github.com/shippinAI/nomen/internal/id"
	"github.com/shippinAI/nomen/internal/notification/handlers"
	_ "github.com/shippinAI/nomen/internal/notification/statik"
	"github.com/shippinAI/nomen/internal/query"
	"github.com/shippinAI/nomen/internal/query/projection"
	"github.com/shippinAI/nomen/internal/queue"
)

var (
	projections []*handler.Handler
)

func Register(
	ctx context.Context,
	userHandlerCustomConfig, quotaHandlerCustomConfig, telemetryHandlerCustomConfig, backChannelLogoutHandlerCustomConfig projection.CustomConfig,
	notificationWorkerConfig handlers.WorkerConfig,
	backChannelLogoutWorkerConfig *handlers.BackChannelLogoutWorkerConfig,
	telemetryCfg handlers.TelemetryPusherConfig,
	externalDomain string,
	externalPort uint16,
	externalSecure bool,
	commands *command.Commands,
	queries *query.Queries,
	es *eventstore.Eventstore,
	otpEmailTmpl func(origin *url.URL) string,
	fileSystemPath string,
	userEncryption, smtpEncryption, smsEncryption crypto.EncryptionAlgorithm,
	queue *queue.Queue,
	httpClient *http.Client,
) {
	if !notificationWorkerConfig.LegacyEnabled {
		queue.ShouldStart()
	}

	// make sure the slice does not contain old values
	projections = nil

	q := handlers.NewNotificationQueries(queries, es, externalDomain, externalPort, externalSecure, fileSystemPath, userEncryption, smtpEncryption, smsEncryption, httpClient)
	c := newChannels(q)
	projections = append(projections, handlers.NewUserNotifier(ctx, projection.ApplyCustomConfig(userHandlerCustomConfig), commands, q, c, otpEmailTmpl, notificationWorkerConfig, queue))
	projections = append(projections, handlers.NewQuotaNotifier(ctx, projection.ApplyCustomConfig(quotaHandlerCustomConfig), commands, q, c))
	projections = append(projections, handlers.NewBackChannelLogoutNotifier(
		ctx,
		projection.ApplyCustomConfig(backChannelLogoutHandlerCustomConfig),
		q,
		queue,
		backChannelLogoutWorkerConfig.MaxAttempts,
	))
	queue.AddWorkers(ctx, handlers.NewBackChannelLogoutWorker(
		commands,
		q,
		es,
		queue,
		c,
		backChannelLogoutWorkerConfig,
		id.SonyFlakeGenerator(),
		httpClient,
	))
	if telemetryCfg.Enabled {
		projections = append(projections, handlers.NewTelemetryPusher(ctx, telemetryCfg, projection.ApplyCustomConfig(telemetryHandlerCustomConfig), commands, q, c))
	}
	if !notificationWorkerConfig.LegacyEnabled {
		queue.AddWorkers(ctx, handlers.NewNotificationWorker(notificationWorkerConfig, commands, q, c))
	}
}

func Start(ctx context.Context) {
	for _, projection := range projections {
		projection.Start(ctx)
	}
}

func SetCurrentState(ctx context.Context, es *eventstore.Eventstore) error {
	if len(projections) == 0 {
		return nil
	}
	position, err := es.LatestPosition(ctx, eventstore.NewSearchQueryBuilder(eventstore.ColumnsMaxPosition).InstanceID(authz.GetInstance(ctx).InstanceID()).OrderDesc().Limit(1))
	if err != nil {
		return err
	}

	for i, projection := range projections {
		logging.WithFields("name", projection.ProjectionName(), "instance", authz.GetInstance(ctx).InstanceID(), "index", fmt.Sprintf("%d/%d", i, len(projections))).Info("set current state of notification projection")
		_, err = projection.Trigger(ctx, handler.WithMinPosition(position))
		if err != nil {
			return err
		}
		logging.WithFields("name", projection.ProjectionName(), "instance", authz.GetInstance(ctx).InstanceID(), "index", fmt.Sprintf("%d/%d", i, len(projections))).Info("current state of notification projection set")
	}
	return nil
}

func ProjectInstance(ctx context.Context) error {
	for i, projection := range projections {
		logging.WithFields("name", projection.ProjectionName(), "instance", authz.GetInstance(ctx).InstanceID(), "index", fmt.Sprintf("%d/%d", i, len(projections))).Info("starting notification projection")
		_, err := projection.Trigger(ctx)
		if err != nil {
			return err
		}
		logging.WithFields("name", projection.ProjectionName(), "instance", authz.GetInstance(ctx).InstanceID(), "index", fmt.Sprintf("%d/%d", i, len(projections))).Info("notification projection done")
	}
	return nil
}

func Projections() []*handler.Handler {
	return projections
}
