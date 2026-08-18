# 11 — Operational blueprint

**Status:** proposed build contract; execution is tracked in
`12-execution-worklist.md`.
**Depends on:** `01-seat-token-contract.md`,
`08-control-surface-and-federation.md` and
`09-guided-identity-experience.md`.

## The operating promise

An owner does the consequential work once: chooses the domain and deployment,
establishes recovery, connects trusted identity providers, defines the first
community and delegates bounded roles. After that, the team runs the private
community from Tessera's guided Shippin surface without routine shell, YAML,
SQL, Podman or identity-protocol work.

The golden path is one sentence:

> Click Tessera in Shippin, complete a guided setup, prove sign-in and recovery,
> invite a community administrator, and leave behind a healthy, observable,
> recoverable identity service that safely reconciles its own desired state.

Tessera is not operational when it merely starts. It is operational when a
different authorized team member can diagnose, maintain and recover it without
the original owner.

## The private-community advantage

Private community is a product constraint we use deliberately, not a smaller
version of public SaaS.

- One Shippin account and workspace context is already known; Tessera does not
  ask for it again.
- Shippin, Tessera and Zuul can ship tested version bundles instead of every
  possible component combination.
- Opinionated secure defaults replace a matrix of optional deployment choices.
- Community roles and recovery contacts can be established during onboarding.
- Desired-state blueprints can keep every private deployment converged.
- Protected diagnostics can be shared with the operating team without making
  customer identity data public.
- The owner retains custody and exportability while Shippin supplies the calm
  control surface and upgrade path.

This advantage does not permit hidden state. Every recommendation must expose
its effects, authority, rollback and recovery path before a write.

## Product boundaries

| system | owns | must not own |
|---|---|---|
| Shippin | member shell, account/workspace context, service catalog, commercial decisions and cross-product navigation | identity truth, credentials or mesh inventory |
| Tessera | subjects, organizations, authentication, federation, sessions, identity policy, keys, audit and identity desired state | billing rules, host inventory or network routes |
| Zuul | installation, nodes, peers, routes and private-mesh policy | a second user directory or reusable identity credential |
| Automaton and other consumers | local token verification and resource enforcement | token minting or provider-specific identity integration |

The browser calls a Shippin server-side adapter. The adapter calls Tessera's
provider-neutral management API. No browser receives an operator credential or
direct access to inherited administration APIs.

Customer and operator surfaces use Tessera names and vocabulary. Upstream names
remain only where provenance, third-party notices, standards history or source
maintenance require them; they do not leak through executable help, environment
keys, errors, routes or ordinary UI copy.

## Experience architecture

The persistent Shippin shell remains mounted at `/member/tessera`. Tessera
supplies the module navigation, data and actions through four layers:

1. **Start** — readiness, first-run outcomes and unfinished setup.
2. **Directory** — people, agent identities, teams, community membership and
   workspace bindings.
3. **Trust** — upstream company sign-in, downstream applications, Zuul
   enrollment and trust health.
4. **Operate** — sessions, audit, backups, upgrades, keys, recovery and safe
   diagnostics.

Advanced protocol and implementation details are available from each object;
they are not the primary navigation.

Every guided operation uses one grammar:

```text
choose outcome → gather discriminating facts → plan → review → apply → verify
```

`plan` is read-only. `apply` is authorized, audited and idempotent. `verify`
tests the customer's outcome rather than treating an API success as proof.

## Lifecycle contract

Each Tessera deployment has one explicit lifecycle state and one desired
revision.

| state | meaning | allowed operator action |
|---|---|---|
| `absent` | no instance is registered | plan installation |
| `preparing` | prerequisites are being checked | fix a named prerequisite or cancel |
| `initializing` | database and first instance are being created | observe; retry only through the same operation id |
| `needs_owner` | service is healthy but ownership/recovery is incomplete | enroll owner and recovery factors |
| `ready` | sign-in, trust, backup and management checks meet policy | normal operation |
| `degraded` | service runs but one required capability is unhealthy | follow typed remediation |
| `maintenance` | an upgrade, rotation or restore owns mutations | observe, resume or perform documented rollback |
| `recovery_required` | normal writes are stopped to protect state | run the guided recovery plan |
| `retired` | sign-in is disabled and retention policy is active | export or complete removal |

State transitions are server-owned. UI labels never infer `ready` from a
single process check. A readiness projection names every contributing check,
its source and observation time.

## Desired state and reconciliation

The existing blueprint engine becomes Tessera's reconciliation foundation.
Shippin and Zuul submit provider-neutral intent; Tessera produces and applies a
reviewable desired-state revision.

Every operation carries:

- a stable operation id and idempotency key;
- account, workspace, subject and actor, including `act` delegation;
- base and desired revisions;
- planned creates, changes and removals;
- required permissions and protected-secret slots;
- progress steps with safe retry classification;
- verification checks;
- rollback or recovery instructions;
- an immutable audit result.

Reapplying an achieved revision is a no-op. A stale reviewed plan returns a
typed conflict and must be replanned. Partial success is not reported as
complete; atomic resource groups roll back, and unavoidable external side
effects are recorded with a compensating action.

## Management API required before production UI writes

The read surface in `08-control-surface-and-federation.md` is extended with the
operational lifecycle:

```text
GET  /tessera/v1/capabilities
GET  /tessera/v1/overview
GET  /tessera/v1/operations/{operation_id}

POST /tessera/v1/installations/plan
POST /tessera/v1/installations/{plan_id}/apply
POST /tessera/v1/installations/{plan_id}/verify

POST /tessera/v1/guides/plan
POST /tessera/v1/guides/{plan_id}/apply
POST /tessera/v1/guides/{plan_id}/verify

POST /tessera/v1/backups/plan
POST /tessera/v1/restores/plan
POST /tessera/v1/upgrades/plan
POST /tessera/v1/trust/rotations/plan
GET  /tessera/v1/support-bundles/{bundle_id}
```

Plan/apply/verify resources share one operation envelope and typed error
catalog. Missing entitlement remains a typed `403`; absence of authentication
is `401`; stale state is `409`; an unmet prerequisite is a typed `422`; a
recoverable dependency outage is `503` with a safe retry classification.

Capability discovery prevents the panel from offering a path the deployed
version cannot execute.

## First-run contract

The first run is complete only when all of these are true:

1. deployment target, public domain, TLS boundary and PostgreSQL ownership are
   validated without exposing credentials;
2. protected master-key and database-secret references exist outside the
   repository and browser;
3. the first owner signs in and enrolls a passkey plus an independent recovery
   method;
4. at least two recovery contacts or an explicit single-owner risk acceptance
   are recorded;
5. issuer discovery, JWKS, token minting and consumer verification pass;
6. transactional email delivery, sender identity and domain status are tested
   before an invitation or recovery message depends on them;
7. a backup is created and a disposable restore verification succeeds;
8. the first community and delegated administrator role are created;
9. alert destinations and maintenance windows are confirmed;
10. the generated recovery packet is stored outside the Tessera deployment;
11. the owner sees a `ready` dashboard with no hidden incomplete step.

The setup may be left safely and resumed from another authorized browser. It
never relies on a one-page ephemeral wizard state.

## Federation quality bar

Upstream identity providers and downstream clients remain visibly separate.
For OIDC, SAML and LDAP connections Tessera must provide:

- discovery or metadata import where the standard permits it;
- safe credential entry with write-only values;
- domain, attribute, group and role mapping preview;
- account-linking and takeover-risk review;
- a test identity or protocol check before enablement;
- last success, last failure and expiring-certificate state;
- rotation without an avoidable authentication outage;
- an emergency local-owner path that does not depend on the failed upstream;
- disable and removal plans that state who will lose access.

Templates accelerate setup but do not bypass review. Provider-specific fields
remain adapter details behind Tessera resources.

SCIM and other directory provisioning are modeled separately from sign-in.
Provisioning creates, updates, suspends and removes subjects or memberships;
federation authenticates them. A connector may offer both, but the review and
audit must never imply that successful sign-in also guarantees lifecycle
deprovisioning.

## Delegated community operation

The smallest useful management roles are:

| role | responsibility | forbidden by default |
|---|---|---|
| platform owner | deployment, trust root, recovery and role delegation | silent impersonation |
| identity administrator | people, groups, federation and sign-in policy | infrastructure or billing mutation |
| community administrator | membership and bounded workspace access | key, provider or global-policy mutation |
| support operator | health, session diagnosis and approved revocation | credential reads or role grants |
| security auditor | immutable audit and posture review | all mutations |
| application integrator | assigned clients, redirects and test sign-in | unrelated clients or directory writes |

Dangerous actions require step-up authentication, an impact preview and an
audit reason. Tessera never makes the original owner the only recovery path.

## Operational quality bar

### Installation and upgrades

- One signed version manifest binds compatible Shippin, Tessera and Zuul
  versions.
- Preflight checks complete before mutation and print actionable remediation.
- Database migrations are forward-tested against the oldest supported version.
- Upgrade plans include backup evidence, compatibility checks and rollback
  limits.
- Repeating install or upgrade with the same operation id is safe.
- Single-node private community is a supported profile, not the only topology;
  documented multi-node behavior defines leader work, job ownership, draining
  and readiness before horizontal scaling is offered.

### Backup and recovery

- Backup includes the database, configuration revision and encrypted key
  material required by the documented recovery model.
- Secret values are never included in a browser support bundle.
- Restore is tested automatically in an isolated target, not assumed from a
  successful archive command.
- Recovery time objective and recovery point objective are measured from
  drills.
- Lost node, lost database, lost upstream IdP and lost administrator are
  separate runbooks and tests.

### Observability and support

- Health separates process, database, issuer, federation, signing, mail,
  backup and consumer-verification checks.
- Metrics and logs carry instance and operation identifiers but no identity
  credentials or authentication answers.
- Audit records configuration intent, actor, before/after revision and result.
- A support bundle is allowlisted, previewable and redacted before export.
- Every degraded state links to one runbook and one safe next action.

### Custody and portability

- Owners can export provider-neutral directory, membership, client, policy and
  audit data without private signing keys or password-equivalent material.
- Import is planned, previewed and collision-checked before mutation.
- Retention, suspension, deletion and legal-hold states are distinct; a delete
  button never invents data policy.
- Portable proxy, LDAP and RADIUS edges are independently supervised adapters.
  They receive narrowly scoped identity configuration and do not become a
  second authority or a hidden part of Zuul's mesh inventory.

### Security

- Asymmetric signing only; consumers never receive a shared verification
  secret.
- Master keys, private keys, client secrets, bootstrap passwords, OAuth
  sessions and application databases never enter the repository.
- Secret reads return presence, age and rotation state, never values.
- Recovery and high-impact operations require step-up and cannot be performed
  by a delegated agent without an explicit human-approved policy.
- Tenant isolation, account linking, redirect validation, SSRF boundaries and
  audit integrity receive dedicated adversarial tests.

## Source provenance, recognition and partnership claims

Tessera deliberately learns from and incorporates work from ZITADEL and
authentik. That history should be preserved accurately and generously.

The evidence baseline on 2026-08-18 is:

- the local ZITADEL source snapshot is `632a5196800c5919e5043d482846ec59d7fad88e`
  and its checkout is dirty; current upstream policy describes the main
  repository as AGPL-3.0-only with named Apache-2.0 and MIT directory
  exceptions;
- the local authentik source snapshot is
  `64d4b97dbd81a65a3dbe8e0e3dc7d13cb8831ff8` and its checkout is dirty;
  authentik's root license is MIT for content outside its listed website,
  enterprise and third-party exceptions;
- the operator reports supplemental approval from both organizations. The
  original correspondence, identity of the grantor, date and scope must be
  archived in protected business records. The repository stores only an
  evidence id and approved public wording, never private messages or tokens;
- donations are recorded as community support, not treated as proof of a
  license grant, endorsement or partnership.

Before any public release, create a source bill of materials that maps every
retained or adapted path to its source commit and applicable terms. Add the
top-level license, required notices, source-offer/compliance path and a public
provenance page. The exact current upstream references are:

- [ZITADEL licensing policy](https://github.com/zitadel/zitadel/blob/main/LICENSING.md)
- [ZITADEL brand and trademark policy](https://zitadel.com/docs/legal/policies/brand-trademark-policy)
- [authentik license](https://github.com/goauthentik/authentik/blob/main/LICENSE)

“Powered and inspired by” may link to upstream projects once each name and logo
use has been reviewed against the archived permission and current brand rules.
“Partner,” “approved,” “endorsed,” referral and commission language is only
published after the corresponding organization has confirmed that exact
relationship and disclosure wording in writing. Training material cites the
specific public source or concept it teaches.

This is how upstream generosity becomes a durable advantage: capture it once,
credit it accurately, and keep the implementation provenance machine-readable.

## Operational acceptance test

A release candidate passes only when an automated environment can demonstrate:

1. empty host to `ready` through the Shippin guide;
2. second application of the same desired state changes nothing;
3. owner passkey, recovery and delegated administrator all work;
4. upstream OIDC sign-in and emergency local-owner sign-in both work;
5. one downstream app and one Zuul installer enroll without reusable embedded
   secrets;
6. seat token verification succeeds for the intended workspace and fails for
   another workspace, tampering, `alg: none` and every `HS*` algorithm;
7. signed-in but unentitled requests receive the typed `403` body;
8. backup restores into an isolated replacement and passes sign-in plus token
   verification;
9. supported upgrade and rollback paths preserve audit and identity state;
10. loss of one node, the upstream IdP and the original owner each has a
    tested recovery path;
11. browser payloads, fixtures, logs, support bundles, image history and the
    repository contain no live secret;
12. mail, provisioning and portable identity-edge failures degrade only their
    named capabilities and retain the emergency owner path;
13. provider-neutral export and previewed import preserve identity references
    and tenant boundaries;
14. the source bill of materials and required notices match the shipped
    artifact.

Until every line passes, the dashboard must describe the capability as preview,
degraded or incomplete rather than operational.
