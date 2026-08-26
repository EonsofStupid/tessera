package flow

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/shippinAI/nomen/backend/v1/domain"
	"github.com/shippinAI/nomen/internal/command"
	zi_domain "github.com/shippinAI/nomen/internal/domain"
	"github.com/shippinAI/nomen/internal/query"
	"github.com/shippinAI/nomen/internal/zerrors"
)

// The stage runners. This file is the one place in backend/v1 that imports
// internal/command, and that is its entire job: a stage owns the *ask* — the
// challenge, the field, the wording — and delegates the *judgment* to the same
// SessionCommands the session v2 API uses. A session built by a flow carries
// the same factor events as one built through the API, because it is built by
// the same code.

// StageDeps is what the runners need. The context they receive must already
// carry the privileged ctx data the API layer prepares — runners never
// escalate on their own.
type StageDeps struct {
	Commands *command.Commands
	Queries  *query.Queries
	// SessionLifetime is set once, at session creation. Factor updates pass
	// zero, which the command layer reads as "keep what it has".
	SessionLifetime time.Duration
}

// NewRunners returns every stage implementation, ready for the executor.
func NewRunners(deps StageDeps) []domain.StageRunner {
	return []domain.StageRunner{
		&identifyRunner{deps: deps},
		&factorRunner{deps: deps, kind: domain.StagePassword, field: "password", fieldType: "password",
			label:  "Password",
			reject: "that password was not accepted",
			check:  func(answer string) command.SessionCommand { return command.CheckPassword(answer) }},
		&factorRunner{deps: deps, kind: domain.StageTOTP, field: "code", fieldType: "code",
			label:  "Authenticator code",
			reject: "that code was not accepted",
			check:  func(answer string) command.SessionCommand { return command.CheckTOTP(answer) }},
		&factorRunner{deps: deps, kind: domain.StageRecoveryCode, field: "code", fieldType: "code",
			label:  "Recovery code",
			reject: "that code was not accepted",
			check:  func(answer string) command.SessionCommand { return command.CheckRecoveryCode(answer) }},
	}
}

// answerString reads one required field out of a submission.
func answerString(answer map[string]any, field string) (string, bool) {
	v, _ := answer[field].(string)
	v = strings.TrimSpace(v)
	return v, v != ""
}

// isAnswerFailure separates "the answer was wrong" from "we are broken".
//
// The command layer throws typed errors: a wrong password or code is an
// invalid-argument, a locked account or missing factor a failed precondition,
// an unknown user a not-found. All of those are the caller's situation, not
// our failure — they become field errors and the stage re-asks. Anything else
// (the eventstore down, a context cancelled) surfaces as the infrastructure
// error it is. Conflating the two tells a customer their password is wrong
// when the database is down, and tells the operator nothing at all.
func isAnswerFailure(err error) bool {
	return zerrors.IsErrorInvalidArgument(err) ||
		zerrors.IsPreconditionFailed(err) ||
		zerrors.IsNotFound(err)
}

// ---- identify --------------------------------------------------------------

type identifyRunner struct{ deps StageDeps }

func (r *identifyRunner) Kind() domain.StageKind { return domain.StageIdentify }

func (r *identifyRunner) Challenge(_ context.Context, exec *domain.Execution, _ domain.FlowStage) (*domain.Challenge, error) {
	return &domain.Challenge{
		Component: "nomen-stage-identify",
		Flow:      exec.Plan.FlowSlug,
		Title:     "Sign in",
		Fields:    []domain.ChallengeField{{Name: "identifier", Type: "text", Label: "Login name", Required: true}},
	}, nil
}

func (r *identifyRunner) Answer(ctx context.Context, exec *domain.Execution, _ domain.FlowStage, answer map[string]any) (map[string][]string, error) {
	identifier, ok := answerString(answer, "identifier")
	if !ok {
		return map[string][]string{"identifier": {"a login name is required"}}, nil
	}

	// One deliberately vague message for every identify failure. "No such
	// user" and "user found" answered differently is an account-enumeration
	// oracle, and login pages are where those get harvested.
	const rejected = "we could not sign you in with that"

	user, err := r.deps.Queries.GetUserByLoginName(ctx, true, identifier)
	if err != nil {
		if isAnswerFailure(err) {
			return map[string][]string{"identifier": {rejected}}, nil
		}
		return nil, err
	}

	set, err := r.deps.Commands.CreateSession(ctx,
		[]command.SessionCommand{command.CheckUser(user.ID, user.ResourceOwner, nil)},
		nil, &zi_domain.UserAgent{}, r.deps.SessionLifetime)
	if err != nil {
		if isAnswerFailure(err) {
			return map[string][]string{"identifier": {rejected}}, nil
		}
		return nil, err
	}

	exec.UserID = user.ID
	exec.ResourceOwner = user.ResourceOwner
	exec.SessionID = set.ID
	exec.SessionToken = set.NewToken
	return nil, nil
}

// ---- the factor stages -----------------------------------------------------

// factorRunner is password, TOTP and recovery-code in one type, because they
// ARE one shape: ask for a secret, hand it to the command layer's check, and
// keep the session token the update returns. The differences — field name,
// wording, which SessionCommand — are configuration. That the three factors
// collapse to data here is the same argument the whole phase makes about
// flows.
type factorRunner struct {
	deps      StageDeps
	kind      domain.StageKind
	field     string
	fieldType string
	label     string
	reject    string
	check     func(answer string) command.SessionCommand
}

func (r *factorRunner) Kind() domain.StageKind { return r.kind }

func (r *factorRunner) Challenge(_ context.Context, exec *domain.Execution, _ domain.FlowStage) (*domain.Challenge, error) {
	return &domain.Challenge{
		Component: "nomen-stage-" + string(r.kind),
		Flow:      exec.Plan.FlowSlug,
		Title:     r.label,
		Fields:    []domain.ChallengeField{{Name: r.field, Type: r.fieldType, Label: r.label, Required: true}},
	}, nil
}

func (r *factorRunner) Answer(ctx context.Context, exec *domain.Execution, _ domain.FlowStage, answer map[string]any) (map[string][]string, error) {
	if exec.SessionID == "" {
		// Validation puts identify first, so reaching here without a session
		// is a bug in the engine, not a bad answer.
		return nil, fmt.Errorf("flow: %s stage reached with no session — identify did not run", r.kind)
	}
	secret, ok := answerString(answer, r.field)
	if !ok {
		return map[string][]string{r.field: {"required"}}, nil
	}

	set, err := r.deps.Commands.UpdateSession(ctx, exec.SessionID,
		[]command.SessionCommand{r.check(secret)}, nil, 0)
	if err != nil {
		if isAnswerFailure(err) {
			return map[string][]string{r.field: {r.reject}}, nil
		}
		return nil, err
	}
	if set.NewToken != "" {
		exec.SessionToken = set.NewToken
	}
	return nil, nil
}
