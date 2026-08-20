package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

var operationTestNow = time.Date(2026, 8, 18, 18, 0, 0, 0, time.UTC)

func validOperationPlan() OperationPlan {
	return OperationPlan{
		ID:              "plan-01",
		Kind:            OperationInstallation,
		Scope:           OperationScope{AccountID: "account-01", WorkspaceID: "workspace-01"},
		BaseRevision:    "state:absent",
		DesiredRevision: "revision-ready-01",
		Digest:          "sha256:" + strings.Repeat("a", 64),
		CreatedAt:       operationTestNow.Add(-time.Minute),
		ExpiresAt:       operationTestNow.Add(time.Minute),
	}
}

func operationRefusal(t *testing.T, err error) *OperationContractError {
	t.Helper()
	var refusal *OperationContractError
	if !errors.As(err, &refusal) {
		t.Fatalf("error = %T %v, want *OperationContractError", err, err)
	}
	return refusal
}

func TestValidateOperationPlanForApply(t *testing.T) {
	t.Parallel()
	plan := validOperationPlan()
	if err := ValidateOperationPlanForApply(plan, plan.Digest, plan.BaseRevision, "apply-01", operationTestNow); err != nil {
		t.Fatalf("ValidateOperationPlanForApply(): %v", err)
	}
}

func TestValidateOperationPlanForApplyRefusals(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		mutate     func(*OperationPlan)
		digest     string
		revision   string
		key        string
		now        time.Time
		wantReason OperationContractRefusal
		wantField  string
		wantRetry  RetryDirective
	}{
		{name: "missing plan id", mutate: func(plan *OperationPlan) { plan.ID = "" }, wantReason: OperationRefusalInvalidPlan, wantField: "plan_id", wantRetry: RetryOperatorAction},
		{name: "unknown kind", mutate: func(plan *OperationPlan) { plan.Kind = "mystery" }, wantReason: OperationRefusalInvalidPlan, wantField: "kind", wantRetry: RetryOperatorAction},
		{name: "missing account", mutate: func(plan *OperationPlan) { plan.Scope.AccountID = " " }, wantReason: OperationRefusalInvalidPlan, wantField: "scope.account_id", wantRetry: RetryOperatorAction},
		{name: "missing base revision", mutate: func(plan *OperationPlan) { plan.BaseRevision = "" }, wantReason: OperationRefusalInvalidPlan, wantField: "base_revision", wantRetry: RetryOperatorAction},
		{name: "missing desired revision", mutate: func(plan *OperationPlan) { plan.DesiredRevision = "" }, wantReason: OperationRefusalInvalidPlan, wantField: "desired_revision", wantRetry: RetryOperatorAction},
		{name: "invalid digest", mutate: func(plan *OperationPlan) { plan.Digest = "sha256:ABC" }, wantReason: OperationRefusalInvalidPlan, wantField: "plan_digest", wantRetry: RetryOperatorAction},
		{name: "missing creation", mutate: func(plan *OperationPlan) { plan.CreatedAt = time.Time{} }, wantReason: OperationRefusalInvalidPlan, wantField: "created_at", wantRetry: RetryOperatorAction},
		{name: "expiration before creation", mutate: func(plan *OperationPlan) { plan.ExpiresAt = plan.CreatedAt }, wantReason: OperationRefusalInvalidPlan, wantField: "expires_at", wantRetry: RetryOperatorAction},
		{name: "missing idempotency key", wantReason: OperationRefusalIdempotencyKeyRequired, wantField: "idempotency_key", wantRetry: RetryOperatorAction},
		{name: "expired", key: "apply-01", now: operationTestNow.Add(2 * time.Minute), wantReason: OperationRefusalPlanExpired, wantField: "expires_at", wantRetry: RetryReplan},
		{name: "digest mismatch", digest: "sha256:" + strings.Repeat("b", 64), key: "apply-01", wantReason: OperationRefusalPlanDigestMismatch, wantField: "plan_digest", wantRetry: RetryReplan},
		{name: "stale revision", revision: "revision-new", key: "apply-01", wantReason: OperationRefusalStaleBaseRevision, wantField: "base_revision", wantRetry: RetryReplan},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			plan := validOperationPlan()
			if test.mutate != nil {
				test.mutate(&plan)
			}
			digest := test.digest
			if digest == "" {
				digest = plan.Digest
			}
			revision := test.revision
			if revision == "" {
				revision = plan.BaseRevision
			}
			now := test.now
			if now.IsZero() {
				now = operationTestNow
			}
			refusal := operationRefusal(t, ValidateOperationPlanForApply(plan, digest, revision, test.key, now))
			if refusal.Reason != test.wantReason || refusal.Field != test.wantField || refusal.Retry != test.wantRetry {
				t.Fatalf("refusal = %#v, want reason %q field %q retry %q", refusal, test.wantReason, test.wantField, test.wantRetry)
			}
		})
	}
}

func TestValidateIdempotencyKey(t *testing.T) {
	t.Parallel()
	for _, key := range []string{"a", "apply:account-01/request.01", strings.Repeat("x", 255)} {
		if err := ValidateIdempotencyKey(key); err != nil {
			t.Errorf("ValidateIdempotencyKey(%q): %v", key, err)
		}
	}
	for _, key := range []string{"", "has space", "line\nbreak", strings.Repeat("x", 256), "unicode-☃"} {
		if err := ValidateIdempotencyKey(key); err == nil {
			t.Errorf("ValidateIdempotencyKey(%q) = nil, want refusal", key)
		}
	}
}

func progressEvent(sequence uint64, phase OperationPhase, status OperationStatus) OperationProgressEvent {
	return OperationProgressEvent{
		Sequence:   sequence,
		Phase:      phase,
		Status:     status,
		StepCode:   "step",
		Retry:      RetryNone,
		ObservedAt: operationTestNow.Add(time.Duration(sequence) * time.Second),
	}
}

func TestValidateProgressAppend(t *testing.T) {
	t.Parallel()
	events := []OperationProgressEvent{
		progressEvent(1, OperationPhasePlan, OperationRunning),
		progressEvent(2, OperationPhaseApply, OperationRunning),
	}
	if err := ValidateProgressAppend(events, progressEvent(3, OperationPhaseVerify, OperationSucceeded)); err != nil {
		t.Fatalf("ValidateProgressAppend(): %v", err)
	}
}

func TestValidateProgressAppendRefusals(t *testing.T) {
	t.Parallel()
	validFirst := progressEvent(1, OperationPhasePlan, OperationRunning)
	tests := []struct {
		name       string
		existing   []OperationProgressEvent
		next       OperationProgressEvent
		wantReason OperationContractRefusal
	}{
		{name: "missing fields", next: OperationProgressEvent{Sequence: 1}, wantReason: OperationRefusalProgressSequence},
		{name: "wrong first sequence", next: progressEvent(2, OperationPhasePlan, OperationRunning), wantReason: OperationRefusalProgressSequence},
		{name: "completed without total", next: func() OperationProgressEvent {
			event := progressEvent(1, OperationPhasePlan, OperationRunning)
			event.CompletedUnits = 1
			return event
		}(), wantReason: OperationRefusalProgressSequence},
		{name: "completed exceeds total", next: func() OperationProgressEvent {
			event := progressEvent(1, OperationPhasePlan, OperationRunning)
			event.CompletedUnits = 2
			event.TotalUnits = 1
			return event
		}(), wantReason: OperationRefusalProgressSequence},
		{name: "success before verify", next: progressEvent(1, OperationPhaseApply, OperationSucceeded), wantReason: OperationRefusalProgressSequence},
		{name: "noncontiguous existing", existing: []OperationProgressEvent{progressEvent(2, OperationPhasePlan, OperationRunning)}, next: progressEvent(3, OperationPhaseApply, OperationRunning), wantReason: OperationRefusalProgressSequence},
		{name: "terminal existing", existing: []OperationProgressEvent{progressEvent(1, OperationPhasePlan, OperationFailed)}, next: progressEvent(2, OperationPhaseApply, OperationRunning), wantReason: OperationRefusalTerminal},
		{name: "event after earlier terminal", existing: []OperationProgressEvent{progressEvent(1, OperationPhasePlan, OperationFailed), progressEvent(2, OperationPhaseApply, OperationRunning)}, next: progressEvent(3, OperationPhaseVerify, OperationRunning), wantReason: OperationRefusalTerminal},
		{name: "phase regression", existing: []OperationProgressEvent{validFirst, progressEvent(2, OperationPhaseApply, OperationRunning)}, next: progressEvent(3, OperationPhasePlan, OperationRunning), wantReason: OperationRefusalProgressPhaseRegressed},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			refusal := operationRefusal(t, ValidateProgressAppend(test.existing, test.next))
			if refusal.Reason != test.wantReason {
				t.Fatalf("refusal reason = %q, want %q", refusal.Reason, test.wantReason)
			}
		})
	}
}

func TestOperationVocabulary(t *testing.T) {
	t.Parallel()
	for _, kind := range []OperationKind{OperationInstallation, OperationGuide, OperationBackup, OperationRestore, OperationUpgrade, OperationTrustRotation} {
		if !kind.Valid() {
			t.Errorf("kind %q is not valid", kind)
		}
	}
	for _, status := range []OperationStatus{OperationQueued, OperationRunning, OperationSucceeded, OperationFailed, OperationCanceling, OperationCanceled} {
		if !status.Valid() {
			t.Errorf("status %q is not valid", status)
		}
	}
	if OperationQueued.Terminal() || OperationRunning.Terminal() || OperationCanceling.Terminal() {
		t.Error("nonterminal status reported terminal")
	}
	if !OperationSucceeded.Terminal() || !OperationFailed.Terminal() || !OperationCanceled.Terminal() {
		t.Error("terminal status reported nonterminal")
	}
}

func TestValidateProtectedSecretBindings(t *testing.T) {
	t.Parallel()
	plan := validOperationPlan()
	plan.Requirements = []OperationRequirement{{
		Code:          "ldap.bind",
		SecretSlot:    "ldap-bind",
		SecretPurpose: SecretPurposeLDAPBind,
	}}
	reference := validSecretReference()
	reference.AccountID = plan.Scope.AccountID
	reference.WorkspaceID = plan.Scope.WorkspaceID
	bindings := []ProtectedSecretBinding{{Slot: "ldap-bind", ReferenceID: reference.ID}}
	references := map[string]SecretReference{reference.ID: reference}
	if err := ValidateProtectedSecretBindings(plan, bindings, references, secretTestNow); err != nil {
		t.Fatalf("ValidateProtectedSecretBindings(): %v", err)
	}

	tests := []struct {
		name       string
		bindings   []ProtectedSecretBinding
		references map[string]SecretReference
		want       OperationContractRefusal
	}{
		{"missing", nil, references, OperationRefusalSecretBindingRequired},
		{"unknown", []ProtectedSecretBinding{{Slot: "ldap-bind", ReferenceID: "unknown"}}, references, OperationRefusalSecretBindingUnknown},
		{"extra slot", []ProtectedSecretBinding{{Slot: "other", ReferenceID: reference.ID}}, references, OperationRefusalSecretBindingInvalid},
		{"duplicate", append(bindings, bindings...), references, OperationRefusalSecretBindingInvalid},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			refusal := operationRefusal(t, ValidateProtectedSecretBindings(plan, test.bindings, test.references, secretTestNow))
			if refusal.Reason != test.want {
				t.Fatalf("reason = %s, want %s", refusal.Reason, test.want)
			}
		})
	}

	wrongTenant := reference
	wrongTenant.AccountID = "other-account"
	if got := operationRefusal(t, ValidateProtectedSecretBindings(plan, bindings, map[string]SecretReference{reference.ID: wrongTenant}, secretTestNow)).Reason; got != OperationRefusalSecretBindingInvalid {
		t.Fatalf("cross-tenant reason = %s", got)
	}
	wrongPurpose := reference
	wrongPurpose.Purpose = SecretPurposeEdgeTLS
	if got := operationRefusal(t, ValidateProtectedSecretBindings(plan, bindings, map[string]SecretReference{reference.ID: wrongPurpose}, secretTestNow)).Reason; got != OperationRefusalSecretBindingInvalid {
		t.Fatalf("wrong-purpose reason = %s", got)
	}
}
