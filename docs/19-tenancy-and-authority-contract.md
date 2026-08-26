# 19 — Tenancy, data ownership and authority contract

**Status:** accepted architecture contract; migration and negative conformance
remain active work.
**Builds on:** `02-standalone-product-contract.md` and
`18-operator-interaction-contract.md`.

## Product-owned storage

Every new Nomen-owned PostgreSQL table, index, trigger, policy, publication
and ClickHouse projection uses the `nomen_` prefix. New unprefixed objects are
rejected in review. Existing compatibility objects are migrated behind
Nomen-owned repositories; runtime code may not introduce a second direct
dependency on them.

PostgreSQL is identity truth. Every tenant-scoped table has a non-null
`tenant_id`, primary or unique keys include tenant context where identity is not
deployment-global, and foreign keys cannot connect rows across tenants.
Deployment-global records are explicitly classified and never inferred from a
missing tenant.

## Row-level security

Community mode uses PostgreSQL row-level security as defense in depth in
addition to application authorization. A transaction sets immutable local
context for tenant, actor and operation; pooled connections clear it before
reuse. A query with absent or malformed tenant context sees no tenant rows and
cannot mutate them.

Table owners do not serve application traffic. Runtime roles cannot bypass RLS.
Migration, backup and break-glass roles are separate, time-bounded where
possible and fully audited. Dedicated mode runs the same policies and tests so
the two profiles do not become different products.

Caches, job queues, idempotency keys, outbox records, object storage, analytics
and replay all include tenant context. Tenant ids come from authenticated
server context, never a trusted browser field.

## Authority modes

Each configurable resource declares exactly one authority mode:

- `local` — the Nomen deployment accepts authorized local management writes;
- `central` — a named control plane supplies signed desired revisions and local
  mutation for that resource is disabled with typed remediation; or
- `delegated` — central policy defines a bounded field set that local owners may
  change without overriding central fields.

Authority is per resource family, not one vague deployment flag. Changes use a
reviewed migration plan that names the final local revision, initial central
revision, conflict policy, rollback boundary and recovery owner. Nomen never
merges two simultaneous writers by last-write-wins.

A central command is authenticated, signed, audience-bound, revision-checked,
idempotent and recorded with its full actor chain. Loss of the central plane
does not make central-owned configuration locally writable. Authentication
continues from the last verified local projection when the resource policy
permits it; otherwise the named capability degrades or fails closed.

## Names and provenance boundary

Nomen runtime identifiers, routes, environment variables, database objects,
telemetry, errors, UI and customer documentation use Nomen vocabulary only.
Imported implementation history is progressively absorbed behind Nomen-owned
packages and schemas. Legally required provenance is retained in release legal
materials but is never used as a runtime product identity or API contract.

The repository-wide migration gate is complete only when the allowlist contains
legal provenance files and explicitly versioned optional adapters—no inherited
runtime package, schema, configuration or customer-facing identifier.

## Done when

- schema lint rejects every new non-`nomen_` object;
- every tenant table passes missing-context, forged-context, pooled-connection,
  backup, analytics, replay and cross-tenant mutation tests;
- dedicated and community deployments run the same tenant suite;
- local-to-central and central-to-local migrations pass conflict, outage,
  rollback and disaster-recovery drills;
- disabling local authority removes the write path from UI, API, CLI and AI
  action discovery without hiding the reason; and
- the runtime-name and dependency gates contain no unexplained allowlist entry.
