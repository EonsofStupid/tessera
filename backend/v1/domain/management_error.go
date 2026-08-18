package domain

import (
	"fmt"
	"strings"
)

type ManagementErrorType string

const (
	ManagementErrorAuthenticationRequired ManagementErrorType = "authentication_required"
	ManagementErrorEntitlementRequired    ManagementErrorType = "entitlement_required"
	ManagementErrorPermissionRequired     ManagementErrorType = "permission_required"
	ManagementErrorStepUpRequired         ManagementErrorType = "step_up_required"
	ManagementErrorConflict               ManagementErrorType = "conflict"
	ManagementErrorInvalidRequest         ManagementErrorType = "invalid_request"
	ManagementErrorRateLimited            ManagementErrorType = "rate_limited"
	ManagementErrorServiceUnavailable     ManagementErrorType = "service_unavailable"
)

type ManagementRemedyKind string

const (
	ManagementRemedySignIn             ManagementRemedyKind = "sign_in"
	ManagementRemedyRequestEntitlement ManagementRemedyKind = "request_entitlement"
	ManagementRemedyRequestPermission  ManagementRemedyKind = "request_permission"
	ManagementRemedyStepUp             ManagementRemedyKind = "step_up"
	ManagementRemedyRefreshAndReview   ManagementRemedyKind = "refresh_and_review"
	ManagementRemedyCorrectRequest     ManagementRemedyKind = "correct_request"
	ManagementRemedyRetryAfter         ManagementRemedyKind = "retry_after"
	ManagementRemedyRetryLater         ManagementRemedyKind = "retry_later"
)

type ManagementRemedy struct {
	Kind  ManagementRemedyKind `json:"kind"`
	Label string               `json:"label"`
}

type ManagementError struct {
	Type                ManagementErrorType `json:"type"`
	Reason              string              `json:"reason"`
	Message             string              `json:"message"`
	Remedy              ManagementRemedy    `json:"remedy"`
	Retry               RetryDirective      `json:"retry"`
	MissingEntitlement  string              `json:"missing_entitlement,omitempty"`
	RequiredPermission  string              `json:"required_permission,omitempty"`
	RequiredAssurance   string              `json:"required_assurance,omitempty"`
	Field               string              `json:"field,omitempty"`
	CurrentRevision     string              `json:"current_revision,omitempty"`
	RetryAfterSeconds   uint32              `json:"retry_after_seconds,omitempty"`
	DiagnosticReference string              `json:"diagnostic_ref,omitempty"`
}

type ManagementErrorEnvelope struct {
	Error ManagementError `json:"error"`
}

func (e ManagementError) Error() string {
	return fmt.Sprintf("tessera management: %s: %s", e.Type, e.Reason)
}

func (e ManagementError) Validate() error {
	spec, ok := ManagementErrorSpecFor(e.Type)
	if !ok {
		return fmt.Errorf("unknown management error type %q", e.Type)
	}
	if strings.TrimSpace(e.Reason) == "" {
		return fmt.Errorf("management error reason is required")
	}
	if strings.TrimSpace(e.Message) == "" {
		return fmt.Errorf("management error message is required")
	}
	if e.Remedy.Kind != spec.Remedy {
		return fmt.Errorf("management error %s requires remedy %s", e.Type, spec.Remedy)
	}
	if strings.TrimSpace(e.Remedy.Label) == "" {
		return fmt.Errorf("management error remedy label is required")
	}
	if e.Retry != spec.Retry && !(e.Type == ManagementErrorServiceUnavailable && e.Retry == RetryOperatorAction) {
		return fmt.Errorf("management error %s requires retry %s", e.Type, spec.Retry)
	}
	if e.Type != ManagementErrorEntitlementRequired && e.MissingEntitlement != "" {
		return fmt.Errorf("missing_entitlement is only valid for entitlement_required")
	}
	if e.Type != ManagementErrorPermissionRequired && e.RequiredPermission != "" {
		return fmt.Errorf("required_permission is only valid for permission_required")
	}
	if e.Type != ManagementErrorStepUpRequired && e.RequiredAssurance != "" {
		return fmt.Errorf("required_assurance is only valid for step_up_required")
	}
	if e.Type != ManagementErrorRateLimited && e.RetryAfterSeconds != 0 {
		return fmt.Errorf("retry_after_seconds is only valid for rate_limited")
	}
	switch e.Type {
	case ManagementErrorEntitlementRequired:
		if strings.TrimSpace(e.MissingEntitlement) == "" {
			return fmt.Errorf("entitlement_required requires missing_entitlement")
		}
	case ManagementErrorPermissionRequired:
		if strings.TrimSpace(e.RequiredPermission) == "" {
			return fmt.Errorf("permission_required requires required_permission")
		}
	case ManagementErrorStepUpRequired:
		if strings.TrimSpace(e.RequiredAssurance) == "" {
			return fmt.Errorf("step_up_required requires required_assurance")
		}
	case ManagementErrorRateLimited:
		if e.RetryAfterSeconds == 0 {
			return fmt.Errorf("rate_limited requires positive retry_after_seconds")
		}
	case ManagementErrorServiceUnavailable:
		if strings.TrimSpace(e.DiagnosticReference) == "" {
			return fmt.Errorf("service_unavailable requires diagnostic_ref")
		}
	}
	return nil
}

type ManagementErrorSpec struct {
	Type       ManagementErrorType
	HTTPStatus int
	GRPCCode   string
	Retry      RetryDirective
	Remedy     ManagementRemedyKind
}

var managementErrorSpecs = []ManagementErrorSpec{
	{ManagementErrorAuthenticationRequired, 401, "UNAUTHENTICATED", RetryOperatorAction, ManagementRemedySignIn},
	{ManagementErrorEntitlementRequired, 403, "PERMISSION_DENIED", RetryOperatorAction, ManagementRemedyRequestEntitlement},
	{ManagementErrorPermissionRequired, 403, "PERMISSION_DENIED", RetryOperatorAction, ManagementRemedyRequestPermission},
	{ManagementErrorStepUpRequired, 403, "PERMISSION_DENIED", RetryOperatorAction, ManagementRemedyStepUp},
	{ManagementErrorConflict, 409, "ABORTED", RetryReplan, ManagementRemedyRefreshAndReview},
	{ManagementErrorInvalidRequest, 422, "INVALID_ARGUMENT", RetryOperatorAction, ManagementRemedyCorrectRequest},
	{ManagementErrorRateLimited, 429, "RESOURCE_EXHAUSTED", RetrySameRequest, ManagementRemedyRetryAfter},
	{ManagementErrorServiceUnavailable, 503, "UNAVAILABLE", RetrySameRequest, ManagementRemedyRetryLater},
}

func ManagementErrorCatalog() []ManagementErrorSpec {
	return append([]ManagementErrorSpec(nil), managementErrorSpecs...)
}

func ManagementErrorSpecFor(errorType ManagementErrorType) (ManagementErrorSpec, bool) {
	for _, spec := range managementErrorSpecs {
		if spec.Type == errorType {
			return spec, true
		}
	}
	return ManagementErrorSpec{}, false
}

func ManagementErrorFromOperationRefusal(refusal OperationContractError) ManagementError {
	errorType := ManagementErrorConflict
	remedy := ManagementRemedyRefreshAndReview
	retry := RetryReplan

	switch refusal.Reason {
	case OperationRefusalInvalidPlan,
		OperationRefusalIdempotencyKeyRequired,
		OperationRefusalInvalidIdempotencyKey:
		errorType = ManagementErrorInvalidRequest
		remedy = ManagementRemedyCorrectRequest
		retry = RetryOperatorAction
	case OperationRefusalProgressSequence,
		OperationRefusalProgressPhaseRegressed:
		errorType = ManagementErrorServiceUnavailable
		remedy = ManagementRemedyRetryLater
		retry = RetryOperatorAction
	}

	result := ManagementError{
		Type:    errorType,
		Reason:  string(refusal.Reason),
		Message: refusal.Detail,
		Remedy: ManagementRemedy{
			Kind:  remedy,
			Label: defaultRemedyLabel(remedy),
		},
		Retry: retry,
		Field: refusal.Field,
	}
	if errorType == ManagementErrorServiceUnavailable {
		result.DiagnosticReference = "operation-progress-contract"
	}
	return result
}

func defaultRemedyLabel(remedy ManagementRemedyKind) string {
	switch remedy {
	case ManagementRemedySignIn:
		return "Sign in"
	case ManagementRemedyRequestEntitlement:
		return "Request access"
	case ManagementRemedyRequestPermission:
		return "Request permission"
	case ManagementRemedyStepUp:
		return "Verify identity"
	case ManagementRemedyRefreshAndReview:
		return "Refresh and review"
	case ManagementRemedyCorrectRequest:
		return "Review input"
	case ManagementRemedyRetryAfter:
		return "Retry when ready"
	case ManagementRemedyRetryLater:
		return "Try again later"
	default:
		return "Review"
	}
}
