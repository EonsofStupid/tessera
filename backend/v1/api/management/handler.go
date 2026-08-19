// Package management exposes Tessera's provider-neutral management reads.
package management

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/EonsofStupid/tessera/backend/v1/domain"
)

const (
	HandlerPrefix      = "/tessera/v1"
	OverviewPermission = "tessera:overview:read"
)

type OverviewGetter interface {
	Get(ctx context.Context, instanceID, issuer string) (domain.Overview, error)
}

type Authorizer interface {
	Authorize(r *http.Request, permission string) (*http.Request, *domain.ManagementError)
}

type InstanceResolver func(context.Context) string
type IssuerResolver func(*http.Request) string

type Handler struct {
	overview   OverviewGetter
	authorizer Authorizer
	instance   InstanceResolver
	issuer     IssuerResolver
}

func NewHandler(overview OverviewGetter, authorizer Authorizer, instance InstanceResolver, issuer IssuerResolver) *Handler {
	return &Handler{overview: overview, authorizer: authorizer, instance: instance, issuer: issuer}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	setJSONNoStore(w)
	if r.URL.Path != "/overview" {
		writeError(w, notFoundError())
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeError(w, invalidMethodError())
		return
	}
	if h.overview == nil || h.authorizer == nil || h.instance == nil || h.issuer == nil {
		writeError(w, unavailableError("overview-handler"))
		return
	}

	authorized, managementError := h.authorizer.Authorize(r, OverviewPermission)
	if managementError != nil {
		writeError(w, *managementError)
		return
	}
	instanceID := strings.TrimSpace(h.instance(authorized.Context()))
	if instanceID == "" {
		writeError(w, unavailableError("overview-instance"))
		return
	}
	overview, err := h.overview.Get(authorized.Context(), instanceID, h.issuer(authorized))
	if err != nil {
		writeError(w, unavailableError("overview-projection"))
		return
	}
	if err := overview.Validate(); err != nil {
		writeError(w, unavailableError("overview-validation"))
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(overview)
}

func setJSONNoStore(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "private, no-store")
}

func writeError(w http.ResponseWriter, managementError domain.ManagementError) {
	status := http.StatusInternalServerError
	if spec, ok := domain.ManagementErrorSpecFor(managementError.Type); ok {
		status = spec.HTTPStatus
	}
	if managementError.Type == domain.ManagementErrorAuthenticationRequired {
		w.Header().Set("WWW-Authenticate", "Bearer")
	}
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(domain.ManagementErrorEnvelope{Error: managementError})
}

func authenticationError() domain.ManagementError {
	return domain.ManagementError{
		Type:    domain.ManagementErrorAuthenticationRequired,
		Reason:  "authentication_required",
		Message: "Sign in to continue.",
		Remedy:  domain.ManagementRemedy{Kind: domain.ManagementRemedySignIn, Label: "Sign in"},
		Retry:   domain.RetryOperatorAction,
	}
}

func permissionError(permission string) domain.ManagementError {
	return domain.ManagementError{
		Type:               domain.ManagementErrorPermissionRequired,
		Reason:             "permission_required",
		Message:            "Your identity does not have permission to read this Tessera resource.",
		Remedy:             domain.ManagementRemedy{Kind: domain.ManagementRemedyRequestPermission, Label: "Request permission"},
		Retry:              domain.RetryOperatorAction,
		RequiredPermission: permission,
	}
}

func unavailableError(reference string) domain.ManagementError {
	return domain.ManagementError{
		Type:                domain.ManagementErrorServiceUnavailable,
		Reason:              "overview_unavailable",
		Message:             "Tessera cannot read this overview right now.",
		Remedy:              domain.ManagementRemedy{Kind: domain.ManagementRemedyRetryLater, Label: "Try again later"},
		Retry:               domain.RetrySameRequest,
		DiagnosticReference: reference,
	}
}

func invalidMethodError() domain.ManagementError {
	return domain.ManagementError{
		Type:    domain.ManagementErrorInvalidRequest,
		Reason:  "method_not_allowed",
		Message: "This resource only supports GET.",
		Remedy:  domain.ManagementRemedy{Kind: domain.ManagementRemedyCorrectRequest, Label: "Review request"},
		Retry:   domain.RetryOperatorAction,
		Field:   "method",
	}
}

func notFoundError() domain.ManagementError {
	return domain.ManagementError{
		Type:    domain.ManagementErrorInvalidRequest,
		Reason:  "resource_not_found",
		Message: "The requested Tessera management resource does not exist.",
		Remedy:  domain.ManagementRemedy{Kind: domain.ManagementRemedyCorrectRequest, Label: "Review request"},
		Retry:   domain.RetryOperatorAction,
		Field:   "path",
	}
}
