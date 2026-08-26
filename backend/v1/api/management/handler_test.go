package management

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shippinAI/nomen/backend/v1/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type overviewGetterStub struct {
	overview domain.Overview
	err      error
}

type capabilityGetterStub struct {
	discovery domain.CapabilityDiscovery
	err       error
}

type deploymentPreflightGetterStub struct {
	preflight domain.DeploymentPreflight
	err       error
}

type ownerEnrollmentGetterStub struct {
	view domain.OwnerEnrollmentView
	err  error
}

func (s ownerEnrollmentGetterStub) Get(context.Context, string) (domain.OwnerEnrollmentView, error) {
	return s.view, s.err
}

func (s deploymentPreflightGetterStub) Get(context.Context, string, string) (domain.DeploymentPreflight, error) {
	return s.preflight, s.err
}

type operatorActionGetterStub struct {
	catalog domain.OperatorActionCatalog
	err     error
}

func (s operatorActionGetterStub) Get(context.Context) (domain.OperatorActionCatalog, error) {
	return s.catalog, s.err
}

func (s capabilityGetterStub) Get(context.Context) (domain.CapabilityDiscovery, error) {
	return s.discovery, s.err
}

func (s overviewGetterStub) Get(context.Context, string, string) (domain.Overview, error) {
	return s.overview, s.err
}

type authorizerStub struct{ failure *domain.ManagementError }

func (s authorizerStub) Authorize(r *http.Request, _ string) (*http.Request, *domain.ManagementError) {
	return r, s.failure
}

type operatorEventRecorderStub struct {
	actor domain.OperatorActor
	batch domain.OperatorEventBatch
	err   error
}

func (s *operatorEventRecorderStub) Record(_ context.Context, actor domain.OperatorActor, batch domain.OperatorEventBatch) error {
	s.actor = actor
	s.batch = batch
	return s.err
}

func TestHandlerReturnsValidatedOverview(t *testing.T) {
	projection := domain.BuildOverview(domain.OverviewFacts{Flows: 1, HumanSeats: 2}, "https://id.nomen.test", 1, time.Unix(100, 0))
	handler := NewHandler(overviewGetterStub{overview: projection}, authorizerStub{}, func(context.Context) string { return "instance-1" }, func(*http.Request) string { return "https://id.nomen.test" })
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/overview", nil))

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
	assert.Equal(t, "private, no-store", recorder.Header().Get("Cache-Control"))
	var got domain.Overview
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&got))
	assert.Equal(t, projection.ResourceRevision, got.ResourceRevision)
}

func TestHandlerPreservesTypedAuthenticationAndPermissionFailures(t *testing.T) {
	tests := []struct {
		name       string
		failure    domain.ManagementError
		wantStatus int
	}{
		{"authentication", authenticationError(), http.StatusUnauthorized},
		{"permission", permissionError(OverviewPermission), http.StatusForbidden},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := NewHandler(overviewGetterStub{}, authorizerStub{failure: &test.failure}, func(context.Context) string { return "instance-1" }, func(*http.Request) string { return "https://id.nomen.test" })
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/overview", nil))
			require.Equal(t, test.wantStatus, recorder.Code)
			var envelope domain.ManagementErrorEnvelope
			require.NoError(t, json.NewDecoder(recorder.Body).Decode(&envelope))
			require.NoError(t, envelope.Error.Validate())
			if test.wantStatus == http.StatusForbidden {
				assert.Equal(t, OverviewPermission, envelope.Error.RequiredPermission)
			}
		})
	}
}

func TestHandlerReturnsProtectedCapabilityDiscovery(t *testing.T) {
	now := time.Unix(100, 0)
	discovery := domain.CapabilityDiscovery{
		SchemaVersion:    1,
		ResourceRevision: "sha256:development",
		ObservedAt:       now,
		Components: []domain.ComponentCompatibility{
			{Role: domain.ComponentNomen, Version: domain.ProductVersion, APIMajor: 1, State: domain.CompatibilityUnknown, Reason: "development_bundle_unattested", ObservedAt: now},
		},
		Capabilities: []domain.CapabilityFact{
			{ID: domain.CapabilityIDLDAPInbound, Status: domain.CapabilityPreview, Exposure: domain.UIExposureDisabled, Reason: "conformance_pending", RequiredComponents: []domain.ComponentRole{domain.ComponentNomen}},
		},
	}
	handler := NewHandler(overviewGetterStub{}, authorizerStub{}, func(context.Context) string { return "instance-1" }, func(*http.Request) string { return "https://id.nomen.test" }).WithCapabilities(capabilityGetterStub{discovery: discovery})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/capabilities", nil))

	require.Equal(t, http.StatusOK, recorder.Code)
	var got domain.CapabilityDiscovery
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&got))
	assert.Equal(t, discovery.ResourceRevision, got.ResourceRevision)
}

func TestHandlerReturnsProtectedDeploymentPreflight(t *testing.T) {
	now := time.Date(2026, time.August, 22, 2, 0, 0, 0, time.UTC)
	preflight := domain.BuildDeploymentPreflight(domain.DeploymentPreflightFacts{
		DatabaseAvailable: true, DatabaseProbeAvailable: true,
		SigningKeys: 1, SigningProbeAvailable: true,
		NotificationConfigured: true, NotificationProbeAvailable: true,
	}, "https://identity.example.test", now)
	handler := NewHandler(overviewGetterStub{}, authorizerStub{}, func(context.Context) string { return "instance-1" }, func(*http.Request) string { return "https://identity.example.test" }).WithDeploymentPreflight(deploymentPreflightGetterStub{preflight: preflight})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/deployment/preflight", nil))

	require.Equal(t, http.StatusOK, recorder.Code)
	var got domain.DeploymentPreflight
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&got))
	assert.Equal(t, preflight.ResourceRevision, got.ResourceRevision)
}

func TestHandlerRequiresDeploymentPreflightPermission(t *testing.T) {
	failure := permissionError(PreflightPermission)
	handler := NewHandler(overviewGetterStub{}, authorizerStub{failure: &failure}, func(context.Context) string { return "instance-1" }, func(*http.Request) string { return "https://identity.example.test" }).WithDeploymentPreflight(deploymentPreflightGetterStub{})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/deployment/preflight", nil))

	require.Equal(t, http.StatusForbidden, recorder.Code)
	var envelope domain.ManagementErrorEnvelope
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&envelope))
	assert.Equal(t, PreflightPermission, envelope.Error.RequiredPermission)
}

func TestHandlerReturnsProtectedOwnerEnrollmentState(t *testing.T) {
	view := domain.BuildOwnerEnrollmentView(nil, time.Date(2026, time.August, 22, 4, 0, 0, 0, time.UTC))
	handler := NewHandler(overviewGetterStub{}, authorizerStub{}, func(context.Context) string { return "instance-1" }, func(*http.Request) string { return "https://identity.example.test" }).WithOwnerEnrollment(ownerEnrollmentGetterStub{view: view})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/deployment/owner-enrollment", nil))

	require.Equal(t, http.StatusOK, recorder.Code)
	var got domain.OwnerEnrollmentView
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&got))
	assert.Equal(t, domain.OwnerEnrollmentPending, got.State)
	assert.Empty(t, got.CeremonyID)
}

func TestHandlerReturnsFailClosedOperatorActionCatalog(t *testing.T) {
	now := time.Now().UTC()
	catalog := domain.OperatorActionCatalog{
		SchemaVersion: 1, ResourceRevision: "sha256:catalog", ObservedAt: now,
		Actions: []domain.OperatorAction{{
			ID: "action.provider_plan", Title: "Plan provider", Consequence: "Creates a reviewed provider plan.",
			Stage: domain.OperatorActionPlan, Method: http.MethodPost, Href: "/nomen/v1/providers:plan",
			IntentSchema: json.RawMessage(`{"type":"object"}`), RequiredPermissions: []string{"nomen.providers.plan"},
			CapabilityID: domain.CapabilityIDUpstreamOIDC, Exposure: domain.UIExposureDisabled, Reason: "conformance_pending",
		}},
	}
	handler := NewHandler(overviewGetterStub{}, authorizerStub{}, func(context.Context) string { return "instance-1" }, func(*http.Request) string { return "https://id.nomen.test" }).WithOperatorActions(operatorActionGetterStub{catalog: catalog})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/operator/actions", nil))

	require.Equal(t, http.StatusOK, recorder.Code)
	var got domain.OperatorActionCatalog
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&got))
	require.Len(t, got.Actions, 1)
	assert.Equal(t, domain.UIExposureDisabled, got.Actions[0].Exposure)
}

func TestHandlerRedactsProjectionFailure(t *testing.T) {
	handler := NewHandler(overviewGetterStub{err: errors.New("SELECT token='do-not-leak'")}, authorizerStub{}, func(context.Context) string { return "instance-1" }, func(*http.Request) string { return "https://id.nomen.test" })
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/overview", nil))

	require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
	assert.NotContains(t, recorder.Body.String(), "SELECT")
	assert.NotContains(t, recorder.Body.String(), "do-not-leak")
	var envelope domain.ManagementErrorEnvelope
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&envelope))
	require.NoError(t, envelope.Error.Validate())
}

func TestHandlerRejectsMethodWithTypedBody(t *testing.T) {
	handler := NewHandler(overviewGetterStub{}, authorizerStub{}, func(context.Context) string { return "instance-1" }, func(*http.Request) string { return "https://id.nomen.test" })
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/overview", nil))

	require.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
	assert.Equal(t, http.MethodGet, recorder.Header().Get("Allow"))
}

func TestHandlerRecordsValidatedOperatorEventWithServerActor(t *testing.T) {
	now := time.Now().UTC()
	body := []byte(`{"events":[{"schema_version":1,"event_id":"9a5af3d5-5456-4266-8b24-c0b954db6819","session_id":"ce46aa0d-bc16-42e7-9975-1476ec5734f8","sequence":1,"occurred_at":"` + now.Format(time.RFC3339Nano) + `","route_id":"route.federation","control_id":"control.provider_create","event_type":"control_activated","outcome":"observed"}]}`)
	repository := &operatorEventRecorderStub{}
	actor := domain.OperatorActor{InstanceID: "instance-1", TenantID: "tenant-1", ActorID: "owner-1"}
	handler := NewHandler(overviewGetterStub{}, authorizerStub{}, func(context.Context) string { return "instance-1" }, func(*http.Request) string { return "https://id.nomen.test" }).WithOperatorEvents(repository, func(context.Context) domain.OperatorActor { return actor })
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/operator/events", bytes.NewReader(body)))

	require.Equal(t, http.StatusAccepted, recorder.Code)
	assert.Equal(t, actor, repository.actor)
	require.Len(t, repository.batch.Events, 1)
	assert.Equal(t, domain.OperatorEventControlActivated, repository.batch.Events[0].EventType)
}

func TestHandlerRejectsSensitiveOperatorEventAttribute(t *testing.T) {
	now := time.Now().UTC()
	body := []byte(`{"events":[{"schema_version":1,"event_id":"9a5af3d5-5456-4266-8b24-c0b954db6819","session_id":"ce46aa0d-bc16-42e7-9975-1476ec5734f8","sequence":1,"occurred_at":"` + now.Format(time.RFC3339Nano) + `","route_id":"route.security","control_id":"control.secret","event_type":"control_activated","attributes":{"password":"canary"}}]}`)
	repository := &operatorEventRecorderStub{}
	handler := NewHandler(overviewGetterStub{}, authorizerStub{}, func(context.Context) string { return "instance-1" }, func(*http.Request) string { return "https://id.nomen.test" }).WithOperatorEvents(repository, func(context.Context) domain.OperatorActor {
		return domain.OperatorActor{InstanceID: "instance-1", TenantID: "tenant-1", ActorID: "owner-1"}
	})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/operator/events", bytes.NewReader(body)))

	require.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
	assert.Empty(t, repository.batch.Events)
	assert.NotContains(t, recorder.Body.String(), "canary")
}
