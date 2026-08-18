package domain

import (
	"os"
	"strings"
	"testing"
)

func TestOperationProtoCoversDomainVocabulary(t *testing.T) {
	t.Parallel()
	contents, err := os.ReadFile("../../../proto/tessera/management/v1/operation.proto")
	if err != nil {
		t.Fatalf("read operation proto: %v", err)
	}
	proto := string(contents)

	wantTokens := []string{
		"OPERATION_KIND_INSTALLATION",
		"OPERATION_KIND_GUIDE",
		"OPERATION_KIND_BACKUP",
		"OPERATION_KIND_RESTORE",
		"OPERATION_KIND_UPGRADE",
		"OPERATION_KIND_TRUST_ROTATION",
		"OPERATION_PHASE_PLAN",
		"OPERATION_PHASE_APPLY",
		"OPERATION_PHASE_VERIFY",
		"OPERATION_STATUS_QUEUED",
		"OPERATION_STATUS_RUNNING",
		"OPERATION_STATUS_SUCCEEDED",
		"OPERATION_STATUS_FAILED",
		"OPERATION_STATUS_CANCELING",
		"OPERATION_STATUS_CANCELED",
		"RETRY_DIRECTIVE_NONE",
		"RETRY_DIRECTIVE_SAME_REQUEST",
		"RETRY_DIRECTIVE_REPLAN",
		"RETRY_DIRECTIVE_OPERATOR_ACTION",
		"OPERATION_REFUSAL_REASON_INVALID_PLAN",
		"OPERATION_REFUSAL_REASON_PLAN_EXPIRED",
		"OPERATION_REFUSAL_REASON_PLAN_DIGEST_MISMATCH",
		"OPERATION_REFUSAL_REASON_STALE_BASE_REVISION",
		"OPERATION_REFUSAL_REASON_IDEMPOTENCY_KEY_REQUIRED",
		"OPERATION_REFUSAL_REASON_INVALID_IDEMPOTENCY_KEY",
		"OPERATION_REFUSAL_REASON_IDEMPOTENCY_KEY_REUSED",
		"OPERATION_REFUSAL_REASON_PROGRESS_SEQUENCE_INVALID",
		"OPERATION_REFUSAL_REASON_PROGRESS_PHASE_REGRESSED",
		"OPERATION_REFUSAL_REASON_OPERATION_TERMINAL",
	}
	for _, token := range wantTokens {
		if !strings.Contains(proto, token) {
			t.Errorf("operation.proto does not contain domain token %s", token)
		}
	}

	wantFields := []string{
		"string base_revision",
		"string desired_revision",
		"string plan_digest",
		"string idempotency_key",
		"repeated OperationProgressEvent progress",
		"OperationFailure failure",
		"repeated ProtectedSecretBinding secret_bindings",
	}
	for _, field := range wantFields {
		if !strings.Contains(proto, field) {
			t.Errorf("operation.proto does not contain contract field %q", field)
		}
	}
}
