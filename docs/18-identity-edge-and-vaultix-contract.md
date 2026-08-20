# 18 — Identity edge and Vaultix integration contract

**Status:** accepted build contract; capability claims remain preview until the
proofs below pass.
**Depends on:** `08-control-surface-and-federation.md`,
`14-operation-contract.md` and `16-capability-discovery-contract.md`.

## Product boundary

Tessera is the Shippin identity and authorization platform. Vaultix is the
Infisical-derived Shippin platform for centralized application secrets,
certificates and privileged access. Zuul installs and configures the private
mesh. Shippin supplies the persistent shell, account/workspace context and
commercial policy.

They integrate; they do not absorb one another:

| system | authoritative state in this integration |
|---|---|
| Tessera | subjects, sessions, authentication journeys, federation, access decisions, identity-edge policy and identity audit |
| Vaultix | secret values, certificate private keys and lifecycle, privileged credentials/leases and custody audit |
| Zuul | nodes, peers, routes, mesh policy and installer lifecycle |
| Shippin | tenant/product context, service composition, invitations to the private ecosystem and customer-facing shell |

Tessera may retain an opaque Vaultix resource reference, purpose, safe status,
rotation/expiry time and audit correlation id. It must never persist, project,
log, blueprint-export or place in an event the resolved secret value.

### Safe reference shape

The Tessera-owned read model is `SecretReference`. It contains only:

- a stable Tessera reference id and tenant scope;
- one closed purpose such as `ldap_bind` or `edge_tls`;
- an opaque `vaultix://` provider reference without user-info, query or
  fragment components;
- lifecycle state, safe revision/version, rotation and expiry times;
- a custody audit correlation id.

It deliberately has no generic metadata map, payload, value, credential,
private-key or provider-response field. A plan binds its named secret slot to
the Tessera reference id, never directly to the Vaultix provider reference.

Enrollment accepts protected bytes through a write-only stream and returns
only `SecretReference`. Resolution is callback-scoped: the custody adapter
provides bytes only to the authorized operation callback, clears its working
copy on return and exposes only a safe receipt. A callback error is wrapped in
a fixed typed failure and its text is not propagated because dependency errors
can accidentally contain credentials. The sole exception is a Tessera domain
error that explicitly implements the custody-safe marker; this preserves its
typed refusal without admitting arbitrary provider error text.

References are purpose- and tenant-bound. Cross-account, cross-workspace and
wrong-purpose use fail before resolution. Expired, revoked, denied or
unavailable material is never served from an unbounded cache.

## What “built in” means

Built in means Tessera owns the provider-neutral configuration, policy,
lifecycle, health, audit and guided experience. A capability may run in the
core or in a separately supervised Tessera edge. Separate processes are an
availability and protocol boundary, not a second product or authority.

Source presence, a successful compile, a configured toggle or a running
process does not prove a capability. `available` requires the conformance
evidence described here for the exact installed bundle.

## Required capabilities and proof

| capability id | customer promise | minimum proof before `available` |
|---|---|---|
| `ldap_outbound` | Tessera connects to an external LDAP/Active Directory service for sign-in, lookup and mapping. | StartTLS and LDAPS certificate validation; bind/search; nested group and attribute mapping; disabled/removed user behavior; timeout/failover; tenant isolation; Vaultix-backed bind credential; OpenLDAP plus the supported AD profile. |
| `ldap_inbound` | A legacy application binds to a Tessera-managed LDAP edge and sees its allowed directory view. | standard client bind/search/group lookup; scoped base DN and schema mapping; cross-tenant negative tests; revocation propagation; TLS-only production mode; restart/reconnect; no password-equivalent stored at the edge. |
| `forward_auth` | An existing reverse proxy delegates each request decision to Tessera. | allow/deny/step-up paths; original method/URL preservation; spoofed identity-header stripping; minimal signed or mutually authenticated result headers; websocket/stream behavior where supported; session revocation; fail-closed dependency loss. |
| `identity_aware_proxy` | A Tessera edge owns the browser sign-in session and proxies an application that cannot integrate itself. | redirect and callback integrity; cookie scope and rotation; upstream host allowlist; CSRF and open-redirect negatives; websocket/stream behavior; logout/revocation; tenant isolation; fail-closed dependency loss. |
| `visual_flow_engine` | An operator builds and maintains sign-in/recovery journeys visually without creating a second runtime model. | graph ↔ canonical schema round trip; typed stage validation; unreachable/cycle/unsafe-terminal rejection; branch simulation; revision diff; idempotent publish; rollback; audit; keyboard and screen-reader operation. |
| `vaultix_secret_custody` | Tessera uses Vaultix for connector credentials, certificate private keys and privileged access without exposing values. | workload authentication; least-privilege path policy; write-only enrollment; resolve/use without persistence; rotation overlap; revoke/expiry behavior; unavailable/denied degradation; audit correlation; seeded-secret scans across DB, events, API, logs and support bundles. |

Each proof result contains a stable conformance id, installed bundle digest,
test-environment profile, verification time, pass/fail result and immutable
evidence digest. The browser receives the identifiers and safe summary, not
raw test artifacts or protected configuration.

## LDAP direction and lifecycle

Direction is always stated from Tessera's point of view:

- **outbound:** Tessera initiates the connection to a customer directory;
- **inbound:** a customer application initiates the connection to a
  Tessera-managed LDAP edge.

Outbound LDAP authentication and directory provisioning are separate effects.
A connector explicitly declares `authenticate`, `import`, `reconcile` and/or
`deprovision`; a successful bind must never imply that removed users will be
suspended automatically. Mapping is previewed against non-persisted sample
results before enablement, and an emergency local-owner path remains available.

Inbound LDAP publishes a tenant-scoped virtual directory from Tessera's
identity authority. Edge configuration is versioned and narrowly scoped. The
edge never becomes a writable shadow directory, password database or source of
entitlement decisions.

## Forward-auth and proxy modes

The guide asks what the application can do, then recommends one of three
distinct integrations:

1. native OIDC/SAML when the application supports it;
2. forward-auth when an existing reverse proxy can enforce Tessera's decision;
3. identity-aware proxy when Tessera must own the browser session and upstream
   request.

Review shows the trust boundary, protected hosts/routes, headers released,
session behavior, outage behavior and rollback. Incoming identity headers are
removed before Tessera adds its own. Identity headers contain the minimum
claims required by policy and are integrity protected between the enforcement
point and application.

## Visual flow engine

The visual editor, YAML blueprints, management API and runtime executor share
one canonical versioned graph. The UI is not allowed to invent nodes or defaults
the runtime cannot represent.

The default experience starts from reviewed templates for passwordless,
password + MFA, company sign-in, recovery and step-up. Progressive disclosure
shows one decision at a time; experts can open the full graph and exact schema.
Before publish, Tessera must:

1. validate stage inputs, reachability, terminal outcomes and retry/lockout
   behavior;
2. simulate success, refusal, cancellation and dependency-failure branches;
3. show the exact resource/policy diff and affected tenants/apps;
4. name rollback and emergency-owner effects;
5. require the permissions and step-up level of every changed resource;
6. publish idempotently against the reviewed base revision and record audit.

## Vaultix integration

Tessera authenticates to Vaultix using a workload identity and short-lived,
purpose-scoped authorization. A reusable root or operator token is forbidden.
The integration adapter supports:

- create or bind a write-only secret slot for a connector;
- resolve a value only inside the process performing the authorized operation;
- request/status/renew/revoke a certificate without taking custody of its
  private key outside the approved runtime boundary;
- request and release a time-bounded privileged credential when an operation
  requires one;
- consume Vaultix rotation/expiry/revocation events and project only safe
  status;
- correlate Tessera operation/audit ids with Vaultix custody audit ids.

Vaultix denial or outage degrades only the capability that needs the protected
material. Cached secret values are not silently extended past policy. Core
local-owner recovery must have an independently documented custody path so a
single integration failure cannot lock out the platform owner.

## Guided Shippin experience

The Shippin shell presents customer outcomes:

- **Connect my company directory** → OIDC/SAML or outbound LDAP;
- **Protect an existing app** → native federation, forward-auth or proxy;
- **Connect a legacy directory app** → inbound LDAP;
- **Customize sign-in or recovery** → template-first visual flow engine;
- **Set up protected credentials** → Vaultix-backed secret/certificate binding.

Each uses `plan → review → apply → verify`. The review names tenant, trust
direction, data released, Vaultix custody effects, outage behavior and
rollback. Protocol vocabulary is progressive detail, never a prerequisite.

## Release gate

Tessera cannot call itself operational for the private cloud until all six
capabilities above either pass their exact conformance suite or are visibly
marked preview/unsupported. The dashboard may not infer support from inherited
routes or binaries, and one passing edge cannot promote the others.
