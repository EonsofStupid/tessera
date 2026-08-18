// Package flows is the executor's HTTP surface.
//
//	POST /flows/v1/{slug}/start              → { execution_id, token, challenge }
//	POST /flows/v1/executions/{executionID}  → next challenge | { session_id, session_token, user_id }
//
// It translates and decides nothing: which stages run is the flow's
// declaration, how answers are judged is the runners', and both live behind
// backend/v1's ports. What this layer owns is HTTP truth — status codes that
// do not lie, and a token check that does not leak.
package flows

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/EonsofStupid/tessera/backend/v1/domain"
	flowstorage "github.com/EonsofStupid/tessera/backend/v1/storage/flow"
	"github.com/EonsofStupid/tessera/internal/api/authz"
	"github.com/EonsofStupid/tessera/internal/command"
	"github.com/EonsofStupid/tessera/internal/query"
)

const (
	HandlerPrefix = "/flows/v1"

	// SystemRole is the role the handler's in-process calls carry. It is
	// registered in start.go against exactly the permissions a login client
	// holds — session.read and session.write — because that is what a flow
	// executor is: the login client, running in-process. Not SYSTEM_OWNER;
	// a login path with instance-delete permissions is an incident report
	// with a delay on it.
	SystemRole = "TESSERA_FLOWS"
	// SystemUserID names these calls in the audit trail.
	SystemUserID = "TESSERA_FLOWS"
)

// Permissions is what SystemRole must map to; start.go registers it.
var Permissions = []string{"session.read", "session.write"}

// Config is the handler's tuning.
type Config struct {
	// ExecutionTTL bounds a login attempt, start to finish.
	ExecutionTTL time.Duration
	// SessionLifetime is what a completed flow's session lives.
	SessionLifetime time.Duration
}

type Handler struct {
	router   chi.Router
	flows    domain.FlowRepository
	execs    domain.ExecutionRepository
	executor *domain.FlowExecutor
}

// NewHandler wires the executor over the given repositories and command layer.
func NewHandler(cfg Config, commands *command.Commands, queries *query.Queries, repo *flowstorage.Repository) *Handler {
	runners := flowstorage.NewRunners(flowstorage.StageDeps{
		Commands:        commands,
		Queries:         queries,
		SessionLifetime: cfg.SessionLifetime,
	})
	h := &Handler{
		flows:    repo,
		execs:    repo,
		executor: domain.NewFlowExecutor(cfg.ExecutionTTL, runners...),
	}
	r := chi.NewRouter()
	r.Post("/{slug}/start", h.start)
	r.Post("/executions/{executionID}", h.advance)
	h.router = r
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) { h.router.ServeHTTP(w, r) }

// privileged returns the request context authorized for the handler's own
// command calls. The customer on the other end of the connection has no
// permissions — that is the point of a login flow — so the executor acts as
// the system login client, under its narrowly-mapped role, with a user id
// that names it in the audit trail.
func privileged(r *http.Request) *http.Request {
	ctx := authz.SetCtxData(r.Context(), authz.CtxData{
		UserID: SystemUserID,
		SystemMemberships: authz.Memberships{{
			MemberType: authz.MemberTypeSystem,
			Roles:      []string{SystemRole},
		}},
	})
	return r.WithContext(ctx)
}

type startResponse struct {
	ExecutionID string            `json:"execution_id"`
	Token       string            `json:"token"`
	ExpiresAt   time.Time         `json:"expires_at"`
	Challenge   *domain.Challenge `json:"challenge"`
}

func (h *Handler) start(w http.ResponseWriter, r *http.Request) {
	r = privileged(r)
	ctx := r.Context()
	instanceID := authz.GetInstance(ctx).InstanceID()
	slug := chi.URLParam(r, "slug")

	flow, found, err := h.flows.FlowBySlug(ctx, instanceID, slug)
	if err != nil {
		internalError(w, err)
		return
	}
	if !found {
		writeJSON(w, http.StatusNotFound, errBody("unknown_flow", "no flow is declared at this slug"))
		return
	}
	if flow.Designation != domain.DesignationAuthentication && flow.Designation != domain.DesignationRecovery {
		writeJSON(w, http.StatusNotFound, errBody("unknown_flow", "this flow cannot be started here"))
		return
	}

	exec, challenge, err := h.executor.Start(ctx, instanceID, flow)
	if err != nil {
		internalError(w, err)
		return
	}
	if err := h.execs.SaveExecution(ctx, exec); err != nil {
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, startResponse{
		ExecutionID: exec.ID,
		Token:       exec.Token,
		ExpiresAt:   exec.ExpiresAt,
		Challenge:   challenge,
	})
}

type advanceRequest struct {
	Token  string         `json:"token"`
	Answer map[string]any `json:"answer"`
}

type doneResponse struct {
	SessionID    string `json:"session_id"`
	SessionToken string `json:"session_token"`
	UserID       string `json:"user_id"`
}

func (h *Handler) advance(w http.ResponseWriter, r *http.Request) {
	r = privileged(r)
	ctx := r.Context()
	instanceID := authz.GetInstance(ctx).InstanceID()
	id := chi.URLParam(r, "executionID")

	var req advanceRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "the body is not the JSON this endpoint takes"))
		return
	}

	exec, found, err := h.execs.ExecutionByID(ctx, instanceID, id)
	if err != nil {
		internalError(w, err)
		return
	}
	// Unknown id and wrong token answer identically, on purpose: an attacker
	// holding a leaked execution id learns nothing about whether it is live.
	if !found || subtle.ConstantTimeCompare([]byte(exec.Token), []byte(req.Token)) != 1 {
		writeJSON(w, http.StatusNotFound, errBody("unknown_execution", "no such execution — start the flow again"))
		return
	}

	challenge, err := h.executor.Advance(ctx, exec, req.Answer)
	switch {
	case errors.Is(err, domain.ErrExecutionExpired):
		_ = h.execs.DeleteExecution(ctx, instanceID, id)
		writeJSON(w, http.StatusGone, errBody("expired", "this sign-in attempt expired — start the flow again"))
		return
	case errors.Is(err, domain.ErrExecutionDone):
		writeJSON(w, http.StatusConflict, errBody("done", "this execution already completed"))
		return
	case err != nil:
		internalError(w, err)
		return
	}

	if exec.Done() {
		// Completion consumes the execution: the session token crosses the
		// wire exactly once, and the id cannot be answered again.
		if err := h.execs.DeleteExecution(ctx, instanceID, id); err != nil {
			internalError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, doneResponse{
			SessionID:    exec.SessionID,
			SessionToken: exec.SessionToken,
			UserID:       exec.UserID,
		})
		return
	}

	// The stage may have accumulated state even when it re-asks (it did not
	// here — but position moves on success and that must persist before the
	// client sees the next challenge, or a crashed handler forgets a passed
	// factor the customer watched succeed).
	if err := h.execs.UpdateExecution(ctx, exec); err != nil {
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, challenge)
}

func errBody(code, msg string) map[string]string {
	return map[string]string{"error": code, "error_description": msg}
}

func internalError(w http.ResponseWriter, err error) {
	// The description names the class, never the internals — the details are
	// ours to read in the logs, not the network's.
	writeJSON(w, http.StatusInternalServerError, errBody("internal", "something on our side failed; it is logged"))
	logErr(err)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
