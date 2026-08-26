// Package overview reads the minimal Nomen-owned facts used by the
// provider-neutral management overview.
package overview

import (
	"context"
	"fmt"

	"github.com/shippinAI/nomen/backend/v1/domain"
	"github.com/shippinAI/nomen/backend/v3/storage/database"
)

type Repository struct {
	pool queryer
}

type queryer interface {
	QueryRow(ctx context.Context, stmt string, args ...any) database.Row
}

func NewRepository(pool queryer) *Repository {
	return &Repository{pool: pool}
}

// Snapshot deliberately reads only Nomen's schema. Inventory and billing
// remain host-product facts and cannot be smuggled into identity counts here.
func (r *Repository) Snapshot(ctx context.Context, instanceID string) (domain.OverviewFacts, error) {
	if instanceID == "" {
		return domain.OverviewFacts{}, fmt.Errorf("overview snapshot requires an instance")
	}

	const statement = `
SELECT
    COUNT(*) FILTER (WHERE occupant = 'human') AS human_seats,
    COUNT(*) FILTER (WHERE occupant = 'agent') AS agent_seats,
    (SELECT COUNT(*) FROM nomen_product.seat_workspaces WHERE instance_id = $1) AS workspace_attachments,
    (SELECT COUNT(*) FROM nomen_product.flows WHERE instance_id = $1) AS flows,
    ARRAY(
        SELECT DISTINCT policy_version
        FROM nomen_product.seats
        WHERE instance_id = $1 AND policy_version <> ''
        ORDER BY policy_version
    ) AS policy_revisions
FROM nomen_product.seats
WHERE instance_id = $1`

	var humanSeats, agentSeats, attachments, flows int64
	var policyRevisions []string
	err := r.pool.QueryRow(ctx, statement, instanceID).Scan(
		&humanSeats,
		&agentSeats,
		&attachments,
		&flows,
		&policyRevisions,
	)
	if err != nil {
		return domain.OverviewFacts{}, fmt.Errorf("read overview facts: %w", err)
	}
	if humanSeats < 0 || agentSeats < 0 || attachments < 0 || flows < 0 {
		return domain.OverviewFacts{}, fmt.Errorf("read overview facts: negative aggregate")
	}
	return domain.OverviewFacts{
		WorkspaceAttachments: uint64(attachments),
		AgentSeats:           uint64(agentSeats),
		HumanSeats:           uint64(humanSeats),
		Flows:                uint64(flows),
		PolicyRevisions:      policyRevisions,
	}, nil
}
