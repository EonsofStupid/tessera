// Package management exposes Nomen's provider-neutral management reads.
package management

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/shippinAI/nomen/backend/v1/domain"
	nomen_management "github.com/shippinAI/nomen/backend/v1/management"
)

const (
	HandlerPrefix             = "/nomen/v1"
	OverviewPermission        = "nomen.overview.read"
	CapabilityPermission      = "nomen.capabilities.read"
	PreflightPermission       = "nomen.deployment.preflight.read"
	OwnerEnrollmentPermission = "nomen.deployment.owner_enrollment.read"
	OperatorEventPermission   = "nomen.operator_events.write"
	OperatorActionPermission  = "nomen.operator_actions.read"
)

type OverviewGetter interface {
	Get(ctx context.Context, instanceID, issuer string) (domain.Overview, error)
}

type CapabilityGetter interface {
	Get(ctx context.Context) (domain.CapabilityDiscovery, error)
}

type DeploymentPreflightGetter interface {
	Get(ctx context.Context, instanceID, issuer string) (domain.DeploymentPreflight, error)
}

type OwnerEnrollmentGetter interface {
	Get(ctx context.Context, instanceID string) (domain.OwnerEnrollmentView, error)
}

type OwnerEnrollmentMutator interface {
	GetWithBootstrap(ctx context.Context, instanceID, authority string) (domain.OwnerEnrollmentView, error)
	Begin(ctx context.Context, instanceID, issuer, authority, idempotencyKey string, request nomen_management.BeginOwnerEnrollmentRequest) (nomen_management.BeginOwnerEnrollmentResult, error)
	Complete(ctx context.Context, instanceID, issuer, authority string, request nomen_management.CompleteOwnerEnrollmentRequest) (nomen_management.CompleteOwnerEnrollmentResult, error)
	ConfirmRecovery(ctx context.Context, instanceID, authority string, request nomen_management.ConfirmOwnerRecoveryRequest) (domain.OwnerEnrollmentView, error)
}

type OperatorActionGetter interface {
	Get(ctx context.Context) (domain.OperatorActionCatalog, error)
}

type Authorizer interface {
	Authorize(r *http.Request, permission string) (*http.Request, *domain.ManagementError)
}

type InstanceResolver func(context.Context) string
type IssuerResolver func(*http.Request) string
type OperatorActorResolver func(context.Context) domain.OperatorActor

type OperatorEventRecorder interface {
	Record(ctx context.Context, actor domain.OperatorActor, batch domain.OperatorEventBatch) error
}

type Handler struct {
	overview        OverviewGetter
	capabilities    CapabilityGetter
	preflight       DeploymentPreflightGetter
	ownerEnrollment OwnerEnrollmentGetter
	ownerMutator    OwnerEnrollmentMutator
	authorizer      Authorizer
	instance        InstanceResolver
	issuer          IssuerResolver
	operatorEvents  OperatorEventRecorder
	operatorActor   OperatorActorResolver
	operatorActions OperatorActionGetter
	clock           func() time.Time
}

func NewHandler(overview OverviewGetter, authorizer Authorizer, instance InstanceResolver, issuer IssuerResolver) *Handler {
	return &Handler{overview: overview, authorizer: authorizer, instance: instance, issuer: issuer, clock: time.Now}
}

func (h *Handler) WithOperatorEvents(recorder OperatorEventRecorder, actor OperatorActorResolver) *Handler {
	h.operatorEvents = recorder
	h.operatorActor = actor
	return h
}

func (h *Handler) WithOperatorActions(actions OperatorActionGetter) *Handler {
	h.operatorActions = actions
	return h
}

func (h *Handler) WithCapabilities(capabilities CapabilityGetter) *Handler {
	h.capabilities = capabilities
	return h
}

func (h *Handler) WithDeploymentPreflight(preflight DeploymentPreflightGetter) *Handler {
	h.preflight = preflight
	return h
}

func (h *Handler) WithOwnerEnrollment(ownerEnrollment OwnerEnrollmentGetter) *Handler {
	h.ownerEnrollment = ownerEnrollment
	if mutator, ok := ownerEnrollment.(OwnerEnrollmentMutator); ok {
		h.ownerMutator = mutator
	}
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	setJSONNoStore(w)
	switch r.URL.Path {
	case "/overview":
		h.serveOverview(w, r)
	case "/capabilities":
		h.serveCapabilities(w, r)
	case "/deployment/preflight":
		h.serveDeploymentPreflight(w, r)
	case "/deployment/owner-enrollment":
		h.serveOwnerEnrollment(w, r)
	case "/deployment/owner-enrollment:begin":
		h.serveBeginOwnerEnrollment(w, r)
	case "/deployment/owner-enrollment:complete":
		h.serveCompleteOwnerEnrollment(w, r)
	case "/deployment/owner-enrollment/recovery:confirm":
		h.serveConfirmOwnerRecovery(w, r)
	case "/operator/events":
		h.serveOperatorEvents(w, r)
	case "/operator/actions":
		h.serveOperatorActions(w, r)
	default:
		writeError(w, notFoundError())
	}
}

func (h *Handler) serveBeginOwnerEnrollment(w http.ResponseWriter, r *http.Request) {
	if !h.ownerMutationReady(w, r) {
		return
	}
	var request nomen_management.BeginOwnerEnrollmentRequest
	if err := decodeStrictJSON(w, r, &request); err != nil {
		writeError(w, invalidOwnerEnrollmentError("body", "owner_enrollment_invalid"))
		return
	}
	result, err := h.ownerMutator.Begin(r.Context(), h.instance(r.Context()), h.issuer(r), bootstrapAuthority(r), r.Header.Get("Idempotency-Key"), request)
	if err != nil {
		writeError(w, ownerEnrollmentManagementError(err))
		return
	}
	if result.Created {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	_ = json.NewEncoder(w).Encode(result)
}

func (h *Handler) serveCompleteOwnerEnrollment(w http.ResponseWriter, r *http.Request) {
	if !h.ownerMutationReady(w, r) {
		return
	}
	var request nomen_management.CompleteOwnerEnrollmentRequest
	if err := decodeStrictJSON(w, r, &request); err != nil {
		writeError(w, invalidOwnerEnrollmentError("body", "owner_enrollment_invalid"))
		return
	}
	result, err := h.ownerMutator.Complete(r.Context(), h.instance(r.Context()), h.issuer(r), bootstrapAuthority(r), request)
	if err != nil {
		writeError(w, ownerEnrollmentManagementError(err))
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}

func (h *Handler) serveConfirmOwnerRecovery(w http.ResponseWriter, r *http.Request) {
	if !h.ownerMutationReady(w, r) {
		return
	}
	var request nomen_management.ConfirmOwnerRecoveryRequest
	if err := decodeStrictJSON(w, r, &request); err != nil {
		writeError(w, invalidOwnerEnrollmentError("body", "owner_enrollment_invalid"))
		return
	}
	view, err := h.ownerMutator.ConfirmRecovery(r.Context(), h.instance(r.Context()), bootstrapAuthority(r), request)
	if err != nil {
		writeError(w, ownerEnrollmentManagementError(err))
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(view)
}

func (h *Handler) ownerMutationReady(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeError(w, invalidMethodError())
		return false
	}
	if h.ownerMutator == nil || h.instance == nil || h.issuer == nil {
		writeError(w, unavailableError("owner-enrollment-mutation"))
		return false
	}
	if strings.TrimSpace(h.instance(r.Context())) == "" {
		writeError(w, unavailableError("owner-enrollment-instance"))
		return false
	}
	return true
}

func decodeStrictJSON(w http.ResponseWriter, r *http.Request, target any) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 128<<10))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request must contain exactly one JSON value")
	}
	return nil
}

func bootstrapAuthority(r *http.Request) string {
	scheme, value, ok := strings.Cut(strings.TrimSpace(r.Header.Get("Authorization")), " ")
	if !ok || !strings.EqualFold(scheme, "Bootstrap") {
		return ""
	}
	return strings.TrimSpace(value)
}

func ownerEnrollmentManagementError(err error) domain.ManagementError {
	var refusal *domain.OwnerEnrollmentError
	if !errors.As(err, &refusal) {
		return unavailableError("owner-enrollment-command")
	}
	if refusal.Field == "bootstrap_authority" {
		result := authenticationError()
		result.Reason = "bootstrap_authority_required"
		result.Message = "Present the deployment bootstrap authority to continue."
		return result
	}
	if refusal.Reason == domain.OwnerEnrollmentInvalid {
		return invalidOwnerEnrollmentError(refusal.Field, string(refusal.Reason))
	}
	return domain.ManagementError{
		Type: domain.ManagementErrorConflict, Reason: string(refusal.Reason),
		Message: "The owner-enrollment ceremony cannot continue from its current state.",
		Remedy:  domain.ManagementRemedy{Kind: domain.ManagementRemedyRefreshAndReview, Label: "Refresh and review"},
		Retry:   domain.RetryReplan,
	}
}

func invalidOwnerEnrollmentError(field, reason string) domain.ManagementError {
	return domain.ManagementError{
		Type: domain.ManagementErrorInvalidRequest, Reason: reason,
		Message: "Correct the owner-enrollment request and try again.",
		Remedy:  domain.ManagementRemedy{Kind: domain.ManagementRemedyCorrectRequest, Label: "Correct request"},
		Retry:   domain.RetryOperatorAction, DiagnosticReference: field,
	}
}

func (h *Handler) serveOwnerEnrollment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeError(w, invalidMethodError())
		return
	}
	if h.ownerEnrollment == nil || h.authorizer == nil || h.instance == nil {
		writeError(w, unavailableError("owner-enrollment-handler"))
		return
	}
	if authority := bootstrapAuthority(r); authority != "" {
		if h.ownerMutator == nil {
			writeError(w, unavailableError("owner-enrollment-bootstrap-read"))
			return
		}
		instanceID := strings.TrimSpace(h.instance(r.Context()))
		if instanceID == "" {
			writeError(w, unavailableError("owner-enrollment-instance"))
			return
		}
		view, err := h.ownerMutator.GetWithBootstrap(r.Context(), instanceID, authority)
		if err != nil {
			writeError(w, ownerEnrollmentManagementError(err))
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(view)
		return
	}
	authorized, managementError := h.authorizer.Authorize(r, OwnerEnrollmentPermission)
	if managementError != nil {
		writeError(w, *managementError)
		return
	}
	instanceID := strings.TrimSpace(h.instance(authorized.Context()))
	if instanceID == "" {
		writeError(w, unavailableError("owner-enrollment-instance"))
		return
	}
	view, err := h.ownerEnrollment.Get(authorized.Context(), instanceID)
	if err != nil {
		writeError(w, unavailableError("owner-enrollment-read"))
		return
	}
	if err := view.Validate(); err != nil {
		writeError(w, unavailableError("owner-enrollment-validation"))
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(view)
}

func (h *Handler) serveDeploymentPreflight(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeError(w, invalidMethodError())
		return
	}
	if h.preflight == nil || h.authorizer == nil || h.instance == nil || h.issuer == nil {
		writeError(w, unavailableError("deployment-preflight-handler"))
		return
	}
	authorized, managementError := h.authorizer.Authorize(r, PreflightPermission)
	if managementError != nil {
		writeError(w, *managementError)
		return
	}
	instanceID := strings.TrimSpace(h.instance(authorized.Context()))
	if instanceID == "" {
		writeError(w, unavailableError("deployment-preflight-instance"))
		return
	}
	preflight, err := h.preflight.Get(authorized.Context(), instanceID, h.issuer(authorized))
	if err != nil {
		writeError(w, unavailableError("deployment-preflight-read"))
		return
	}
	if err := preflight.Validate(); err != nil {
		writeError(w, unavailableError("deployment-preflight-validation"))
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(preflight)
}

func (h *Handler) serveOperatorActions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeError(w, invalidMethodError())
		return
	}
	if h.operatorActions == nil || h.authorizer == nil {
		writeError(w, unavailableError("operator-action-handler"))
		return
	}
	authorized, managementError := h.authorizer.Authorize(r, OperatorActionPermission)
	if managementError != nil {
		writeError(w, *managementError)
		return
	}
	catalog, err := h.operatorActions.Get(authorized.Context())
	if err != nil {
		writeError(w, unavailableError("operator-action-discovery"))
		return
	}
	if err := catalog.Validate(); err != nil {
		writeError(w, unavailableError("operator-action-validation"))
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(catalog)
}

func (h *Handler) serveOperatorEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeError(w, invalidMethodError())
		return
	}
	if h.operatorEvents == nil || h.operatorActor == nil || h.authorizer == nil || h.clock == nil {
		writeError(w, unavailableError("operator-event-handler"))
		return
	}
	authorized, managementError := h.authorizer.Authorize(r, OperatorEventPermission)
	if managementError != nil {
		writeError(w, *managementError)
		return
	}
	defer r.Body.Close()
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 128<<10))
	decoder.DisallowUnknownFields()
	var batch domain.OperatorEventBatch
	if err := decoder.Decode(&batch); err != nil {
		writeError(w, invalidOperatorEventError("body", "operator_event_invalid"))
		return
	}
	if err := batch.Validate(h.clock().UTC()); err != nil {
		writeError(w, invalidOperatorEventError("events", "operator_event_invalid"))
		return
	}
	actor := h.operatorActor(authorized.Context())
	if err := actor.Validate(); err != nil {
		writeError(w, unavailableError("operator-event-actor"))
		return
	}
	if err := h.operatorEvents.Record(authorized.Context(), actor, batch); err != nil {
		writeError(w, unavailableError("operator-event-record"))
		return
	}
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(struct {
		Accepted int `json:"accepted"`
	}{Accepted: len(batch.Events)})
}

func (h *Handler) serveOverview(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) serveCapabilities(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeError(w, invalidMethodError())
		return
	}
	if h.capabilities == nil || h.authorizer == nil {
		writeError(w, unavailableError("capability-handler"))
		return
	}
	authorized, managementError := h.authorizer.Authorize(r, CapabilityPermission)
	if managementError != nil {
		writeError(w, *managementError)
		return
	}
	discovery, err := h.capabilities.Get(authorized.Context())
	if err != nil {
		writeError(w, unavailableError("capability-discovery"))
		return
	}
	if err := discovery.Validate(); err != nil {
		writeError(w, unavailableError("capability-validation"))
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(discovery)
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
		Message:            "Your identity does not have permission to read this Nomen resource.",
		Remedy:             domain.ManagementRemedy{Kind: domain.ManagementRemedyRequestPermission, Label: "Request permission"},
		Retry:              domain.RetryOperatorAction,
		RequiredPermission: permission,
	}
}

func unavailableError(reference string) domain.ManagementError {
	return domain.ManagementError{
		Type:                domain.ManagementErrorServiceUnavailable,
		Reason:              "overview_unavailable",
		Message:             "Nomen cannot read this overview right now.",
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

func invalidOperatorEventError(field, reason string) domain.ManagementError {
	return domain.ManagementError{
		Type:    domain.ManagementErrorInvalidRequest,
		Reason:  reason,
		Message: "The semantic operator event does not match the public schema.",
		Remedy:  domain.ManagementRemedy{Kind: domain.ManagementRemedyCorrectRequest, Label: "Correct event"},
		Retry:   domain.RetryOperatorAction,
		Field:   field,
	}
}

func notFoundError() domain.ManagementError {
	return domain.ManagementError{
		Type:    domain.ManagementErrorInvalidRequest,
		Reason:  "resource_not_found",
		Message: "The requested Nomen management resource does not exist.",
		Remedy:  domain.ManagementRemedy{Kind: domain.ManagementRemedyCorrectRequest, Label: "Review request"},
		Retry:   domain.RetryOperatorAction,
		Field:   "path",
	}
}
