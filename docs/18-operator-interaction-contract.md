# 18 — Operator interaction and AI action contract

**Status:** accepted v1 contract; execution APIs remain preview until their
durable audit and authorization suites pass.
**Builds on:** `14-operation-contract.md`,
`15-management-error-contract.md` and
`16-capability-discovery-contract.md`.

## Purpose

Tessera has one operator model for humans, automation and dedicated AI
specialists. A browser click is not privileged because it came from a human,
and an AI does not gain a hidden administrative surface. Both discover the
same typed actions, plan the same effects, satisfy the same authorization and
step-up requirements, and produce the same audit evidence.

The semantic event stream answers what an operator intended, which public
control or API action they used, what resource revision they observed, and
whether the server accepted or refused it. It complements diagnostics and
optional session replay; pixels and DOM recordings are never the source of
truth for an identity decision.

## Stable identifiers

Every route, guide, control, suggestion and action has a stable, documented id.
Display text may be localized; automation never selects a control by visible
label or screen coordinates.

Examples:

```text
route.federation
control.provider.create
guide.application.oidc
suggestion.redirect.local_development
action.provider.plan_create
```

Renaming a visible label does not change its id. Removing an id is a versioned
compatibility change.

## Semantic runtime events

`POST /tessera/v1/operator/events` accepts a bounded batch. Each event carries:

- `schema_version`, currently `1`;
- a client-generated UUID `event_id` for deduplication;
- `occurred_at` and a monotonic sequence within the browser session;
- stable `route_id`, `control_id`, `event_type` and optional `action_id`;
- safe state such as selected tab, capability state and resource revision;
- a correlation id returned by a plan or operation; and
- outcome `observed`, `accepted`, `refused` or `failed` when one exists.

The server derives actor, tenant, deployment, session and authorization context
from authentication. A browser cannot assert them. Unknown fields, excessive
batch size, stale timestamps, duplicate sequence values and values outside the
public attribute allowlist are refused with a typed error.

Forbidden event data includes passwords, tokens, assertions, cookies, secret
values, recovery material, authenticator answers, raw form bodies, free-form
prompts and customer document content. Event schemas use allowlists rather than
best-effort redaction.

The accepted event and its outbox record are one PostgreSQL transaction.
ClickHouse receives a tenant-scoped analytical projection asynchronously. Loss
of analytics never blocks authentication or an authorized management action.

## AI action surface

`GET /tessera/v1/operator/actions` publishes the actions available to the
authenticated principal. Every entry defines:

- stable action id and JSON Schema for intent;
- required permissions, assurance and capability ids;
- whether it is a read, plan, apply, verify or cancel operation;
- consequence, reversibility and protected-input slots;
- current resource revision and safe seed suggestions; and
- links to terminology and operator guidance.

Mutation tools do not accept arbitrary browser events as commands. They call
the versioned management resource named by the action. High-impact work uses
the immutable plan digest, explicit review, idempotency key, step-up and verify
grammar from contract 14. AI-originated operations carry an actor-chain entry
that identifies the requesting human or service, the specialist identity and
the model/tool version without storing a prompt.

AI may draft and explain. It cannot silently approve its own high-impact plan,
weaken assurance, reveal a protected value, cross a tenant boundary or convert
a preview capability into an operational one.

## JSON-guided presentation

Guides and teaching cards are server resources with a versioned JSON schema.
They may select only registered Tessera components and design tokens. The
schema supports:

- terminology, consequence-first explanation and progressive detail;
- typed inputs with safe seed suggestions and source labels;
- capability, permission and deployment-profile conditions;
- reviewed plan summaries and verification evidence;
- color, illustration and reduced-motion-safe animation tokens; and
- action ids, never executable script or arbitrary markup.

Seed suggestions are examples or derived public values, not hidden defaults.
The UI shows why a suggestion is safe, validates it locally for responsiveness,
and treats the server as the final authority.

## Session replay boundary

Replay is optional and off by default. Enabling it requires an explicit tenant
policy, purpose, retention period and role. Authentication, recovery, secret,
token, assertion and protected-reference surfaces are blocked from capture at
the component boundary. Input values are masked by default; allowlisting a
non-sensitive field is a reviewed schema change.

Replay chunks are tenant-scoped, encrypted, access-audited and deleted by
retention policy. Centralized hosting may replace local replay storage only
through the authority contract; disabling local storage must be verified before
central ingestion is accepted.

## Done when

- every interactive control in `@tessera/ui` has a stable id and emits a schema-
  valid semantic event;
- event ingest is authenticated, tenant-derived, idempotent, durable and backed
  by an outbox;
- secret canaries never appear in semantic events, logs, replay or analytics;
- a human and an AI specialist produce equivalent plans and operation evidence;
- JSON guides reject unknown components, scripts, unsafe suggestions and
  capability promotion;
- replay privacy, access, deletion and cross-tenant negative suites pass; and
- an operator can reconstruct the intent and outcome of every management
  mutation without relying on session replay.
