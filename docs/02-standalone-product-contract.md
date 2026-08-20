# 02 — Standalone product contract

**Status:** accepted product boundary; implementation remains pre-release.

## Promise

A customer can deploy Tessera, complete guided owner setup, manage identity and
access, integrate applications, prove sign-in and recovery, observe the system,
upgrade it and restore it without installing or contacting Shippin.

This promise is the release boundary. A protocol, adapter or workflow remains
`preview` and fail-closed until its conformance suite passes against a real
target. A visible control or successful unit test does not make a capability
operational.

## Independence invariants

A standalone or managed-customer deployment:

- uses Tessera-owned product names, routes, configuration and documentation;
- exposes Tessera's complete web application, management API and CLI;
- starts and reaches readiness without a Shippin account, token, adapter, route
  or service discovery record;
- does not require Zuul for networking or Vaultix for initial bootstrap;
- accepts optional secret-manager and mesh adapters through versioned Tessera
  interfaces;
- keeps every tenant, signing key, session and audit record inside the selected
  deployment boundary; and
- can export its configuration and identity data using documented, tested
  procedures.

An integration may add context or navigation. It may not become the only way to
configure, recover or operate Tessera.

## Required standalone surfaces

| Surface | Required outcome |
|---|---|
| Guided setup | preflight, database initialization, public issuer, TLS boundary, first owner, MFA, independent recovery and readiness proof |
| Administration | organizations, users, groups, service identities, applications, roles, policies, providers, sessions and audit |
| End-user account | profile, authenticators, active sessions, consent, recovery and security events |
| Protocols | OIDC/OAuth 2.x, SAML and each promoted federation or directory profile with published metadata and conformance evidence |
| Access edges | outbound/inbound LDAP, forward-auth and proxy capabilities promoted independently and failed closed when unavailable |
| Automation | stable management API, CLI and declarative blueprints with idempotency and typed errors |
| Operations | health, readiness, metrics, structured logs, audit export, backup, restore, upgrade, rollback and key rotation |
| Deployment | one signed OCI image tested rootless with Podman and through the supported orchestrator profile |

The standalone web application is not a temporary console. It is the canonical
guided product experience and must consume the same public management contracts
available to managed operators and adapters.

## Deployment modes

### Standalone self-hosted

The customer owns the runtime, PostgreSQL, public domain and custody choices.
Tessera guides setup and reports exact prerequisites without assuming a control
plane above it.

### Managed customer

The operator provisions isolated Tessera deployments and performs approved
maintenance through Tessera's management and operation contracts. Customer
owners retain their own roles, recovery path, audit visibility and export path.
Managed access is explicit, time-bounded and auditable; it is never silent
impersonation.

### Embedded host integration

After standalone promotion, a host product may mount the Tessera UI module and
map its own navigation or account context through an adapter. The adapter calls
versioned Tessera APIs and capability discovery. Direct browser dependence on a
host-only operator credential is forbidden.

Shippin's seat token is one such integration profile. Disabling it must not
remove any standalone IAM function.

## Dependency ownership

| Component | Requirement | Owner |
|---|---|---|
| PostgreSQL | required durable store | deployment operator |
| public DNS and TLS | required outside development | deployment operator |
| SMTP or another notification transport | required before recovery is promoted | deployment operator through a Tessera adapter |
| Vaultix | optional recommended secret custody | Vaultix; Tessera stores references |
| Zuul | optional mesh and private-access integration | Zuul; Tessera owns identity and enrollment policy |
| Shippin | optional commercial shell and ecosystem integration | Shippin adapter |

No optional component may change the meaning of Tessera's core identity data or
be required to recover a standalone owner.

## Security invariants

- Asymmetric signatures only; consumers never verify a Tessera token with a
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

Tessera is ready for the first managed customer only when all of these are
reproducible from a clean host:

1. Build and verify a signed OCI image and source provenance.
2. Install rootless with generated runtime-only secrets and no Shippin service.
3. Complete guided owner enrollment with phishing-resistant MFA and independent
   recovery.
4. Create an organization, users, groups, application and service identity from
   Tessera's UI, and repeat the same outcomes through the API.
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

The Shippin adapter and embedded shell are deliberately outside this gate. They
begin only after the same build is operable as Tessera by itself.
