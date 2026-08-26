// Package management assembles Nomen's provider-neutral management reads.
package management

import (
	"context"
	"fmt"
	"time"

	"github.com/shippinAI/nomen/backend/v1/domain"
	"github.com/shippinAI/nomen/internal/query"
)

type OverviewSource interface {
	Snapshot(ctx context.Context, instanceID string) (domain.OverviewFacts, error)
}

type SigningKeyCounter interface {
	ActiveSigningKeys(ctx context.Context) (uint32, error)
}

type Clock func() time.Time

type OverviewService struct {
	source OverviewSource
	keys   SigningKeyCounter
	clock  Clock
}

func NewOverviewService(source OverviewSource, keys SigningKeyCounter, clock Clock) *OverviewService {
	if clock == nil {
		clock = time.Now
	}
	return &OverviewService{source: source, keys: keys, clock: clock}
}

func (s *OverviewService) Get(ctx context.Context, instanceID, issuer string) (domain.Overview, error) {
	if s.source == nil || s.keys == nil {
		return domain.Overview{}, fmt.Errorf("overview service is incomplete")
	}
	facts, err := s.source.Snapshot(ctx, instanceID)
	if err != nil {
		return domain.Overview{}, fmt.Errorf("overview facts unavailable: %w", err)
	}
	keys, err := s.keys.ActiveSigningKeys(ctx)
	if err != nil {
		return domain.Overview{}, fmt.Errorf("signing-key facts unavailable: %w", err)
	}
	overview := domain.BuildOverview(facts, issuer, keys, s.clock())
	if err := overview.Validate(); err != nil {
		return domain.Overview{}, fmt.Errorf("assembled overview is invalid: %w", err)
	}
	return overview, nil
}

// QuerySigningKeyCounter reports only a usable active asymmetric web key. It
// never counts an HMAC/shared-secret path as verification readiness.
type QuerySigningKeyCounter struct {
	queries *query.Queries
}

func NewQuerySigningKeyCounter(queries *query.Queries) *QuerySigningKeyCounter {
	return &QuerySigningKeyCounter{queries: queries}
}

func (s *QuerySigningKeyCounter) ActiveSigningKeys(ctx context.Context) (uint32, error) {
	if s.queries == nil {
		return 0, fmt.Errorf("signing-key query is unavailable")
	}
	key, err := s.queries.GetActiveSigningWebKey(ctx)
	if err != nil {
		return 0, err
	}
	if key == nil || key.KeyID == "" || key.Key == nil {
		return 0, fmt.Errorf("active signing key is incomplete")
	}
	return 1, nil
}
