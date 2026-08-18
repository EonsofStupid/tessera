package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManagementErrorCatalog(t *testing.T) {
	t.Parallel()

	want := map[ManagementErrorType]struct {
		httpStatus int
		grpcCode   string
	}{
		ManagementErrorAuthenticationRequired: {401, "UNAUTHENTICATED"},
		ManagementErrorEntitlementRequired:    {403, "PERMISSION_DENIED"},
		ManagementErrorPermissionRequired:     {403, "PERMISSION_DENIED"},
		ManagementErrorStepUpRequired:         {403, "PERMISSION_DENIED"},
		ManagementErrorConflict:               {409, "ABORTED"},
		ManagementErrorInvalidRequest:         {422, "INVALID_ARGUMENT"},
		ManagementErrorRateLimited:            {429, "RESOURCE_EXHAUSTED"},
		ManagementErrorServiceUnavailable:     {503, "UNAVAILABLE"},
	}

	catalog := ManagementErrorCatalog()
	require.Len(t, catalog, len(want))
	for _, spec := range catalog {
		expected, ok := want[spec.Type]
		require.True(t, ok, "unexpected catalog type %q", spec.Type)
		assert.Equal(t, expected.httpStatus, spec.HTTPStatus)
		assert.Equal(t, expected.grpcCode, spec.GRPCCode)
		delete(want, spec.Type)
	}
	assert.Empty(t, want)

	entitlement, ok := ManagementErrorSpecFor(ManagementErrorEntitlementRequired)
	require.True(t, ok)
	assert.Equal(t, 403, entitlement.HTTPStatus, "missing entitlement must never map to 401")
	_, ok = ManagementErrorSpecFor("future_error")
	assert.False(t, ok)
}

func TestManagementErrorValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		error   ManagementError
		wantErr string
	}{
		{"authentication", validManagementError(ManagementErrorAuthenticationRequired), ""},
		{"entitlement", validManagementError(ManagementErrorEntitlementRequired), ""},
		{"permission", validManagementError(ManagementErrorPermissionRequired), ""},
		{"step up", validManagementError(ManagementErrorStepUpRequired), ""},
		{"conflict", validManagementError(ManagementErrorConflict), ""},
		{"invalid", validManagementError(ManagementErrorInvalidRequest), ""},
		{"rate limit", validManagementError(ManagementErrorRateLimited), ""},
		{"unavailable", validManagementError(ManagementErrorServiceUnavailable), ""},
		{"unknown", ManagementError{Type: "future"}, "unknown management error type"},
		{"reason required", without(validManagementError(ManagementErrorConflict), func(err *ManagementError) { err.Reason = "" }), "reason is required"},
		{"message required", without(validManagementError(ManagementErrorConflict), func(err *ManagementError) { err.Message = "" }), "message is required"},
		{"remedy kind", without(validManagementError(ManagementErrorConflict), func(err *ManagementError) { err.Remedy.Kind = ManagementRemedySignIn }), "requires remedy"},
		{"remedy label", without(validManagementError(ManagementErrorConflict), func(err *ManagementError) { err.Remedy.Label = "" }), "label is required"},
		{"retry", without(validManagementError(ManagementErrorConflict), func(err *ManagementError) { err.Retry = RetryNone }), "requires retry"},
		{"missing entitlement", without(validManagementError(ManagementErrorEntitlementRequired), func(err *ManagementError) { err.MissingEntitlement = "" }), "requires missing_entitlement"},
		{"missing permission", without(validManagementError(ManagementErrorPermissionRequired), func(err *ManagementError) { err.RequiredPermission = "" }), "requires required_permission"},
		{"missing assurance", without(validManagementError(ManagementErrorStepUpRequired), func(err *ManagementError) { err.RequiredAssurance = "" }), "requires required_assurance"},
		{"missing retry after", without(validManagementError(ManagementErrorRateLimited), func(err *ManagementError) { err.RetryAfterSeconds = 0 }), "positive retry_after_seconds"},
		{"missing diagnostic", without(validManagementError(ManagementErrorServiceUnavailable), func(err *ManagementError) { err.DiagnosticReference = "" }), "requires diagnostic_ref"},
		{"unexpected entitlement", without(validManagementError(ManagementErrorConflict), func(err *ManagementError) { err.MissingEntitlement = "tessera:manage" }), "only valid for entitlement_required"},
		{"unexpected permission", without(validManagementError(ManagementErrorConflict), func(err *ManagementError) { err.RequiredPermission = "tessera.idp.write" }), "only valid for permission_required"},
		{"unexpected assurance", without(validManagementError(ManagementErrorConflict), func(err *ManagementError) { err.RequiredAssurance = "phishing_resistant" }), "only valid for step_up_required"},
		{"unexpected retry after", without(validManagementError(ManagementErrorConflict), func(err *ManagementError) { err.RetryAfterSeconds = 30 }), "only valid for rate_limited"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.error.Validate()
			if test.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, test.wantErr)
		})
	}
}

func TestManagementErrorFromOperationRefusal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		reason   OperationContractRefusal
		wantType ManagementErrorType
	}{
		{OperationRefusalInvalidPlan, ManagementErrorInvalidRequest},
		{OperationRefusalIdempotencyKeyRequired, ManagementErrorInvalidRequest},
		{OperationRefusalInvalidIdempotencyKey, ManagementErrorInvalidRequest},
		{OperationRefusalPlanExpired, ManagementErrorConflict},
		{OperationRefusalPlanDigestMismatch, ManagementErrorConflict},
		{OperationRefusalStaleBaseRevision, ManagementErrorConflict},
		{OperationRefusalIdempotencyKeyReused, ManagementErrorConflict},
		{OperationRefusalTerminal, ManagementErrorConflict},
		{OperationRefusalProgressSequence, ManagementErrorServiceUnavailable},
		{OperationRefusalProgressPhaseRegressed, ManagementErrorServiceUnavailable},
	}

	for _, test := range tests {
		t.Run(string(test.reason), func(t *testing.T) {
			got := ManagementErrorFromOperationRefusal(OperationContractError{
				Reason: test.reason,
				Field:  "example",
				Detail: "safe explanation",
				Retry:  RetryNone,
			})
			assert.Equal(t, test.wantType, got.Type)
			require.NoError(t, got.Validate())
		})
	}
}

func validManagementError(errorType ManagementErrorType) ManagementError {
	spec, _ := ManagementErrorSpecFor(errorType)
	result := ManagementError{
		Type:    errorType,
		Reason:  "example_reason",
		Message: "A safe explanation.",
		Remedy: ManagementRemedy{
			Kind:  spec.Remedy,
			Label: defaultRemedyLabel(spec.Remedy),
		},
		Retry: spec.Retry,
	}
	switch errorType {
	case ManagementErrorEntitlementRequired:
		result.MissingEntitlement = "tessera:manage"
	case ManagementErrorPermissionRequired:
		result.RequiredPermission = "tessera.idp.write"
	case ManagementErrorStepUpRequired:
		result.RequiredAssurance = "phishing_resistant"
	case ManagementErrorRateLimited:
		result.RetryAfterSeconds = 30
	case ManagementErrorServiceUnavailable:
		result.DiagnosticReference = "diag_example"
	}
	return result
}

func without(value ManagementError, mutate func(*ManagementError)) ManagementError {
	mutate(&value)
	return value
}
