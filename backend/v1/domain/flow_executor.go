package domain

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// Challenge is what a stage asks the client, tagged the way Nomen tags
// them: a component name any renderer can dispatch on. The panel's custom UI
// renders components, not endpoints — which is what lets login look like the
// product while the engine stays invisible.
type Challenge struct {
	Component string           `json:"component"`
	Flow      string           `json:"flow"`
	Title     string           `json:"title,omitempty"`
	Fields    []ChallengeField `json:"fields,omitempty"`
	// Errors are keyed by field; the empty key is the whole-stage error. A
	// failed factor stays on its stage with its errors — it does not advance,
	// and it does not end the execution.
	Errors map[string][]string `json:"errors,omitempty"`
}

// ChallengeField describes one input so a client can render it without
// knowing the stage kind.
type ChallengeField struct {
	Name     string `json:"name"`
	Type     string `json:"type"` // text | password | code
	Label    string `json:"label"`
	Required bool   `json:"required"`
}

// Execution is one client's walk through a plan. It lives server-side; the
// client holds the ID and the Token, nothing else — knowing someone's
// execution id must not be the ability to answer their challenges.
type Execution struct {
	ID         string    `json:"id"`
	Token      string    `json:"token"`
	InstanceID string    `json:"instance_id"`
	Plan       *Plan     `json:"plan"`
	Position   int       `json:"position"`
	ExpiresAt  time.Time `json:"expires_at"`

	// Accumulated by stages as they pass: identify sets the user and opens
	// the session; factor stages add checks to it.
	UserID        string `json:"user_id,omitempty"`
	SessionID     string `json:"session_id,omitempty"`
	SessionToken  string `json:"session_token,omitempty"`
	ResourceOwner string `json:"resource_owner,omitempty"`
}

// Done reports whether every planned stage has passed.
func (e *Execution) Done() bool { return e.Position >= len(e.Plan.Stages) }

// current returns the stage the execution is waiting on.
func (e *Execution) current() FlowStage { return e.Plan.Stages[e.Position] }

// StageRunner is a stage implementation. Implementations live in storage —
// they are the one place allowed to reach the command layer — and register by
// kind.
type StageRunner interface {
	Kind() StageKind
	// Challenge renders what this stage asks.
	Challenge(ctx context.Context, exec *Execution, stage FlowStage) (*Challenge, error)
	// Answer judges a submission. Field errors mean the stage stays and
	// re-asks with those errors; a nil, nil return means the stage passed and
	// the runner has recorded whatever it accumulated on the execution. An
	// error means infrastructure failed, not that the answer was wrong — the
	// two must never be conflated, because one is the client's to fix and the
	// other is ours.
	Answer(ctx context.Context, exec *Execution, stage FlowStage, answer map[string]any) (map[string][]string, error)
}

// The executor's refusals.
var (
	ErrExecutionExpired = errors.New("flow: this execution expired — start the flow again")
	ErrExecutionDone    = errors.New("flow: this execution already completed")
	ErrNoRunner         = errors.New("flow: no runner registered for stage kind")
)

// FlowExecutor walks executions through plans. It owns sequence and nothing
// else: what a stage asks and how an answer is judged belong to the runners.
type FlowExecutor struct {
	runners map[StageKind]StageRunner
	now     func() time.Time
	newID   func() string
	ttl     time.Duration
}

// NewFlowExecutor registers runners by kind. Duplicates panic at construction,
// where the stack names the registration — same rule as the blueprint engine.
func NewFlowExecutor(ttl time.Duration, runners ...StageRunner) *FlowExecutor {
	m := make(map[StageKind]StageRunner, len(runners))
	for _, r := range runners {
		if _, dup := m[r.Kind()]; dup {
			panic(fmt.Sprintf("flow: two runners registered for kind %q", r.Kind()))
		}
		m[r.Kind()] = r
	}
	return &FlowExecutor{runners: m, now: time.Now, newID: randomID, ttl: ttl}
}

// WithClock and WithIDs exist for tests; production takes the defaults.
func (x *FlowExecutor) WithClock(now func() time.Time) *FlowExecutor { x.now = now; return x }
func (x *FlowExecutor) WithIDs(newID func() string) *FlowExecutor    { x.newID = newID; return x }

// Start plans the flow and returns a new execution with its first challenge.
func (x *FlowExecutor) Start(ctx context.Context, instanceID string, f *Flow) (*Execution, *Challenge, error) {
	plan, err := PlanFlow(f)
	if err != nil {
		return nil, nil, err
	}
	// Every planned kind must have a runner before anything starts: a flow
	// that would strand a customer at stage three is refused at stage zero.
	for _, st := range plan.Stages {
		if _, ok := x.runners[st.Kind]; !ok {
			return nil, nil, fmt.Errorf("%w %q (flow %q)", ErrNoRunner, st.Kind, f.Slug)
		}
	}
	exec := &Execution{
		ID:         x.newID(),
		Token:      x.newID(),
		InstanceID: instanceID,
		Plan:       plan,
		ExpiresAt:  x.now().Add(x.ttl),
	}
	ch, err := x.challenge(ctx, exec)
	return exec, ch, err
}

// Advance judges an answer for the current stage.
//
// Three outcomes, deliberately distinct on the wire:
//   - the stage failed: the same challenge returns carrying field errors, and
//     the position does not move;
//   - the stage passed and stages remain: the next challenge returns;
//   - the last stage passed: nil challenge, and the execution is Done.
func (x *FlowExecutor) Advance(ctx context.Context, exec *Execution, answer map[string]any) (*Challenge, error) {
	if exec.Done() {
		return nil, ErrExecutionDone
	}
	if x.now().After(exec.ExpiresAt) {
		return nil, ErrExecutionExpired
	}

	stage := exec.current()
	runner := x.runners[stage.Kind]
	if runner == nil {
		return nil, fmt.Errorf("%w %q", ErrNoRunner, stage.Kind)
	}

	fieldErrs, err := runner.Answer(ctx, exec, stage, answer)
	if err != nil {
		return nil, err
	}
	if len(fieldErrs) > 0 {
		ch, err := runner.Challenge(ctx, exec, stage)
		if err != nil {
			return nil, err
		}
		ch.Errors = fieldErrs
		return ch, nil
	}

	exec.Position++
	if exec.Done() {
		return nil, nil
	}
	return x.challenge(ctx, exec)
}

func (x *FlowExecutor) challenge(ctx context.Context, exec *Execution) (*Challenge, error) {
	stage := exec.current()
	return x.runners[stage.Kind].Challenge(ctx, exec, stage)
}

func randomID() string {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		panic(err) // the platform's CSPRNG failing is not a recoverable state
	}
	return hex.EncodeToString(b)
}
