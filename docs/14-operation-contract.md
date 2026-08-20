# 14 — Plan, apply and verify operation contract

**Status:** accepted v1 contract for P1.2.
**Builds on:** `11-operational-blueprint.md` and
`13-deployment-lifecycle-contract.md`.

## One operation grammar

Installation, guided setup, backup, restore, upgrade and trust rotation use the
same grammar:

```text
intent → immutable plan → reviewed digest → apply → verify → terminal result
```

Domain-specific APIs own their intent and verification checks. The operation
contract owns review integrity, idempotency, progress, retry and stale-state
behavior. A new guide may add effects; it may not invent a different way to
apply them.

## Resource split

### Plan

A plan is non-mutating and immutable. It contains:

- `plan_id` and operation kind;
- account and optional workspace scope;
- `base_revision`, the exact relevant state that was reviewed;
- `desired_revision`, the deterministic desired state after the plan;
- `plan_digest`, covering scope, revisions, effects, requirements, permissions
  and planned verification without secret values;
- ordered creates, updates and removals with stable effect ids;
- prerequisites and protected-secret slots;
- permissions required by apply;
- outcome checks that verification will execute;
- creation and expiration times.

The base revision is never empty. A new installation uses the explicit
revision representing `absent`; empty cannot mean both “absent” and “the caller
forgot to provide a revision.”

Planning may read live state and perform safe discovery. It never writes an
identity resource, accepts a secret value or reserves success.

### Operation

Apply creates or resolves one durable operation before the first side effect.
An operation contains:

- `operation_id`, `plan_id`, kind and scope;
- current phase and status;
- copied base, desired and plan revisions;
- monotonic progress events;
- whether cancellation is currently safe;
- creation/update/completion times;
- a typed terminal failure when unsuccessful.

The accepted plan remains available with the operation even after its ordinary
review expiration. An operator can understand what ran without reconstructing
the old projection.

## Apply preconditions

Apply requires all of these before any mutation:

1. a syntactically valid idempotency key;
2. an existing, unexpired plan;
3. the caller-provided reviewed digest exactly matching `plan_digest`;
4. current relevant state exactly matching `base_revision`;
5. every required permission;
6. every prerequisite either satisfied or explicitly resolved by an accepted
   protected reference;
7. step-up evidence when the plan marks a high-impact effect.

Digest mismatch is not repaired by returning the new plan. It is refused so the
customer sees and reviews the changed effects. A stale base revision is also
refused before writes and must be replanned.

## Idempotency

Every plan, apply and explicit verify request carries `Idempotency-Key` at the
HTTP boundary and the same value in non-HTTP transports.

The key is scoped to caller identity, account, route and operation kind. It is
stored as a one-way digest, not treated as an authentication credential and not
returned in reads or logs.

- Same scope, key and canonical request returns the original plan, operation or
  verification result.
- Same scope and key with a different canonical request returns the typed
  conflict `idempotency_key_reused`.
- Apply records the operation before side effects. Losing the response and
  retrying resolves that operation rather than starting another.
- Idempotency records live at least as long as the associated resource and
  audit retention. Expiry is never shorter than a client retry window.

A key is 1–255 visible ASCII characters with no whitespace or control bytes.
That constraint keeps it safe across HTTP, gRPC metadata and logs that record
only its digest.

## Revisions and digest

Revisions are opaque, stable tokens. Clients compare and echo them; they do not
parse or increment them.

`desired_revision` is deterministic for the same canonical intent and base
state. `plan_digest` is a SHA-256 digest of canonical plan content and uses the
wire spelling `sha256:<lowercase hex>`. It excludes:

- secret values and secret-provider responses;
- timestamps that do not change effects;
- diagnostics or localized presentation;
- progress and verification observations made after planning.

The canonicalization implementation lands with the first planner. Until then,
the contract tests validate the digest shape and apply comparison semantics.

## Progress

Progress is an append-only sequence. Each event carries:

- a sequence beginning at 1 and increasing by exactly one;
- phase: `plan`, `apply` or `verify`;
- status: `queued`, `running`, `succeeded`, `failed`, `canceling` or `canceled`;
- a stable step code and consequence-first summary;
- completed and total work units only when the total is knowable;
- retry directive: `none`, `same_request`, `replan` or `operator_action`;
- an observation time and safe diagnostic reference.

Phases never move backward. No event follows terminal `succeeded`, `failed` or
`canceled`. A percentage is derived only when total work is known; the server
does not invent one for indeterminate external waits.

Diagnostics contain identifiers and remediation, never secret values,
authentication answers or raw provider assertions.

## Completion and verification

Apply success advances to verification; it does not finish the operation.
Terminal `succeeded` means every required outcome check passed. A green database
write followed by a failed discovery, sign-in, token or restore check is a
failed operation in phase `verify`, with the resulting lifecycle state decided
by `13-deployment-lifecycle-contract.md`.

Verification may be retried through the same operation when its retry directive
allows `same_request`. A verification that observes changed source state
requires `replan` rather than silently blessing a different outcome.

## Conflicts and refusals

P1.3 assigns HTTP/gRPC mappings. P1.2 establishes the stable operation reasons:

| reason | category | remedy |
|---|---|---|
| `invalid_plan` | invalid request/record | repair the named plan field |
| `plan_expired` | stale review | plan and review again |
| `plan_digest_mismatch` | review conflict | fetch/replan and review exact effects |
| `stale_base_revision` | state conflict | replan from current state |
| `idempotency_key_required` | invalid request | supply a key |
| `invalid_idempotency_key` | invalid request | use visible non-whitespace ASCII, max 255 |
| `idempotency_key_reused` | request conflict | use the original request or a new key |
| `progress_sequence_invalid` | corrupt operation state | stop the worker and repair from audit |
| `progress_phase_regressed` | corrupt operation state | stop the worker and repair from audit |
| `operation_terminal` | state conflict | read the terminal result; do not append work |

## Wire schema

`proto/tessera/management/v1/operation.proto` is the provider-neutral wire
vocabulary. Domain-specific plan endpoints reuse `OperationPlan`; apply and
verify requests use the shared id and digest fields. The standalone browser
reaches these resources through Tessera's same-origin management API. External
panels may use a server-side adapter with the same wire contract; an adapter is
never a product prerequisite.

P1.4 capability discovery declares which operation kinds and schema versions a
deployment supports.

## Done when

- valid apply preconditions and every refusal reason are tested;
- progress sequencing, phase monotonicity and terminal behavior are tested;
- protobuf lint and generation pass from tracked source;
- the domain package imports no transport, database or inherited provider code;
- source provenance classifies the new Tessera-owned protocol separately from
  the inherited compatibility protocol tree.
