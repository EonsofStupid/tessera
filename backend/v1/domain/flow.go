package domain

import (
	"errors"
	"fmt"
	"strings"
)

// A flow is an ordered set of stages, declared as a blueprint and executed
// one challenge at a time. The model is Authentik's — flow, stage, plan,
// challenge — reimplemented here because it is a model rather than a library:
// identity, MFA and recovery become configurations of one engine instead of
// three code paths.

// FlowDesignation says what completing the flow means. Two designations now;
// enrollment and invalidation join when something needs them.
type FlowDesignation string

const (
	DesignationAuthentication FlowDesignation = "authentication"
	DesignationRecovery       FlowDesignation = "recovery"
)

// ParseDesignation refuses the unknown outright — like DesiredState and unlike
// Basis, there is no safe default. A misspelled designation must not quietly
// become a login flow.
func ParseDesignation(raw string) (FlowDesignation, error) {
	switch FlowDesignation(strings.TrimSpace(strings.ToLower(raw))) {
	case DesignationAuthentication:
		return DesignationAuthentication, nil
	case DesignationRecovery:
		return DesignationRecovery, nil
	default:
		return "", fmt.Errorf("unknown designation %q (want %q or %q)",
			raw, DesignationAuthentication, DesignationRecovery)
	}
}

// StageKind names a stage implementation. The engine never learns what a kind
// does — it asks the registered [StageRunner] — but the kinds are enumerated
// here so a flow naming a stage nobody implements is refused at validation,
// not discovered by the first customer to reach it.
type StageKind string

const (
	// StageIdentify turns a login name into a pending user and starts the
	// session with CheckUser.
	StageIdentify StageKind = "identify"
	// StagePassword, StageTOTP and StageRecoveryCode verify a factor by
	// delegating to the command layer's session checks. The engine owns
	// sequence; the command layer owns truth.
	StagePassword     StageKind = "password"
	StageTOTP         StageKind = "totp"
	StageRecoveryCode StageKind = "recovery_code"
)

var knownStageKinds = map[StageKind]bool{
	StageIdentify:     true,
	StagePassword:     true,
	StageTOTP:         true,
	StageRecoveryCode: true,
}

// FlowStage is one step of a flow, in declaration order.
type FlowStage struct {
	Kind StageKind `json:"kind"`
	// Config is stage-specific and decoded strictly by the stage
	// implementation, exactly as blueprint attrs are: a typo'd key is an
	// error, never silence.
	Config map[string]any `json:"config,omitempty"`
}

// Flow is the declared object. Slug is its identity — it is what the start
// URL names — and stages execute in order.
type Flow struct {
	Slug        string          `json:"slug"`
	Title       string          `json:"title,omitempty"`
	Designation FlowDesignation `json:"designation"`
	Stages      []FlowStage     `json:"stages"`
}

// Validate is structural, in the blueprint style: every refusal at once, each
// naming the stage the way a review comment would.
func (f *Flow) Validate() error {
	var errs []error
	if strings.TrimSpace(f.Slug) == "" {
		errs = append(errs, errors.New("a flow needs a slug — it is the URL"))
	}
	if _, err := ParseDesignation(string(f.Designation)); err != nil {
		errs = append(errs, fmt.Errorf("flow %q: %w", f.Slug, err))
	}
	if len(f.Stages) == 0 {
		errs = append(errs, fmt.Errorf("flow %q has no stages — completing it would mean nothing was checked", f.Slug))
	}
	identifies := 0
	for i, st := range f.Stages {
		if !knownStageKinds[st.Kind] {
			errs = append(errs, fmt.Errorf("flow %q stage %d: unknown kind %q", f.Slug, i+1, st.Kind))
		}
		if st.Kind == StageIdentify {
			identifies++
			if i != 0 {
				errs = append(errs, fmt.Errorf("flow %q stage %d: identify must be first — every later stage needs a pending user", f.Slug, i+1))
			}
		}
	}
	// Factor stages act on the identified user's session; a flow that never
	// identifies would check factors against nobody.
	if identifies == 0 {
		errs = append(errs, fmt.Errorf("flow %q never identifies — factor stages would check against nobody", f.Slug))
	}
	if identifies > 1 {
		errs = append(errs, fmt.Errorf("flow %q identifies %d times — a flow has one subject", f.Slug, identifies))
	}
	return errors.Join(errs...)
}

// Plan is what the planner produced for one execution: the stages that apply,
// in order. v1 planning is the flow's stages verbatim; the seam exists so
// policy-gated bindings land here without the executor changing.
type Plan struct {
	FlowSlug    string          `json:"flow_slug"`
	Designation FlowDesignation `json:"designation"`
	Stages      []FlowStage     `json:"stages"`
}

// PlanFlow is the planner. Deliberately small and deliberately present: this
// is where Authentik evaluates per-binding policies and attaches
// re-evaluation markers, and where ours will when something needs them.
// Callers depend on the seam, not on the shortcut behind it.
func PlanFlow(f *Flow) (*Plan, error) {
	if err := f.Validate(); err != nil {
		return nil, err
	}
	return &Plan{
		FlowSlug:    f.Slug,
		Designation: f.Designation,
		Stages:      f.Stages,
	}, nil
}
