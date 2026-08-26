# 02 — Standalone product contract

**Status:** accepted product boundary. Product version is `1.0.0-alpha` and
frozen until Nomen Vault, Nomen Mesh, and the first IAM vertical slice are
built. Implementation of those capabilities remains preview.

## Promise

A customer can deploy Nomen, complete guided owner setup, manage identity and
access, integrate applications, prove sign-in and recovery, observe the system,
upgrade it and restore it without installing or contacting Shippin.

This promise is the release boundary. A protocol, adapter or workflow remains
`preview` and fail-closed until its conformance suite passes against a real
target. A visible control or successful unit test does not make a capability
operational.

## Independence invariants

A standalone or managed-customer deployment:

- uses Nomen-owned product names, routes, configuration and documentation;
- exposes Nomen's complete web application, management API and CLI;
- starts and reaches readiness without a Shippin account, token, adapter, route
  or service discovery record;
- does not require Nomen Mesh or Nomen Vault for initial bootstrap;
- accepts optional secret-manager and mesh adapters through versioned Nomen
  interfaces as **enterprise** features (`26-editions-and-demo-tier.md`);
- keeps every tenant, signing key, session and audit record inside the selected
  deployment boundary; and
- can export its configuration and identity data using documented, tested
  procedures.

An integration may add context or navigation. It may not become the only way to
configure, recover or operate Nomen.

## Required standalone surfaces

| Surface | Required outcome |
|---|---|
| Guided setup | preflight, database initialization, public issuer, TLS boundary, first owner, MFA, independent recovery and readiness proof |
| Administration | organizations, users, groups, service identities, applications, roles, policies, providers, sessions and audit |
| End-user account | profile, authenticators, active sessions, consent, recovery and security events |
| Protocols | OIDC/OAuth 2.x, SAML and each promoted federation or directory profile with published metadata and conformance evidence |
| Access edges | outbound/inbound LDAP, forward-auth and proxy capabilities promoted independently and failed closed when unavailable |
| Automation | stable management API, CLI and declarative blueprints with idempotency and typed errors |
| Operator intelligence | typed human/AI action discovery, semantic runtime events, JSON-guided teaching and privacy-safe optional replay |
| Operations | health, readiness, metrics, structured logs, audit export, backup, restore, upgrade, rollback and key rotation |
| Deployment | one signed OCI image tested rootless with Podman and through the supported orchestrator profile |

The standalone web application is not a temporary console. It is the canonical
guided product experience and must consume the same public management contracts
available to managed operators and adapters.

The browser experience has two deliberate layers at Nomen's own origin:

- `/` leads to the public product entry surface. It explains what this
  deployment is, which capabilities are live, where authority is stored and
  how an operator enters the product without requiring a Shippin shell or an
  authenticated session.
- `/ui/console/overview` is the operator workspace. Its management facts remain
  authentication-gated and server-authoritative; the public entry surface must
  never leak tenant counts, identities, policy, audit evidence or deployment
  secrets.

The public surface, operator workspace, sign-in callback and protocol endpoints
ship from the same Nomen artifact. A release must exercise the public entry,
operator navigation, sign-in handoff and a responsive viewport in a real
browser. Rendering a static mock or validating components without the
production HTTP boundary does not satisfy this requirement.

It is a Nomen-owned React and TypeScript application served by the Nomen
runtime. Browser trust-boundary payloads are parsed with versioned ArkType
schemas before entering application state; TypeScript assertions alone never
validate runtime data. The Go service remains authoritative and independently
validates every command. A routing or rendering framework is an implementation
choice, not part of the product contract, and unused full-stack frameworks are
forbidden from the runtime dependency graph.

The same versioned feature module may later be composed into a host shell, but
its routes, resources and behavior remain usable at Nomen's own origin.

Every meaningful interaction uses the stable ids, semantic events and shared
human/AI action grammar in `18-operator-interaction-contract.md`. Presentation
may become more visual or animated without turning screen coordinates, display
text or replay into an administrative API.

## Deployment modes

### Standalone self-hosted

The customer owns the runtime, PostgreSQL, public domain and custody choices.
Nomen guides setup and reports exact prerequisites without assuming a control
plane above it.

### Managed customer

The operator provisions isolated Nomen deployments and performs approved
maintenance through Nomen's management and operation contracts. Customer
owners retain their own roles, recovery path, audit visibility and export path.
Managed access is explicit, time-bounded and auditable; it is never silent
impersonation.

The first managed release supports two explicit isolation profiles:

- **dedicated** — one Nomen runtime and logical PostgreSQL database for one
  customer; databases may share a hardened PostgreSQL cluster; and
- **community** — one Nomen deployment with multiple organization tenants,
  invite-only onboarding and tenant-scoped authorization, caches, audit and
  analytics.

Both profiles run the same product build and IAM conformance suite. Community
tenant owners cannot perform deployment-global operations.

Product-owned storage, row-level security and local/central authority follow
`19-tenancy-and-authority-contract.md`. Central management disables conflicting
local writes explicitly; two control planes never become simultaneous sources
of truth.

### Embedded host integration

After standalone promotion, a host product may mount the Nomen UI module and
map its own navigation or account context through an adapter. The adapter calls
versioned Nomen APIs and capability discovery. Direct browser dependence on a
host-only operator credential is forbidden.

Shippin's seat token is one such integration profile. Disabling it must not
remove any standalone IAM function.

## Dependency ownership

| Component | Requirement | Owner |
|---|---|---|
| PostgreSQL | required durable store | deployment operator |
| ClickHouse | required OLAP projection for managed production; never identity truth | deployment operator through Nomen's analytics relay |
| public DNS and TLS | required outside development | deployment operator |
| SMTP or another notification transport | required before recovery is promoted | deployment operator through a Nomen adapter |
| Vaultix | required for managed production and optional for self-hosting | Vaultix; Nomen stores references |
| Zuul | optional mesh and private-access integration | Zuul; Nomen owns identity and enrollment policy |
| Shippin | optional commercial shell and ecosystem integration | Shippin adapter |

No optional component may change the meaning of Nomen's core identity data or
be required to recover a standalone owner.

Nomen also ships a least-privilege deployment operator for its own rootless
Podman resources. It owns preflight, initialization, backup, restore, upgrade
and rollback for Nomen only; it is not a general infrastructure inventory or
mesh controller.

## Transactional and analytical data

PostgreSQL is the only authority for identity, policy, configuration,
operations and audit events. A transactional outbox projects redacted,
tenant-scoped analytical facts into ClickHouse with event-id deduplication and
resumable checkpoints. ClickHouse is rebuildable and never sits on an
authentication or authorization request path. Its outage degrades analytics,
not identity truth.

The analytics API is Nomen-owned. Browsers and host integrations never query
ClickHouse directly. Tokens, assertions, credentials, recovery material and
secret values are forbidden from analytical facts.

## Security invariants

- Asymmetric signatures only; consumers never verify a Nomen token with a
  shared secret.
- Authentication failure and authorization failure remain distinct. Missing
  entitlement is `403` with a typed body.
- Bootstrap credentials are single-use, expire, and never appear in images,
  source control, logs or process arguments.
- Recovery requires an independently held factor or documented break-glass
  ceremony and creates immutable audit evidence.
- Management mutations are authorized server-side, idempotent where retry is
  expected, and return typed remediation.
- Cross-tenant access, downgrade, rollback, restore and signing-key rotation
  each have destructive and negative-path tests.
- Secrets used by protocol adapters are resolved only at execution time and
  are never returned through the management API.

## Capability states

Every discoverable capability reports one of:

- `unsupported` — absent from this build;
- `preview` — visible for evaluation, fail-closed for production use;
- `operational` — its required conformance and failure suites pass for the
  declared deployment profile; or
- `degraded` — previously operational, but a named dependency or health proof
  is currently failing.

The UI renders server facts and never infers support from version strings or
route presence.

## Standalone release gate

Nomen is ready for the first managed customer only when all of these are
reproducible from a clean host:

1. Build and verify a signed OCI image and source provenance.
2. Install rootless with generated runtime-only secrets and no Shippin service.
3. Complete guided owner enrollment with phishing-resistant MFA and independent
   recovery.
4. Create an organization, users, groups, application and service identity from
   Nomen's UI, and repeat the same outcomes through the API.
5. Complete OIDC and SAML sign-in/logout, session revocation and key rotation.
6. Exercise every capability labelled `operational` against a real disposable
   target, including negative and unavailable-dependency cases.
7. Back up, destroy the disposable runtime, restore it and prove issuer,
   identity, policy, audit and recovery continuity.
8. Upgrade from the previous supported release, verify migrations, and execute
   the documented rollback boundary.
9. Demonstrate tenant isolation, least-privilege delegated administration,
   audit export and managed-access expiration.
10. Pass Linux and Windows client conformance where a capability claims both.
11. Pass the dedicated and community tenancy suites with identical IAM
    behavior and no cross-tenant read, mutation, cache, invite or analytics
    access.
12. Rebuild ClickHouse projections from PostgreSQL audit/outbox truth and prove
    that analytics loss never blocks authentication.
13. Drive the same reviewed mutation through the browser and AI action surface,
    then prove equivalent authorization, operation and audit evidence.
14. Prove semantic-event and optional replay secret exclusion, tenant isolation,
    retention deletion and central-authority behavior.

The first managed customer also waits for outbound and inbound LDAP,
forward-auth, the identity-aware proxy, visual-flow execution, Vaultix custody,
backup/restore and upgrade conformance. Those capabilities may be reviewed in
preview during development but cannot be promoted or used to waive this gate.

The Shippin adapter and embedded shell are deliberately outside this gate. They
begin only after the same build is operable as Nomen by itself.
