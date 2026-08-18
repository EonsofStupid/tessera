package domain

import "context"

// FlowRepository is where declared flows live — the same port shape as
// [SeatRepository], for the same reason: blueprints author flows, the executor
// reads them, and neither learns a table name.
type FlowRepository interface {
	// FlowBySlug returns a flow, or found=false when nothing declares it —
	// which is a caller's 404, not an error.
	FlowBySlug(ctx context.Context, instanceID, slug string) (*Flow, bool, error)
	// SetFlow writes the whole flow, stages replaced, not merged — a
	// blueprint declares what should be true.
	SetFlow(ctx context.Context, instanceID string, flow *Flow) error
	// RemoveFlow deletes one; removing the absent is success.
	RemoveFlow(ctx context.Context, instanceID, slug string) error
}

// ExecutionRepository persists in-progress executions for the HTTP surface.
// The token is stored and compared server-side; a caller that cannot present
// it does not get to answer.
type ExecutionRepository interface {
	SaveExecution(ctx context.Context, exec *Execution) error
	// ExecutionByID loads one, found=false when unknown or already reaped.
	ExecutionByID(ctx context.Context, instanceID, id string) (*Execution, bool, error)
	// UpdateExecution persists position and everything stages accumulated.
	UpdateExecution(ctx context.Context, exec *Execution) error
	// DeleteExecution removes one — completion consumes the execution.
	DeleteExecution(ctx context.Context, instanceID, id string) error
}
