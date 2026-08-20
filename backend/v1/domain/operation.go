package domain

import (
	"fmt"
	"strings"
	"time"
)

type OperationKind string

const (
	OperationInstallation  OperationKind = "installation"
	OperationGuide         OperationKind = "guide"
	OperationBackup        OperationKind = "backup"
	OperationRestore       OperationKind = "restore"
	OperationUpgrade       OperationKind = "upgrade"
	OperationTrustRotation OperationKind = "trust_rotation"
)

func (k OperationKind) Valid() bool {
	switch k {
	case OperationInstallation, OperationGuide, OperationBackup, OperationRestore, OperationUpgrade, OperationTrustRotation:
		return true
	default:
		return false
	}
}

type OperationPhase string

const (
	OperationPhasePlan   OperationPhase = "plan"
	OperationPhaseApply  OperationPhase = "apply"
	OperationPhaseVerify OperationPhase = "verify"
)

func (p OperationPhase) Valid() bool {
	return p == OperationPhasePlan || p == OperationPhaseApply || p == OperationPhaseVerify
}

type OperationStatus string

const (
	OperationQueued    OperationStatus = "queued"
	OperationRunning   OperationStatus = "running"
	OperationSucceeded OperationStatus = "succeeded"
	OperationFailed    OperationStatus = "failed"
	OperationCanceling OperationStatus = "canceling"
	OperationCanceled  OperationStatus = "canceled"
)

func (s OperationStatus) Valid() bool {
	switch s {
	case OperationQueued, OperationRunning, OperationSucceeded, OperationFailed, OperationCanceling, OperationCanceled:
		return true
	default:
		return false
	}
}

func (s OperationStatus) Terminal() bool {
	return s == OperationSucceeded || s == OperationFailed || s == OperationCanceled
}

type RetryDirective string

const (
	RetryNone           RetryDirective = "none"
	RetrySameRequest    RetryDirective = "same_request"
	RetryReplan         RetryDirective = "replan"
	RetryOperatorAction RetryDirective = "operator_action"
)

func (r RetryDirective) Valid() bool {
	return r == RetryNone || r == RetrySameRequest || r == RetryReplan || r == RetryOperatorAction
}

type OperationEffectAction string

const (
	EffectCreate OperationEffectAction = "create"
	EffectUpdate OperationEffectAction = "update"
	EffectRemove OperationEffectAction = "remove"
)

type OperationScope struct {
	AccountID   string `json:"account_id"`
	WorkspaceID string `json:"workspace_id,omitempty"`
}

type OperationEffect struct {
	ID                 string                `json:"id"`
	Action             OperationEffectAction `json:"action"`
	ResourceType       string                `json:"resource_type"`
	ResourceID         string                `json:"resource_id,omitempty"`
	Summary            string                `json:"summary"`
	Irreversible       bool                  `json:"irreversible"`
	RequiredPermission string                `json:"required_permission"`
}

type OperationRequirement struct {
	Code          string        `json:"code"`
	Satisfied     bool          `json:"satisfied"`
	Consequence   string        `json:"consequence"`
	Remediation   string        `json:"remediation,omitempty"`
	SecretSlot    string        `json:"secret_slot,omitempty"`
	SecretPurpose SecretPurpose `json:"secret_purpose,omitempty"`
}

type ProtectedSecretBinding struct {
	Slot        string `json:"slot"`
	ReferenceID string `json:"reference_id"`
}

type PlannedVerification struct {
	Code     string `json:"code"`
	Required bool   `json:"required"`
	Summary  string `json:"summary"`
}

type OperationPlan struct {
	ID                  string                 `json:"plan_id"`
	Kind                OperationKind          `json:"kind"`
	Scope               OperationScope         `json:"scope"`
	BaseRevision        string                 `json:"base_revision"`
	DesiredRevision     string                 `json:"desired_revision"`
	Digest              string                 `json:"plan_digest"`
	Effects             []OperationEffect      `json:"effects"`
	Requirements        []OperationRequirement `json:"requirements"`
	RequiredPermissions []string               `json:"required_permissions"`
	Verifications       []PlannedVerification  `json:"verifications"`
	CreatedAt           time.Time              `json:"created_at"`
	ExpiresAt           time.Time              `json:"expires_at"`
}

type OperationProgressEvent struct {
	Sequence       uint64          `json:"sequence"`
	Phase          OperationPhase  `json:"phase"`
	Status         OperationStatus `json:"status"`
	StepCode       string          `json:"step_code"`
	Summary        string          `json:"summary"`
	CompletedUnits uint64          `json:"completed_units,omitempty"`
	TotalUnits     uint64          `json:"total_units,omitempty"`
	Retry          RetryDirective  `json:"retry"`
	DiagnosticRef  string          `json:"diagnostic_ref,omitempty"`
	ObservedAt     time.Time       `json:"observed_at"`
}

type Operation struct {
	ID              string                   `json:"operation_id"`
	PlanID          string                   `json:"plan_id"`
	Kind            OperationKind            `json:"kind"`
	Scope           OperationScope           `json:"scope"`
	BaseRevision    string                   `json:"base_revision"`
	DesiredRevision string                   `json:"desired_revision"`
	PlanDigest      string                   `json:"plan_digest"`
	Phase           OperationPhase           `json:"phase"`
	Status          OperationStatus          `json:"status"`
	Progress        []OperationProgressEvent `json:"progress"`
	Cancelable      bool                     `json:"cancelable"`
	CreatedAt       time.Time                `json:"created_at"`
	UpdatedAt       time.Time                `json:"updated_at"`
	CompletedAt     *time.Time               `json:"completed_at,omitempty"`
	Failure         *OperationContractError  `json:"failure,omitempty"`
}

type OperationContractRefusal string

const (
	OperationRefusalInvalidPlan            OperationContractRefusal = "invalid_plan"
	OperationRefusalPlanExpired            OperationContractRefusal = "plan_expired"
	OperationRefusalPlanDigestMismatch     OperationContractRefusal = "plan_digest_mismatch"
	OperationRefusalStaleBaseRevision      OperationContractRefusal = "stale_base_revision"
	OperationRefusalIdempotencyKeyRequired OperationContractRefusal = "idempotency_key_required"
	OperationRefusalInvalidIdempotencyKey  OperationContractRefusal = "invalid_idempotency_key"
	OperationRefusalIdempotencyKeyReused   OperationContractRefusal = "idempotency_key_reused"
	OperationRefusalProgressSequence       OperationContractRefusal = "progress_sequence_invalid"
	OperationRefusalProgressPhaseRegressed OperationContractRefusal = "progress_phase_regressed"
	OperationRefusalTerminal               OperationContractRefusal = "operation_terminal"
	OperationRefusalSecretBindingRequired  OperationContractRefusal = "secret_binding_required"
	OperationRefusalSecretBindingUnknown   OperationContractRefusal = "secret_binding_unknown"
	OperationRefusalSecretBindingInvalid   OperationContractRefusal = "secret_binding_invalid"
)

type OperationContractError struct {
	Reason OperationContractRefusal `json:"reason"`
	Field  string                   `json:"field,omitempty"`
	Detail string                   `json:"detail"`
	Retry  RetryDirective           `json:"retry"`
}

func (e *OperationContractError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("operation contract: %s (%s): %s", e.Reason, e.Field, e.Detail)
	}
	return fmt.Sprintf("operation contract: %s: %s", e.Reason, e.Detail)
}

func ValidateOperationPlanForApply(plan OperationPlan, reviewedDigest, currentRevision, idempotencyKey string, now time.Time) error {
	if err := validateOperationPlan(plan); err != nil {
		return err
	}
	if err := ValidateIdempotencyKey(idempotencyKey); err != nil {
		return err
	}
	if !now.Before(plan.ExpiresAt) {
		return operationError(OperationRefusalPlanExpired, "expires_at", "the reviewed plan has expired", RetryReplan)
	}
	if reviewedDigest != plan.Digest {
		return operationError(OperationRefusalPlanDigestMismatch, "plan_digest", "the applied digest is not the reviewed plan digest", RetryReplan)
	}
	if currentRevision != plan.BaseRevision {
		return operationError(OperationRefusalStaleBaseRevision, "base_revision", "source state changed after planning", RetryReplan)
	}
	return nil
}

// ValidateProtectedSecretBindings resolves only safe reference projections.
// The provider reference and protected bytes never enter the operation request.
func ValidateProtectedSecretBindings(plan OperationPlan, bindings []ProtectedSecretBinding, references map[string]SecretReference, now time.Time) error {
	required := make(map[string]SecretPurpose)
	for _, requirement := range plan.Requirements {
		if requirement.SecretSlot == "" {
			if requirement.SecretPurpose.Valid() {
				return operationError(OperationRefusalSecretBindingInvalid, "requirements.secret_purpose", "a secret purpose requires a named slot", RetryReplan)
			}
			continue
		}
		if !safeIdentifier(requirement.SecretSlot) || !requirement.SecretPurpose.Valid() {
			return operationError(OperationRefusalSecretBindingInvalid, "requirements.secret_slot", "every protected slot requires a safe name and known purpose", RetryReplan)
		}
		if _, duplicate := required[requirement.SecretSlot]; duplicate {
			return operationError(OperationRefusalSecretBindingInvalid, "requirements.secret_slot", "the reviewed plan repeats a protected slot", RetryReplan)
		}
		required[requirement.SecretSlot] = requirement.SecretPurpose
	}

	seen := make(map[string]struct{}, len(bindings))
	for _, binding := range bindings {
		purpose, expected := required[binding.Slot]
		if !expected {
			return operationError(OperationRefusalSecretBindingInvalid, "secret_bindings.slot", "the binding names a slot absent from the reviewed plan", RetryReplan)
		}
		if _, duplicate := seen[binding.Slot]; duplicate {
			return operationError(OperationRefusalSecretBindingInvalid, "secret_bindings.slot", "a protected slot may be bound exactly once", RetryReplan)
		}
		seen[binding.Slot] = struct{}{}
		reference, found := references[binding.ReferenceID]
		if !found || binding.ReferenceID == "" {
			return operationError(OperationRefusalSecretBindingUnknown, "secret_bindings.reference_id", "the protected reference is not present", RetryOperatorAction)
		}
		if err := reference.Usable(now); err != nil {
			return operationError(OperationRefusalSecretBindingInvalid, "secret_bindings.reference_id", "the protected reference is not active", RetryOperatorAction)
		}
		if reference.AccountID != plan.Scope.AccountID || reference.WorkspaceID != plan.Scope.WorkspaceID {
			return operationError(OperationRefusalSecretBindingInvalid, "secret_bindings.reference_id", "the protected reference belongs to another tenant scope", RetryOperatorAction)
		}
		if reference.Purpose != purpose {
			return operationError(OperationRefusalSecretBindingInvalid, "secret_bindings.reference_id", "the protected reference has the wrong purpose", RetryOperatorAction)
		}
	}
	for slot := range required {
		if _, present := seen[slot]; !present {
			return operationError(OperationRefusalSecretBindingRequired, "secret_bindings", "the reviewed plan has an unresolved protected slot", RetryOperatorAction)
		}
	}
	return nil
}

func validateOperationPlan(plan OperationPlan) error {
	switch {
	case strings.TrimSpace(plan.ID) == "":
		return operationError(OperationRefusalInvalidPlan, "plan_id", "a stable plan id is required", RetryOperatorAction)
	case !plan.Kind.Valid():
		return operationError(OperationRefusalInvalidPlan, "kind", "a known operation kind is required", RetryOperatorAction)
	case strings.TrimSpace(plan.Scope.AccountID) == "":
		return operationError(OperationRefusalInvalidPlan, "scope.account_id", "an account boundary is required", RetryOperatorAction)
	case strings.TrimSpace(plan.BaseRevision) == "":
		return operationError(OperationRefusalInvalidPlan, "base_revision", "an explicit base revision is required", RetryOperatorAction)
	case strings.TrimSpace(plan.DesiredRevision) == "":
		return operationError(OperationRefusalInvalidPlan, "desired_revision", "an explicit desired revision is required", RetryOperatorAction)
	case !validPlanDigest(plan.Digest):
		return operationError(OperationRefusalInvalidPlan, "plan_digest", "want sha256 followed by 64 lowercase hexadecimal characters", RetryOperatorAction)
	case plan.CreatedAt.IsZero():
		return operationError(OperationRefusalInvalidPlan, "created_at", "creation time is required", RetryOperatorAction)
	case plan.ExpiresAt.IsZero() || !plan.ExpiresAt.After(plan.CreatedAt):
		return operationError(OperationRefusalInvalidPlan, "expires_at", "expiration must be after creation", RetryOperatorAction)
	default:
		return nil
	}
}

func ValidateIdempotencyKey(key string) error {
	if key == "" {
		return operationError(OperationRefusalIdempotencyKeyRequired, "idempotency_key", "a key is required", RetryOperatorAction)
	}
	if len(key) > 255 {
		return operationError(OperationRefusalInvalidIdempotencyKey, "idempotency_key", "the key exceeds 255 bytes", RetryOperatorAction)
	}
	for _, value := range []byte(key) {
		if value < 0x21 || value > 0x7e {
			return operationError(OperationRefusalInvalidIdempotencyKey, "idempotency_key", "use visible ASCII without whitespace or control bytes", RetryOperatorAction)
		}
	}
	return nil
}

func ValidateProgressAppend(existing []OperationProgressEvent, next OperationProgressEvent) error {
	if !next.Phase.Valid() || !next.Status.Valid() || !next.Retry.Valid() || strings.TrimSpace(next.StepCode) == "" || next.ObservedAt.IsZero() {
		return operationError(OperationRefusalProgressSequence, "progress", "phase, status, retry, step code and observation time are required", RetryNone)
	}
	if (next.TotalUnits == 0 && next.CompletedUnits != 0) || (next.TotalUnits != 0 && next.CompletedUnits > next.TotalUnits) {
		return operationError(OperationRefusalProgressSequence, "progress.work_units", "completed units require and cannot exceed total units", RetryNone)
	}
	if next.Status == OperationSucceeded && next.Phase != OperationPhaseVerify {
		return operationError(OperationRefusalProgressSequence, "progress.status", "operation success is only valid after verification", RetryNone)
	}
	wantSequence := uint64(1)
	if len(existing) > 0 {
		for index, event := range existing {
			if event.Sequence != uint64(index+1) {
				return operationError(OperationRefusalProgressSequence, "progress.sequence", "existing progress is not contiguous", RetryNone)
			}
			if index < len(existing)-1 && event.Status.Terminal() {
				return operationError(OperationRefusalTerminal, "progress.status", "existing progress continues after a terminal event", RetryNone)
			}
		}
		wantSequence = existing[len(existing)-1].Sequence + 1
	}
	if next.Sequence != wantSequence {
		return operationError(OperationRefusalProgressSequence, "progress.sequence", fmt.Sprintf("got %d, want %d", next.Sequence, wantSequence), RetryNone)
	}
	if len(existing) == 0 {
		return nil
	}
	previous := existing[len(existing)-1]
	if previous.Status.Terminal() {
		return operationError(OperationRefusalTerminal, "progress.status", "a terminal operation cannot accept another event", RetryNone)
	}
	if operationPhaseRank(next.Phase) < operationPhaseRank(previous.Phase) {
		return operationError(OperationRefusalProgressPhaseRegressed, "progress.phase", fmt.Sprintf("cannot move from %s back to %s", previous.Phase, next.Phase), RetryNone)
	}
	return nil
}

func validPlanDigest(value string) bool {
	if len(value) != len("sha256:")+64 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	for _, char := range value[len("sha256:"):] {
		if char < '0' || char > '9' {
			if char < 'a' || char > 'f' {
				return false
			}
		}
	}
	return true
}

func operationPhaseRank(phase OperationPhase) int {
	switch phase {
	case OperationPhasePlan:
		return 1
	case OperationPhaseApply:
		return 2
	case OperationPhaseVerify:
		return 3
	default:
		return 0
	}
}

func operationError(reason OperationContractRefusal, field, detail string, retry RetryDirective) error {
	return &OperationContractError{Reason: reason, Field: field, Detail: detail, Retry: retry}
}
