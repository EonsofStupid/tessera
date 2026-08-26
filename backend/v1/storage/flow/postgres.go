// Package flow implements [domain.FlowRepository] and
// [domain.ExecutionRepository] over Postgres, and the `nomen/flow`
// blueprint applier on top of them.
package flow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/shippinAI/nomen/backend/v1/domain"
	"github.com/shippinAI/nomen/backend/v3/storage/database"
)

// Repository reads and writes flows and executions.
type Repository struct {
	pool database.Pool
}

func NewRepository(pool database.Pool) *Repository { return &Repository{pool: pool} }

var (
	_ domain.FlowRepository      = (*Repository)(nil)
	_ domain.ExecutionRepository = (*Repository)(nil)
)

// FlowBySlug implements [domain.FlowRepository].
func (r *Repository) FlowBySlug(ctx context.Context, instanceID, slug string) (*domain.Flow, bool, error) {
	return flowBySlug(ctx, r.pool, instanceID, slug)
}

func flowBySlug(ctx context.Context, qe database.QueryExecutor, instanceID, slug string) (*domain.Flow, bool, error) {
	f := &domain.Flow{Slug: slug}
	var designation string
	err := qe.QueryRow(ctx,
		`SELECT title, designation::TEXT FROM nomen_product.flows WHERE instance_id = $1 AND slug = $2`,
		instanceID, slug).Scan(&f.Title, &designation)
	if err != nil {
		if errors.Is(err, &database.NoRowFoundError{}) {
			return nil, false, nil
		}
		return nil, false, err
	}
	f.Designation = domain.FlowDesignation(designation)

	rows, err := qe.Query(ctx,
		`SELECT kind, config FROM nomen_product.flow_stages
		 WHERE instance_id = $1 AND flow_slug = $2 ORDER BY position`,
		instanceID, slug)
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var (
			st  domain.FlowStage
			raw []byte
		)
		if err := rows.Scan(&st.Kind, &raw); err != nil {
			return nil, false, err
		}
		if len(raw) > 0 {
			if err := json.Unmarshal(raw, &st.Config); err != nil {
				return nil, false, fmt.Errorf("flow %s: stored stage config is not JSON: %w", slug, err)
			}
		}
		f.Stages = append(f.Stages, st)
	}
	return f, true, rows.Err()
}

// SetFlow implements [domain.FlowRepository] — its own transaction, for
// callers that arrive without one. The applier passes the engine's instead.
func (r *Repository) SetFlow(ctx context.Context, instanceID string, flow *domain.Flow) (err error) {
	tx, err := r.pool.Begin(ctx, &database.TransactionOptions{
		IsolationLevel: database.IsolationLevelReadCommitted,
		AccessMode:     database.AccessModeReadWrite,
	})
	if err != nil {
		return err
	}
	defer func() { err = tx.End(ctx, err) }()
	return upsertFlow(ctx, tx, instanceID, flow)
}

func upsertFlow(ctx context.Context, qe database.QueryExecutor, instanceID string, flow *domain.Flow) error {
	if err := flow.Validate(); err != nil {
		return err
	}
	if _, err := qe.Exec(ctx, `
INSERT INTO nomen_product.flows (instance_id, slug, title, designation)
VALUES ($1, $2, $3, $4::nomen_product.flow_designation)
ON CONFLICT (instance_id, slug) DO UPDATE SET
    title = EXCLUDED.title, designation = EXCLUDED.designation`,
		instanceID, flow.Slug, flow.Title, string(flow.Designation),
	); err != nil {
		return err
	}
	// Stages replaced whole: position is meaning, and merging positions from
	// two declarations produces an order nobody wrote.
	if _, err := qe.Exec(ctx,
		`DELETE FROM nomen_product.flow_stages WHERE instance_id = $1 AND flow_slug = $2`,
		instanceID, flow.Slug); err != nil {
		return err
	}
	for i, st := range flow.Stages {
		cfg, err := json.Marshal(st.Config)
		if err != nil {
			return err
		}
		if st.Config == nil {
			cfg = []byte("{}")
		}
		if _, err := qe.Exec(ctx, `
INSERT INTO nomen_product.flow_stages (instance_id, flow_slug, position, kind, config)
VALUES ($1, $2, $3, $4, $5)`,
			instanceID, flow.Slug, i, string(st.Kind), cfg); err != nil {
			return err
		}
	}
	return nil
}

// RemoveFlow implements [domain.FlowRepository].
func (r *Repository) RemoveFlow(ctx context.Context, instanceID, slug string) error {
	_, err := removeFlow(ctx, r.pool, instanceID, slug)
	return err
}

func removeFlow(ctx context.Context, qe database.QueryExecutor, instanceID, slug string) (int64, error) {
	return qe.Exec(ctx,
		`DELETE FROM nomen_product.flows WHERE instance_id = $1 AND slug = $2`,
		instanceID, slug)
}

// ---- executions ------------------------------------------------------------

// SaveExecution implements [domain.ExecutionRepository].
func (r *Repository) SaveExecution(ctx context.Context, exec *domain.Execution) error {
	plan, err := json.Marshal(exec.Plan)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `
INSERT INTO nomen_product.flow_executions
  (instance_id, id, token, plan, position, user_id, session_id, session_token, resource_owner, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		exec.InstanceID, exec.ID, exec.Token, plan, exec.Position,
		exec.UserID, exec.SessionID, exec.SessionToken, exec.ResourceOwner, exec.ExpiresAt)
	return err
}

// ExecutionByID implements [domain.ExecutionRepository].
func (r *Repository) ExecutionByID(ctx context.Context, instanceID, id string) (*domain.Execution, bool, error) {
	exec := &domain.Execution{InstanceID: instanceID, ID: id}
	var plan []byte
	err := r.pool.QueryRow(ctx, `
SELECT token, plan, position, user_id, session_id, session_token, resource_owner, expires_at
FROM nomen_product.flow_executions WHERE instance_id = $1 AND id = $2`,
		instanceID, id).Scan(
		&exec.Token, &plan, &exec.Position,
		&exec.UserID, &exec.SessionID, &exec.SessionToken, &exec.ResourceOwner, &exec.ExpiresAt)
	if err != nil {
		if errors.Is(err, &database.NoRowFoundError{}) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if err := json.Unmarshal(plan, &exec.Plan); err != nil {
		return nil, false, fmt.Errorf("execution %s: stored plan is not JSON: %w", id, err)
	}
	return exec, true, nil
}

// UpdateExecution implements [domain.ExecutionRepository].
func (r *Repository) UpdateExecution(ctx context.Context, exec *domain.Execution) error {
	_, err := r.pool.Exec(ctx, `
UPDATE nomen_product.flow_executions
SET position = $3, user_id = $4, session_id = $5, session_token = $6, resource_owner = $7
WHERE instance_id = $1 AND id = $2`,
		exec.InstanceID, exec.ID, exec.Position,
		exec.UserID, exec.SessionID, exec.SessionToken, exec.ResourceOwner)
	return err
}

// DeleteExecution implements [domain.ExecutionRepository].
func (r *Repository) DeleteExecution(ctx context.Context, instanceID, id string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM nomen_product.flow_executions WHERE instance_id = $1 AND id = $2`,
		instanceID, id)
	return err
}
