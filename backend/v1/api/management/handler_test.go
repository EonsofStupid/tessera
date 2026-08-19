package management

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/EonsofStupid/tessera/backend/v1/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type overviewGetterStub struct {
	overview domain.Overview
	err      error
}

func (s overviewGetterStub) Get(context.Context, string, string) (domain.Overview, error) {
	return s.overview, s.err
}

type authorizerStub struct{ failure *domain.ManagementError }

func (s authorizerStub) Authorize(r *http.Request, _ string) (*http.Request, *domain.ManagementError) {
	return r, s.failure
}

func TestHandlerReturnsValidatedOverview(t *testing.T) {
	projection := domain.BuildOverview(domain.OverviewFacts{Flows: 1, HumanSeats: 2}, "https://id.shippin.ai", 1, time.Unix(100, 0))
	handler := NewHandler(overviewGetterStub{overview: projection}, authorizerStub{}, func(context.Context) string { return "instance-1" }, func(*http.Request) string { return "https://id.shippin.ai" })
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
			handler := NewHandler(overviewGetterStub{}, authorizerStub{failure: &test.failure}, func(context.Context) string { return "instance-1" }, func(*http.Request) string { return "https://id.shippin.ai" })
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

func TestHandlerRedactsProjectionFailure(t *testing.T) {
	handler := NewHandler(overviewGetterStub{err: errors.New("SELECT token='do-not-leak'")}, authorizerStub{}, func(context.Context) string { return "instance-1" }, func(*http.Request) string { return "https://id.shippin.ai" })
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
	handler := NewHandler(overviewGetterStub{}, authorizerStub{}, func(context.Context) string { return "instance-1" }, func(*http.Request) string { return "https://id.shippin.ai" })
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/overview", nil))

	require.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
	assert.Equal(t, http.MethodGet, recorder.Header().Get("Allow"))
}
