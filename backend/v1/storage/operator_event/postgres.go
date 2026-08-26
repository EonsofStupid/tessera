// Package operator_event persists Nomen's semantic operator events and their
// analytical outbox in one tenant-scoped PostgreSQL transaction.
package operator_event

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/shippinAI/nomen/backend/v1/domain"
	"github.com/shippinAI/nomen/backend/v3/storage/database"
)

type Repository struct {
	pool database.Pool
}

func NewRepository(pool database.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Record(ctx context.Context, actor domain.OperatorActor, batch domain.OperatorEventBatch) (err error) {
	if err := actor.Validate(); err != nil {
		return err
	}
	tx, err := r.pool.Begin(ctx, &database.TransactionOptions{
		IsolationLevel: database.IsolationLevelReadCommitted,
		AccessMode:     database.AccessModeReadWrite,
	})
	if err != nil {
		return fmt.Errorf("begin operator event transaction: %w", err)
	}
	defer func() { err = tx.End(ctx, err) }()

	if _, err = tx.Exec(ctx, `SELECT set_config('nomen_product.instance_id', $1, true), set_config('nomen_product.tenant_id', $2, true)`, actor.InstanceID, actor.TenantID); err != nil {
		return fmt.Errorf("set operator event tenant context: %w", err)
	}

	for _, event := range batch.Events {
		attributes, marshalErr := json.Marshal(event.Attributes)
		if marshalErr != nil {
			return fmt.Errorf("encode operator event attributes: %w", marshalErr)
		}
		payload, marshalErr := json.Marshal(event)
		if marshalErr != nil {
			return fmt.Errorf("encode operator event outbox: %w", marshalErr)
		}
		if _, err = tx.Exec(ctx, `
WITH inserted AS (
    INSERT INTO nomen_product.nomen_operator_events (
        instance_id, tenant_id, event_id, session_id, sequence, occurred_at,
        route_id, control_id, event_type, action_id, resource_revision,
        correlation_id, outcome, attributes, actor_id, agent_id
    ) VALUES ($1, $2, $3::uuid, $4::uuid, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14::jsonb, $15, $16)
    ON CONFLICT (instance_id, tenant_id, event_id) DO NOTHING
    RETURNING instance_id, tenant_id, event_id
)
INSERT INTO nomen_product.nomen_outbox (instance_id, tenant_id, event_id, topic, payload)
SELECT instance_id, tenant_id, event_id, 'operator.event.v1', $17::jsonb
FROM inserted
ON CONFLICT DO NOTHING`,
			actor.InstanceID, actor.TenantID, event.EventID, event.SessionID,
			event.Sequence, event.OccurredAt, event.RouteID, event.ControlID,
			string(event.EventType), event.ActionID, event.ResourceRevision,
			event.CorrelationID, string(event.Outcome), attributes, actor.ActorID,
			actor.AgentID, payload,
		); err != nil {
			return fmt.Errorf("record operator event: %w", err)
		}
	}
	return nil
}
