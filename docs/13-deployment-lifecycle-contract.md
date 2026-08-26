# 13 — Deployment lifecycle contract

**Status:** accepted v1 contract for P1.1.
**Builds on:** the lifecycle quality bar in
`11-operational-blueprint.md`.

## Purpose

One server-owned state describes whether a Nomen deployment may accept
ordinary identity mutations. The Nomen management application renders that
state and its evidence; optional host adapters may project it but no client
infers lifecycle from a process check, an HTTP status or the last button
somebody clicked.

This is deployment lifecycle, not user, session, federation or billing state.

## States

| state | contract meaning |
|---|---|
| `absent` | no Nomen instance is registered |
| `preparing` | prerequisites are being checked; no durable initialization has begun |
| `initializing` | database and first instance initialization owns deployment writes |
| `needs_owner` | core service is healthy; owner enrollment and recovery are incomplete |
| `ready` | every required readiness check currently meets policy |
| `degraded` | service is available but at least one required capability is unhealthy |
| `maintenance` | an upgrade, restore or rotation operation owns ordinary mutations |
| `recovery_required` | ordinary mutations are stopped until a guided recovery operation succeeds |
| `retired` | sign-in is disabled and retention/export policy owns remaining state |

`ready` is never a stored synonym for “the process answered.” Its projection
must be backed by the contributing checks defined by P2.2.

## Transition graph

A transition to the current state is always valid and changes nothing. This is
the lifecycle half of idempotency: retrying an operation after losing its
response cannot fail merely because the first request reached the intended
state.

Every state-changing transition is listed below. Anything not listed is
refused.

| from | may change to | why |
|---|---|---|
| `absent` | `preparing` | installation starts with read-only preflight |
| `preparing` | `absent`, `initializing` | cancel before durable mutation, or begin initialization |
| `initializing` | `needs_owner`, `degraded`, `recovery_required` | initialize successfully, remain diagnosable, or stop for recovery |
| `needs_owner` | `ready`, `degraded`, `recovery_required`, `retired` | finish ownership, expose a failed dependency, recover, or retire before use |
| `ready` | `degraded`, `maintenance`, `recovery_required`, `retired` | health loss, planned maintenance, emergency stop or retirement |
| `degraded` | `ready`, `maintenance`, `recovery_required`, `retired` | recover health, repair under maintenance, stop writes or retire |
| `maintenance` | `ready`, `degraded`, `recovery_required`, `retired` | verify success, expose residual damage, require recovery or retire |
| `recovery_required` | `maintenance`, `retired` | recovery must own writes before readiness can be re-established |
| `retired` | `absent` | complete retention/export and remove the registered deployment |

Important refusals are deliberate:

- `absent` cannot jump directly to `ready`; setup evidence cannot be skipped.
- `needs_owner` cannot enter maintenance; ownership and independent recovery
  must exist before routine upgrades or rotations.
- `recovery_required` cannot jump directly to `ready`; a recovery operation
  enters `maintenance`, performs work, then verifies the resulting posture.
- `retired` cannot be reactivated. A new deployment starts again at `absent`.

Authorization is evaluated separately. A transition being valid does not mean
the caller may perform it.

## Typed refusals

The domain returns `DeploymentTransitionError` with one stable reason:

| reason | meaning | expected API mapping in P1.3 |
|---|---|---|
| `unknown_current_state` | stored/current state is not in this contract | typed `500`; stop mutations and repair state |
| `unknown_target_state` | caller requested a state outside this contract | typed `422`; correct the request |
| `transition_not_allowed` | both states exist but the edge is absent | typed `409`; re-read and plan a valid operation |

Messages are diagnostic; consumers branch on the reason. Each error also
carries the current and target values.

## Domain API

```go
ValidateDeploymentTransition(current, target DeploymentState) error
AllowedDeploymentTransitions(current DeploymentState) ([]DeploymentState, error)
```

`ValidateDeploymentTransition` returns nil for an allowed change and for the
same valid state. `AllowedDeploymentTransitions` includes the current state as
the first value so a planner can represent an achieved idempotent target.

The canonical order is the order in the state table above. API schemas and UI
choices reuse that order rather than sorting strings independently.

## Done when

- every pair of known states is tested as allowed or refused;
- same-state retries are accepted for all known states;
- unknown current and target values produce their distinct typed reasons;
- allowed-transition discovery agrees with validation;
- no transport, database or inherited provider package is imported into the
  lifecycle domain.
