# 07 — Flows: the plan

**Status:** build plan for Phase 4. The model is Authentik's
(`authentik/flows/planner.py`, `challenge.py`, `stage.py`); the implementation
is ours, in Go, over `backend/v1`.

## What a flow is here

An ordered set of stages, declared as a blueprint, executed one challenge at a
time. The client — panel, CLI, curl — receives a component-tagged JSON
challenge, answers it, and gets the next one, until the flow completes into a
Zitadel session with every answered factor checked. Identity, MFA and recovery
stop being three code paths and become three YAML files.

```
POST /flows/v1/{slug}/start                → { execution_id, token, challenge }
POST /flows/v1/executions/{execution_id}   → next challenge | { session_id, … }
```

## Decisions, and why they are not preferences

**Stages delegate; they never verify.** The password stage drives
`command.CheckPassword`, TOTP drives `CheckTOTP`, recovery codes drive
`CheckRecoveryCode` — the same `SessionCommand`s the session v2 API uses. The
flow engine owns *sequence*; Zitadel's command layer owns *truth*. A stage
that reimplemented password hashing would be a second implementation
authority, which is the exact thing this repository exists to prevent.

**A flow is a blueprint model.** `tessera/flow` entries land in
`tessera.flows` + `tessera.flow_stages` through the same engine as seats:
login configuration arrives as reviewed YAML, applies atomically, converges to
a no-op, and restores itself on boot. Phase 3 was built so Phase 4 could be
declared.

**The session accumulates incrementally.** `identify` creates the session with
`CheckUser`; each factor stage updates it with its check as it is answered.
Every factor is verified the moment it is submitted — a wrong password fails
at the password stage with a field error, not at the end as a heap of
failures nobody can attribute.

**Execution state lives on the server; the client holds two opaque strings.**
The plan, its position and the session id sit in `tessera.flow_executions`;
the client carries `execution_id` plus a per-execution token required on every
answer, so knowing someone's execution id is not the ability to answer their
challenges. Executions expire; an expired one says so and points at `start`.

**Challenges are component-tagged, like Authentik's.** `component:
"tessera-stage-password"` plus typed fields. Self-describing to any renderer,
and the panel's custom UI — the actual product — renders against components,
not against endpoints.

**The planner is deliberately small, and the seam is named.** v1 planning is:
take the flow's stages in order. Authentik's policy-per-binding machinery and
re-evaluation markers are real and deferred — the seam is `Planner.Plan`, and
policies arrive there when something needs them, not before.

## Where it lives

| piece | path |
|---|---|
| Flow, Stage, Plan, Challenge, executor semantics | `backend/v1/domain/flow*.go` |
| migration 002: flows, flow_stages, flow_executions | `backend/v1/storage/migration` |
| flow repository + `tessera/flow` applier | `backend/v1/storage/flow` |
| stage implementations over `internal/command` | `backend/v1/storage/flow` (the one place allowed to import command) |
| the executor HTTP API | `internal/api/flows` (mounted like the idp handler) |
| the three flows, as files | `blueprints/dev/*.yaml` |

## Sequence

**4.1 — the flow domain.** Flow/Stage/Plan/Challenge types, validation in the
blueprint style (every refusal a named error), and the executor state machine
proven over fake stages: advance on success, field-error and stay on failure,
expire, complete. *Done when* the executor's semantics are fully tested with
no database and no HTTP.

**4.2 — flows are declared.** Migration 002, the flow repository, the
`tessera/flow` applier registered beside seats. *Done when* a flow blueprint
applies atomically, re-applies as a no-op, and comes back on boot — the same
three proofs seats have, extended to the second model.

**4.3 — real stages and the HTTP surface.** identify/password/totp/
recovery-code stages delegating to `SessionCommand`s; `/flows/v1` handler
wired in `start.go`. *Done when* a session created through a flow carries the
same factors as one created through the session v2 API.

**4.4 — three configurations, one engine.** `login-password.yaml`,
`login-mfa.yaml`, `recovery.yaml` in `blueprints/dev`; `dev/flow-probe.sh`
drives the password flow end to end over HTTP against a human user and
asserts the session's password factor. *Done when* the roadmap's sentence is
literally true on this box: password, TOTP and recovery are three YAML files
executed by one engine.

## Out of scope, on purpose

Policy-gated bindings and re-evaluation markers (the planner seam is where
they land), enrollment and invalidation designations, flow import from
Authentik's schema, WebAuthn/passkey stages (the check exists upstream; the
stage joins when a consumer needs it), and Phase 6's full account-recovery
design — the recovery *flow* here proves the engine executes the
configuration; recovering a lost second factor safely is its own document.
