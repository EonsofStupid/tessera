# 03 — Standalone product roadmap and worklist

**Status:** active. The release boundary is
`02-standalone-product-contract.md`.

The complete Nomen capability ledger, external certification tracks,
evidence manifests, collaboration checkpoints and legal/release prerequisites
are governed by `22-certification-and-parity-program.md`. This roadmap provides
the existing product milestones; the certification program supplies the
complete capability ledger and promotion sequence.
Current non-secret product-owner/distribution decisions and unresolved release
inputs are recorded in `23-release-governance-record.md`; existing branch and
worktree preservation is governed by `24-work-preservation-inventory.md`.

## Ordering rule

Work proceeds in this order:

1. Nomen standalone product.
2. Nomen provider-neutral managed-customer operation using the same
   standalone product build and contracts.
3. Shippin-managed adapter and embedded shell consuming those contracts.
4. Optional Vaultix and Zuul integrations.

An integration cannot close a standalone milestone. A Shippin screenshot is not
evidence that Nomen is deployable; the same outcome must first work in
Nomen's own UI and API with Shippin absent.

The product may present a public commercial entry page while remaining
invite-only behind AngryVibes LLC's private panel. Public presentation does not
create public self-service enrollment or expose authenticated management facts.

## Evidence labels

- **Done:** implemented and its stated proof passes on this repository branch.
- **Active:** implementation exists but the milestone's full real-target or
  operational proof is incomplete.
- **Planned:** contract or design only.
- **Blocked:** an external estate is required and the exact evidence is named.

Capabilities remain `preview` unless their promotion row is Done.

The unit of delivery is the capability defined by
`20-capability-delivery-standard.md`. A page, container, inherited package,
database migration or passing unit test cannot be marked as an operational IAM
milestone on its own.

## Foundation inventory already available

| Work | Classification | Evidence |
|---|---|---|
| Inherited IAM core builds from this repository | Inventory | generated source and Go build path; Nomen promotion evidence incomplete |
| Asymmetric Shippin seat-token profile | Verified optional profile | Nomen mint plus Automaton verifier negative tests |
| Nomen seat storage and CLI | Verified optional profile | repository and mint-path tests |
| Declarative blueprints | Inventory | validation, idempotent apply and PostgreSQL atomicity proof; standalone operator journey incomplete |
| Authentication-flow implementation | Inventory | password, MFA and recovery executor tests exist; Nomen browser, recovery and release evidence incomplete |

These assets reduce implementation time but do not establish an operational
Nomen capability. They enter the intake and promotion process in
`20-capability-delivery-standard.md`.

## Measured runtime baseline — 2026-08-21

The first clean-database Podman drill produced the following facts. These are
starting measurements, not release evidence:

| Measurement | Observed | Required direction |
|---|---:|---|
| PostgreSQL application tables | 186 | inventory ownership and migration behavior for every table used by an operational capability |
| Tables with RLS enabled and forced | 2 | every tenant-owned Nomen table and query path fails closed under forged, missing and pooled tenant context |
| Nomen-prefixed tables | 2 | all new product-owned objects use the prefix; inherited objects receive an explicit absorption migration |
| Inherited application schemas | 9 | remove runtime product identity and converge on a documented Nomen storage boundary without parallel identity truth |
| Rootless application/database containers | 2 healthy | preserve clean initialization, failure resume and steady-state restart in the automated harness |
| Production-image Playwright tests | 2 passing | expand to every project in `21-first-iam-vertical-slice.md`; baseline smoke tests cannot promote IAM |
| Browser runtime validator | ArkType 2.2.3 | cover every environment, token, management and command response entering browser state |
| Browser end-to-end runner | Playwright 1.62.1 | execute against the release image and disposable PostgreSQL, never the Vite preview |

The first initialization intentionally failed on a generated password that did
not satisfy the configured complexity policy, then resumed successfully with a
policy-compliant runtime secret against the partially initialized database.
The harness must preserve this interrupted-initialization case. Steady-state
restart against the initialized volume and the production-image browser smoke
suite both pass.

## P0 — Correct the product boundary

| ID | Work | Status | Acceptance |
|---|---|---|---|
| P0.1 | Make Nomen standalone product law authoritative | Done | charter, agent rules and README forbid a Shippin runtime dependency |
| P0.2 | Preserve the seat token as an optional integration profile | Done | Automaton contract remains versioned and unchanged on the wire |
| P0.3 | Inventory host-product assumptions in code, config, docs and UI | Planned | every assumption is removed, made generic or isolated in an adapter package |
| P0.4 | Establish product-language and dependency-boundary tests | Active | primary product-document regression test passes; UI/config and adapter-import checks remain |
| P0.5 | Absorb inherited runtime names and schemas | Planned | runtime/API/config/database/telemetry scans pass; only legal provenance and versioned optional adapters remain allowlisted |

## P1 — Standalone deployment

| ID | Work | Depends on | Acceptance |
|---|---|---|---|
| P1.1 | Produce minimal signed OCI image | P0.3 | Active — local OCI image runs non-root without a shell and exposes version/provenance; signing remains |
| P1.2 | Define `NOMEN_*` configuration contract | P0.3 | Done — runtime and Podman examples use the Nomen namespace |
| P1.3 | Implement read-only preflight | P1.2 | Active — protected v1 API and guided UI evaluate PostgreSQL, WebAuthn-valid issuer, TLS, asymmetric signing and notification delivery with typed remediation; DNS/port/storage capacity probes and authorized production evidence remain |
| P1.4 | Implement idempotent initialize/start lifecycle | P1.3 | interrupted initialization resumes safely and repeated start changes nothing |
| P1.5 | Ship rootless Podman artifacts | P1.1–P1.4 | Active — Quadlet generation and hardened image pass; clean-host database/bootstrap drill remains |
| P1.6 | Ship orchestrator profile | P1.5 | the same image and conformance suite pass without a second behavior path |
| P1.7 | Ship least-privilege Nomen operator | P1.3–P1.5 | it controls only named Nomen Podman resources and performs resumable install/backup/restore/upgrade operations |

## P2 — Canonical Nomen experience

| ID | Work | Depends on | Acceptance |
|---|---|---|---|
| P2.1 | Define versioned management resources and typed errors | P0 | Active — overview, capabilities, actions and events share typed authorization/errors; mutation resources remain |
| P2.2 | Build standalone shell and navigation | P2.1 | Active — direct Nomen routes render, but disabled placeholders and source-text tests are not operational UI evidence; production Playwright coverage remains |
| P2.3 | Build guided first-owner enrollment | P1.4, P2.2 | Active — constant-time runtime bootstrap authority, deterministic resumable challenge, real WebAuthn verification, forced-RLS public-credential persistence, one-time recovery export/confirmation, restart persistence and Chromium virtual-authenticator evidence pass; linking the enrolled owner into the canonical user/role/session model and disposable-restore evidence remain before promotion |
| P2.4 | Build guided organization and application setup | P2.1–P2.3 | a new operator completes first OIDC application without protocol jargon |
| P2.5 | Build end-user account and security surface | P2.1 | profile, factors, sessions, consent and recovery are usable without admin access |
| P2.6 | Add capability discovery | P2.1 | Active — runtime and UI are fail-closed; signed production evidence publication remains |
| P2.7 | Export the canonical UI module | P2.2 | Active — runtime embeds `@nomen/ui`; registry publication and host-composition proof remain |
| P2.8 | Prove dedicated and community tenancy | P2.1–P2.6 | one dedicated customer and two community tenants pass the same identity journeys without cross-tenant leakage |
| P2.9 | Add semantic operator events | P2.1–P2.2 | Active — current shell controls emit allowlisted events into RLS event/outbox storage; live tenant and redaction conformance remains |
| P2.10 | Publish human/AI action discovery | P1.7, P2.1 | Active — fail-closed catalog publishes schemas and consequences; executable plan/apply/verify actions remain |
| P2.11 | Build JSON-guided teaching components | P2.2, P2.6 | versioned cards, terminology, safe seed suggestions and reduced-motion visuals reject scripts and false promotion |

## P3 — IAM completeness and conformance

| ID | Work | Depends on | Acceptance |
|---|---|---|---|
| P3.1 | OIDC/OAuth application lifecycle | P2.1 | discovery, code flow with PKCE, refresh, logout, revocation, rotation and negative suites pass |
| P3.2 | SAML service-provider lifecycle | P2.1 | metadata, signing/encryption, login/logout and rollover suites pass |
| P3.3 | Passkeys, TOTP, recovery and session controls | P2.3 | enrollment, loss, replay, lockout, revocation and recovery ceremonies pass |
| P3.4 | External IdP federation | P2.4 | OIDC and SAML upstream profiles pass linking, collision and unavailable-provider tests |
| P3.5 | Outbound LDAP | P2.6 | OpenLDAP suite passes now; Microsoft AD profile remains preview until STIG 2025 evidence passes |
| P3.6 | Inbound LDAP | P2.6 | bind, search, groups, paging, TLS, tenant isolation and abuse suites pass against real clients |
| P3.7 | Forward-auth and identity-aware proxy | P2.6 | real upstream applications pass header, cookie, websocket, logout and bypass-resistance suites |
| P3.8 | Visual flow editor and execution | P2.2 | graph validation, simulation, publish, revision, rollback and runtime equivalence pass |
| P3.9 | Provisioning and directory lifecycle | P2.1 | create/update/suspend/reactivate, idempotency, rollback and stale-revision tests pass against production adapters |

P3 work is executed as vertical capability slices. Each row must satisfy every
layer in `20-capability-delivery-standard.md`; backend-only and UI-only work do
not travel as separate claims of completion.

## P4 — Commercial-grade operations

| ID | Work | Depends on | Acceptance |
|---|---|---|---|
| P4.1 | Health, readiness, metrics and structured logs | P1 | failures identify the owning dependency without leaking sensitive data |
| P4.2 | Immutable identity audit and export | P2.1 | authentication, administration, delegation, recovery and managed access are queryable and exportable |
| P4.3 | Backup and restore | P1, P4.2 | destructive restore drill proves issuer, keys, identity, policy, audit and recovery continuity |
| P4.4 | Upgrade, migration and rollback | P1 | supported-version matrix passes forward migration and documented rollback boundary |
| P4.5 | Signing and encryption key rotation | P3.1–P3.2 | overlap, cache, failure and recovery tests pass without accepting retired keys |
| P4.6 | High availability and failure recovery | P4.1–P4.5 | node loss, PostgreSQL interruption and dependency degradation preserve stated guarantees |
| P4.7 | Security release gate | all P1–P4 | threat model, dependency scan, provenance, secret scan, tenant-isolation and protocol suites pass |
| P4.8 | Add ClickHouse OLAP projection | P4.2 | transactional outbox, replay, deduplication, tenant-scoped aggregate API and dependency-loss tests pass without placing analytics on the auth path |
| P4.9 | Add opt-in privacy-safe session replay | P2.9, P4.2 | protected surfaces are uncapturable; tenant access, encryption, retention and deletion negative suites pass |
| P4.10 | Enforce prefixed storage and PostgreSQL RLS | P2.8 | every new object uses `nomen_`; forged/missing context and pooled-connection tenant suites fail closed |
| P4.11 | Add local/central/delegated authority | P2.10, P4.10 | one writer per resource family; signed revisions, outage, conflict, rollback and recovery drills pass |

## P5 — Managed customers

| ID | Work | Depends on | Acceptance |
|---|---|---|---|
| P5.1 | Define managed deployment enrollment | P4.7 | customer deployment grants explicit scoped management without surrendering owner recovery |
| P5.2 | Define managed operation contract | P5.1 | preflight, install, upgrade, rotate, backup and restore are resumable operations with typed status |
| P5.3 | Add delegated operator roles | P5.1 | least privilege, approval, expiry and revocation tests pass |
| P5.4 | Add managed fleet overview | P5.2 | operator sees versions, readiness, capability degradation and due maintenance without tenant data leakage |
| P5.5 | Prove customer offboarding and export | P5.2–P5.4 | customer can revoke managed access and leave with documented data/config export |

## P6 — Optional platform integrations

| ID | Work | Depends on | Acceptance |
|---|---|---|---|
| P6.1 | Vaultix runtime secret adapter | P4.7 | workload authentication resolves value-blind references; unavailable custody fails closed; managed production requires the adapter while standalone bootstrap does not |
| P6.2 | Zuul enrollment and private access | P4.7 | target enrolls without reusable embedded credentials; mesh loss does not corrupt identity state |
| P6.3 | Shippin server-side adapter | P5 | adapter maps account/commercial context through versioned Nomen APIs only |
| P6.4 | Shippin embedded Nomen module | P2.2, P6.3 | clicking Nomen mounts the same product module and capabilities without forked identity logic |
| P6.5 | Shippin seat-token profile promotion | P6.3 | existing Automaton verification remains green and profile disablement leaves standalone behavior unchanged |

## Current execution tranche — first real IAM journey

The only active product tranche is `21-first-iam-vertical-slice.md`. Work is
performed in this dependency order:

1. inventory the clean PostgreSQL schema, running routes, existing protocol
   implementation and inherited product identifiers from the production image;
2. establish a repeatable real-container test harness with PostgreSQL and no
   Shippin dependency;
3. remove unused TanStack Start, adopt ArkType for browser trust-boundary
   schemas, and establish Playwright against the production build;
4. absorb and rename the persistence and service paths required by the slice
   into Nomen-owned boundaries without creating a second identity core;
5. implement bootstrap owner, organization, application, invitation, OIDC PKCE,
   entitlement, session revocation and audit as one server-owned journey;
6. expose the journey through the guided Nomen UI and equivalent Clyffy
   plan/apply/verify actions; and
7. pass protocol negatives, tenant isolation, browser accessibility, restart,
   backup/restore, secret scanning and Automaton consumer verification against
   one release image digest.

No additional navigation-only dashboard work is part of this tranche. LDAP,
SAML federation, proxy access and visual flows remain queued behind this slice
so they inherit one proven implementation, UI, operator and evidence pattern.

The first managed customer is not invited until P4.7 passes. Shippin UI work
starts at P6.3, after Nomen has already proven itself as the product being
integrated.
