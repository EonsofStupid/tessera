// Package management assembles Nomen's deployment preflight without
// exposing probe errors or configuration payloads.
package management

import (
	"context"
	"fmt"
	"time"

	"github.com/shippinAI/nomen/backend/v1/domain"
	"github.com/shippinAI/nomen/internal/query"
)

type DatabaseHealthChecker interface {
	Health(ctx context.Context) error
}

type NotificationConfigurationChecker interface {
	Configured(ctx context.Context, instanceID string) (bool, error)
}

type DeploymentPreflightService struct {
	database      DatabaseHealthChecker
	keys          SigningKeyCounter
	notifications NotificationConfigurationChecker
	clock         Clock
}

func NewDeploymentPreflightService(database DatabaseHealthChecker, keys SigningKeyCounter, notifications NotificationConfigurationChecker, clock Clock) *DeploymentPreflightService {
	if clock == nil {
		clock = time.Now
	}
	return &DeploymentPreflightService{database: database, keys: keys, notifications: notifications, clock: clock}
}

func (s *DeploymentPreflightService) Get(ctx context.Context, instanceID, issuer string) (domain.DeploymentPreflight, error) {
	if s.database == nil || s.keys == nil || s.notifications == nil || s.clock == nil {
		return domain.DeploymentPreflight{}, fmt.Errorf("deployment preflight service is incomplete")
	}
	facts := domain.DeploymentPreflightFacts{}
	if err := s.database.Health(ctx); err == nil {
		facts.DatabaseProbeAvailable = true
		facts.DatabaseAvailable = true
	}
	if keys, err := s.keys.ActiveSigningKeys(ctx); err == nil {
		facts.SigningProbeAvailable = true
		facts.SigningKeys = keys
	}
	if configured, err := s.notifications.Configured(ctx, instanceID); err == nil {
		facts.NotificationProbeAvailable = true
		facts.NotificationConfigured = configured
	}
	preflight := domain.BuildDeploymentPreflight(facts, issuer, s.clock())
	if err := preflight.Validate(); err != nil {
		return domain.DeploymentPreflight{}, fmt.Errorf("assembled deployment preflight is invalid: %w", err)
	}
	return preflight, nil
}

type QueryDatabaseHealth struct {
	queries *query.Queries
}

func NewQueryDatabaseHealth(queries *query.Queries) *QueryDatabaseHealth {
	return &QueryDatabaseHealth{queries: queries}
}

func (p *QueryDatabaseHealth) Health(ctx context.Context) error {
	if p.queries == nil {
		return fmt.Errorf("database health query is unavailable")
	}
	return p.queries.Health(ctx)
}

type QueryNotificationConfiguration struct {
	queries *query.Queries
}

func NewQueryNotificationConfiguration(queries *query.Queries) *QueryNotificationConfiguration {
	return &QueryNotificationConfiguration{queries: queries}
}

func (p *QueryNotificationConfiguration) Configured(ctx context.Context, instanceID string) (bool, error) {
	if p.queries == nil || instanceID == "" {
		return false, fmt.Errorf("notification configuration query is unavailable")
	}
	configuration, err := p.queries.SMTPConfigActive(ctx, instanceID)
	if err != nil {
		return false, err
	}
	return configuration != nil && configuration.ID != "", nil
}
