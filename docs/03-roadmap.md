# 03 — Standalone product roadmap and worklist

**Status:** active. The release boundary is
`02-standalone-product-contract.md`.

## Ordering rule

Work proceeds in this order:

1. Tessera standalone product.
2. Tessera managed-customer operation.
3. Optional Vaultix and Zuul integrations.
4. Shippin adapter and embedded shell.

An integration cannot close a standalone milestone. A Shippin screenshot is not
evidence that Tessera is deployable; the same outcome must first work in
Tessera's own UI and API with Shippin absent.

## Evidence labels

- **Done:** implemented and its stated proof passes on this repository branch.
- **Active:** implementation exists but the milestone's full real-target or
  operational proof is incomplete.
- **Planned:** contract or design only.
- **Blocked:** an external estate is required and the exact evidence is named.

Capabilities remain `preview` unless their promotion row is Done.

## Foundation already proved

| Work | Status | Evidence |
|---|---|---|
| Tessera module builds from this repository | Done | generated source and Go build path |
| Asymmetric Shippin seat-token profile | Done | Tessera mint plus Automaton verifier negative tests |
| Tessera seat storage and CLI | Done | repository and mint-path tests |
| Declarative blueprints | Done | validation, idempotent apply and PostgreSQL atomicity proof |
| Configured authentication flows | Done | password, MFA and recovery configurations use one executor; negative input fails closed |

These proofs are valuable but do not satisfy the standalone release gate.

## P0 — Correct the product boundary

| ID | Work | Status | Acceptance |
|---|---|---|---|
| P0.1 | Make Tessera standalone product law authoritative | Done | charter, agent rules and README forbid a Shippin runtime dependency |
| P0.2 | Preserve the seat token as an optional integration profile | Done | Automaton contract remains versioned and unchanged on the wire |
| P0.3 | Inventory host-product assumptions in code, config, docs and UI | Planned | every assumption is removed, made generic or isolated in an adapter package |
| P0.4 | Establish product-language and dependency-boundary tests | Active | primary product-document regression test passes; UI/config and adapter-import checks remain |
| P0.5 | Absorb inherited runtime names and schemas | Planned | runtime/API/config/database/telemetry scans pass; only legal provenance and versioned optional adapters remain allowlisted |

## P1 — Standalone deployment

| ID | Work | Depends on | Acceptance |
|---|---|---|---|
| P1.1 | Produce minimal signed OCI image | P0.3 | Active — local OCI image runs non-root without a shell and exposes version/provenance; signing remains |
| P1.2 | Define `TESSERA_*` configuration contract | P0.3 | Done — runtime and Podman examples use the Tessera namespace |
| P1.3 | Implement read-only preflight | P1.2 | runtime, PostgreSQL, DNS/TLS, ports, storage and notification failures return typed remediation |
| P1.4 | Implement idempotent initialize/start lifecycle | P1.3 | interrupted initialization resumes safely and repeated start changes nothing |
| P1.5 | Ship rootless Podman artifacts | P1.1–P1.4 | Active — Quadlet generation and hardened image pass; clean-host database/bootstrap drill remains |
| P1.6 | Ship orchestrator profile | P1.5 | the same image and conformance suite pass without a second behavior path |
| P1.7 | Ship least-privilege Tessera operator | P1.3–P1.5 | it controls only named Tessera Podman resources and performs resumable install/backup/restore/upgrade operations |

## P2 — Canonical Tessera experience

| ID | Work | Depends on | Acceptance |
|---|---|---|---|
| P2.1 | Define versioned management resources and typed errors | P0 | Active — overview, capabilities, actions and events share typed authorization/errors; mutation resources remain |
| P2.2 | Build standalone shell and navigation | P2.1 | Done — direct Tessera URL exposes Start, Directory, Applications, Federation, Access, Flows, Security, Audit and Settings |
| P2.3 | Build guided first-owner enrollment | P1.4, P2.2 | owner enrolls passkey/MFA and exports independent recovery; resume is tested |
| P2.4 | Build guided organization and application setup | P2.1–P2.3 | a new operator completes first OIDC application without protocol jargon |
| P2.5 | Build end-user account and security surface | P2.1 | profile, factors, sessions, consent and recovery are usable without admin access |
| P2.6 | Add capability discovery | P2.1 | Active — runtime and UI are fail-closed; signed production evidence publication remains |
| P2.7 | Export the canonical UI module | P2.2 | Active — runtime embeds `@tessera/ui`; registry publication and host-composition proof remain |
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
| P4.10 | Enforce prefixed storage and PostgreSQL RLS | P2.8 | every new object uses `tessera_`; forged/missing context and pooled-connection tenant suites fail closed |
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
| P6.3 | Shippin server-side adapter | P5 | adapter maps account/commercial context through versioned Tessera APIs only |
| P6.4 | Shippin embedded Tessera module | P2.2, P6.3 | clicking Tessera mounts the same product module and capabilities without forked identity logic |
| P6.5 | Shippin seat-token profile promotion | P6.3 | existing Automaton verification remains green and profile disablement leaves standalone behavior unchanged |

## Next execution tranche

The next work is P0.3–P1.7 and P2.1–P2.2: inventory and isolate host
assumptions, lock the configuration namespace, build the standalone image,
implement preflight, expose truthful management discovery and load Tessera's
own product shell from a rootless Podman installation without Shippin. In
parallel only where it does not change that dependency direction, continue
P3.5's Microsoft AD lab profile and the production lifecycle adapter required
by P3.9.

The first managed customer is not invited until P4.7 passes. Shippin UI work
starts at P6.3, after Tessera has already proven itself as the product being
integrated.
