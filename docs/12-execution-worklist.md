# 12 — Execution worklist

**Status:** proposed dependency-ordered delivery plan.
**Outcome:** satisfy the operational acceptance test in
`11-operational-blueprint.md` without mixing unrelated work in one branch or
pull request.

## How this list is used

Each work item has one owner, one branch, one reviewable contract change and
one evidence-producing exit test. An item is not complete because code exists;
its **done when** statement must pass in automation or a recorded recovery
drill.

Statuses are `not started`, `in progress`, `blocked` or `done`. “Almost done”
is not a status. Dependencies are item ids, not implied ordering.

Every implementation pull request must state:

- contract or threat model changed;
- desired-state and migration effect;
- secret and tenant boundary;
- rollback/recovery effect;
- tests and operational evidence;
- source provenance for imported or adapted work.

## Program ledger

| epic | current state | starts when |
|---|---|---|
| P0 — release/provenance | ready | this plan is accepted |
| P1 — contracts/security | ready | this plan is accepted |
| P2 — control plane | waiting | P1 operation contracts are merged |
| P3 — installation | waiting | P2 operation foundation is proven |
| P4 — read surface | waiting | overview and adapter schemas are stable |
| P5 — guided writes | waiting | common guide grammar and relevant domain adapter exist |
| P6 — federation | waiting | secret and threat-model gates pass |
| P7 — delegation | waiting | permission and audit contracts pass |
| P8 — Zuul/edges | waiting | enrollment and portable-adapter contracts pass |
| P9 — operations | waiting | installation and secret boundaries pass |
| P10 — release | waiting | P3 through P9 meet their exit tests |

## P0 — Establish the release and provenance boundary

These items gate public artifacts and mutation-heavy feature work. They do not
block local research or read-only mocks.

| id | work | depends on | done when |
|---|---|---|---|
| P0.1 | Inventory every retained/adapted upstream path and generate a machine-readable source bill of materials with source repository, commit and license expression. | — | a clean checkout reproduces the inventory and fails CI on an unmapped imported path |
| P0.2 | Archive the operator-reported ZITADEL and authentik permissions in protected business records; store non-secret evidence ids, scope, dates and approved public wording in governance metadata. | — | a reviewer can resolve each evidence id without private correspondence entering Git |
| P0.3 | Decide and add Tessera's top-level license, notices, source availability path and distribution obligations using P0.1/P0.2. | P0.1, P0.2 | source and container distributions carry the reviewed files and CI verifies them |
| P0.4 | Approve the public provenance page and exact “powered/inspired by” wording; separately record any confirmed partner, referral or commission terms and disclosures. | P0.2, P0.3 | public copy cannot imply endorsement or partnership beyond archived approval |
| P0.5 | Pin clean upstream source snapshots and document the current dirty-checkout differences before accepting further imports. | P0.1 | imported changes are reproducible from a clean commit plus reviewed patch series |
| P0.6 | Inventory inherited customer/operator wording and namespaces; replace vendor-branded CLI, environment, error, route and ordinary UI surfaces while preserving required provenance and notices. | P0.1, P0.3 | automated scans permit upstream names only in a reviewed provenance/dependency allowlist |

## P1 — Freeze lifecycle, API and security contracts

| id | work | depends on | done when |
|---|---|---|---|
| P1.1 | Convert the lifecycle table into domain types, transition rules and typed refusal reasons. | — | table-driven tests cover every allowed and refused transition |
| P1.2 | Define the shared plan/apply/verify operation schema, idempotency rules, progress events and conflict semantics. | P1.1 | OpenAPI/protobuf schema and contract tests agree on retries and stale plans |
| P1.3 | Define the typed management error catalog for `401`, `403`, `409`, `422`, `429` and `503`. | P1.2 | Shippin adapter fixtures render a distinct remedy for every type |
| P1.4 | Define capability discovery and version compatibility for Shippin, Tessera and Zuul. | P1.2 | unsupported UI paths are hidden or disabled from server facts, never guessed |
| P1.5 | Write threat models for bootstrap, account linking, upstream federation, downstream clients, recovery, support bundles and Zuul enrollment. | P1.2 | every threat has a prevention, detection and recovery control with a test owner |
| P1.6 | Define the delegated role/permission matrix and step-up rules. | P1.3, P1.5 | tests prove each role's allowed and forbidden actions, including typed `403`s |
| P1.7 | Define audit event schema and retention/export contract. | P1.2, P1.6 | every mutation contract names its required audit event and actor chain |

## P2 — Build the operational control plane

| id | work | depends on | done when |
|---|---|---|---|
| P2.1 | Add durable operation storage, idempotency keys, leases and progress events. | P1.2, P1.7 | duplicate requests converge on one operation and a crashed worker resumes safely |
| P2.2 | Implement lifecycle projection from explicit checks and revisions. | P1.1, P2.1 | no single process check can incorrectly produce `ready` |
| P2.3 | Extend blueprint application into a reconciler with plan, diff, stale-revision rejection and verification hooks. | P1.2, P2.1 | apply twice is a no-op; concurrent stale apply returns typed `409` |
| P2.4 | Implement protected secret references and write-only secret slots. | P1.5, P2.1 | reads expose presence/age only and scans find no value in logs, events or browser payloads |
| P2.5 | Implement capability and operation endpoints. | P1.4, P2.1, P2.2 | management clients can resume an operation after disconnect without guessing state |
| P2.6 | Add operation cancellation and compensating-action rules. | P2.3 | each cancellable step is proven safe; non-cancellable steps explain why and how to recover |

## P3 — Make installation repeatable

| id | work | depends on | done when |
|---|---|---|---|
| P3.1 | Model deployment prerequisites: supported runtime, PostgreSQL, domain, TLS boundary, ports, storage and secret provider. | P1.4, P1.5 | preflight is read-only and returns typed, actionable failures |
| P3.2 | Implement installation plan/apply/verify using the operation engine. | P2.2–P2.5, P3.1 | empty supported host reaches `needs_owner`; retry with same id changes nothing unexpectedly |
| P3.3 | Support Shippin Cloud supervision and rootless private Podman from one signed image and version manifest. | P0.3, P1.4, P3.2 | both targets pass the same conformance suite and use `TESSERA_*` configuration |
| P3.4 | Implement resumable owner enrollment with passkey and independent recovery. | P1.5, P1.6, P3.2 | closing the browser mid-setup and resuming elsewhere is tested |
| P3.5 | Generate the external recovery packet and enforce its acknowledgement. | P3.4 | packet contains no reusable browser-disclosed server secret and restores access in a drill |
| P3.6 | Create the first community, recovery contacts and delegated administrator. | P1.6, P3.4 | a delegated administrator completes a bounded member task without owner access |
| P3.7 | Configure and verify sender domain, transactional email and notification delivery without returning provider credentials. | P2.4, P3.2 | invite and recovery messages pass outcome tests before either feature is marked ready |

## P4 — Deliver the read-first control surface

| id | work | depends on | done when |
|---|---|---|---|
| P4.1 | Implement `GET /tessera/v1/overview` with source revision and observation times. | P2.2, P2.5 | fixture and live projection satisfy one schema and missing facts never appear healthy |
| P4.2 | Publish `shippin.tessera` in the service catalog and add the server-side management adapter. | P1.3, P1.4, P4.1 | browser traffic contains only member-scoped calls and no operator credential |
| P4.3 | Implement `/member/tessera` shell context and Start dashboard. | P4.2 | clicking Tessera updates the persistent shell without opening a vendor console |
| P4.4 | Implement Infrastructure, AI and Customers lenses over one authority. | P4.1, P4.3 | the views share stable subject/workspace ids and do not duplicate identity state |
| P4.5 | Implement consequence-first health, activity and remediation components. | P1.3, P4.1 | every degraded card links to a safe next action and optional diagnostics |
| P4.6 | Add accessibility, responsive behavior and empty/loading/degraded states. | P4.3–P4.5 | keyboard, screen-reader and narrow-layout acceptance tests pass |

## P5 — Turn the six guides into real operations

| id | work | depends on | done when |
|---|---|---|---|
| P5.1 | Implement the common guide planner, review model and operation progress UI. | P2.3, P4.3 | no guide invents its own apply semantics or hides a removal/permission effect |
| P5.2 | Implement “Invite my team.” | P3.6, P5.1 | invite, membership, seat and sign-in verification complete as one resumable operation |
| P5.3 | Implement “Use company sign-in.” | P6.1–P6.5, P5.1 | mapped test sign-in succeeds before enablement and local recovery remains available |
| P5.4 | Implement “Connect an app.” | P6.6–P6.8, P5.1 | discovery, redirect, grant, audience and token verification checks pass |
| P5.5 | Implement “Add an AI agent.” | P1.6, P5.1 | service identity, agent seat, scopes and `act` policy are reviewed and audited |
| P5.6 | Implement “Set up private access.” | P8.1–P8.4, P5.1 | one Zuul target enrolls without reusable embedded credentials |
| P5.7 | Implement “Secure or recover access.” | P3.4, P7.4, P9.3, P5.1 | passkey/MFA, session concern and recovery branch to tested outcome checks |

## P6 — Complete federation safely

| id | work | depends on | done when |
|---|---|---|---|
| P6.1 | Normalize upstream OIDC, SAML and LDAP resources behind Tessera API types. | P1.5, P2.4 | inherited provider types do not cross the browser or Shippin adapter boundary |
| P6.2 | Implement discovery/metadata import and SSRF-safe fetch policy. | P6.1 | hostile URL, redirect and oversized metadata tests fail closed |
| P6.3 | Implement write-only credentials, certificate status and rotation plans. | P2.4, P6.1 | a provider rotates without reads returning secret material |
| P6.4 | Implement mapping preview for domain, subject, attributes, groups and roles. | P6.1 | preview uses a non-persisted test assertion and identifies collisions |
| P6.5 | Implement account-link and takeover protections plus emergency local-owner policy. | P1.5, P6.4 | conflicting identity and failed upstream scenarios cannot seize or lock out ownership |
| P6.6 | Normalize downstream OIDC and SAML clients, grants, redirects, audiences and occupant classes. | P1.5, P2.4 | client resources remain provider-neutral and workspace-bounded |
| P6.7 | Implement public/PKCE and confidential client creation and rotation. | P6.6 | confidential values are shown once through a protected path and never returned later |
| P6.8 | Implement client verification for discovery, redirect and seat-token consumption. | P6.6, P6.7 | Automaton and a reference customer app pass positive and negative token tests |
| P6.9 | Add federation health jobs, expiry warnings and safe diagnostics. | P6.2–P6.8 | simulated outage and expiring certificate produce actionable degraded states |
| P6.10 | Normalize SCIM provisioning separately from sign-in and implement previewed create/update/suspend/deprovision operations. | P1.5, P2.3, P6.1 | provisioning failure cannot silently leave a removed upstream user active |

## P7 — Make community delegation calm

| id | work | depends on | done when |
|---|---|---|---|
| P7.1 | Implement community, group, membership and role desired-state resources. | P1.6, P2.3 | reviewed blueprints and guided writes converge on the same state |
| P7.2 | Build team queues for invitations, access requests and expiring access. | P4.4, P7.1 | a community administrator clears the queue without global permissions |
| P7.3 | Add access review and offboarding plans with impact previews. | P7.1, P7.2 | removing a person lists sessions, clients, workspaces and recovery effects first |
| P7.4 | Add session search, diagnosis and bounded revocation. | P1.6, P4.5 | support can revoke the intended sessions without viewing credentials or broad directory writes |
| P7.5 | Add immutable audit exploration and export for auditors. | P1.7, P4.3 | actor/subject delegation, revision and result are queryable for every mutation |
| P7.6 | Add notifications and escalation routing for owner, identity and community roles. | P7.1–P7.5 | each alert class reaches a role capable of its documented remedy |
| P7.7 | Implement provider-neutral export, collision-previewed import, retention, suspension and deletion workflows. | P1.5, P2.3, P7.1, P7.5 | a private owner can move their configuration and directory without exporting private keys or password equivalents |

## P8 — Bind Zuul without merging domains

| id | work | depends on | done when |
|---|---|---|---|
| P8.1 | Define the Tessera–Zuul enrollment contract, scopes, audience and audit correlation ids. | P1.2, P1.5 | contract names identity facts only; no mesh inventory enters Tessera |
| P8.2 | Implement browser PKCE enrollment for interactive private access. | P6.6–P6.8, P8.1 | enrollment binds one subject and workspace and survives callback replay tests |
| P8.3 | Implement device or guided one-time enrollment for headless targets. | P2.4, P8.1 | installer contains no reusable shared secret and expired codes fail safely |
| P8.4 | Implement enrollment verify, revoke and re-enroll operations. | P8.2, P8.3 | lost target and operator-change drills leave no orphan reusable credential |
| P8.5 | Expose Zuul identity attachment health in the Infrastructure lens. | P4.4, P8.4 | UI links correlated identity/enrollment state without copying peer/routes state |
| P8.6 | Define the portable identity-edge contract for forward-auth proxy, LDAP and RADIUS adapters. | P1.4, P1.5, P6.6 | adapters receive narrow, versioned configuration and never become identity authorities |
| P8.7 | Package, supervise, rotate and health-check portable identity edges independently from Zuul mesh state. | P2.4, P3.3, P8.6 | loss or upgrade of one edge degrades only its named capability and preserves core sign-in |

## P9 — Build recovery, maintenance and support

| id | work | depends on | done when |
|---|---|---|---|
| P9.1 | Define backup contents, encryption, retention, RPO/RTO and secret-provider responsibilities. | P1.5, P2.4 | backup contract names every required and deliberately excluded artifact |
| P9.2 | Implement scheduled backup and isolated automated restore verification. | P9.1, P3.3 | dashboard reports last verified restore, not merely last archive |
| P9.3 | Implement guided restore for lost database/node and guided owner recovery. | P3.5, P9.2 | fresh replacement reaches verified `ready` within measured RTO |
| P9.4 | Implement signed version manifests, upgrade preflight, maintenance state and compatibility checks. | P1.4, P2.2, P3.3 | oldest supported version upgrades in automation with state and audit preserved |
| P9.5 | Implement rollback boundaries and forward-recovery instructions per migration. | P9.4 | failure injection at each migration step has a proven supported outcome |
| P9.6 | Implement signing/master/client key rotation plans and overlapping verification windows. | P2.4, P6.3 | consumers continue verifying across rotation and retired keys stop signing |
| P9.7 | Implement allowlisted, redacted and previewable support bundles. | P1.5, P4.5 | seeded credentials and personal data do not appear in exported bundles |
| P9.8 | Run loss-of-upstream, loss-of-node, loss-of-database and loss-of-owner drills. | P6.5, P9.2–P9.7 | each drill has measured result, evidence and an assigned remediation owner |
| P9.9 | Define and test single-node and multi-node topology, leader jobs, draining, queue ownership and capacity limits. | P2.1, P3.3, P9.4 | offered topology survives node replacement without duplicate operations or false readiness |

## P10 — Release hardening

| id | work | depends on | done when |
|---|---|---|---|
| P10.1 | Build the full golden-path conformance environment and browser automation. | P3–P9 | every operational acceptance assertion runs from a clean environment |
| P10.2 | Add tenant-isolation, redirect, account-link, CSRF, SSRF, token, session and audit-integrity adversarial suites. | P1.5, P6–P9 | negative tests fail closed with typed, non-leaking errors |
| P10.3 | Add load and soak tests for sign-in, federation health, projections, audit and reconciliation. | P2–P9 | agreed service objectives hold without unbounded queues or database growth |
| P10.4 | Add dependency, container, SBOM, signature and secret scanning release gates. | P0.1, P3.3 | each artifact is signed, attributable and reproducible from the release commit |
| P10.5 | Complete accessibility, localization and guided-language review with non-expert operators. | P4–P9 | representative operators complete install, federation and recovery without protocol coaching |
| P10.6 | Run a two-person handoff drill: the original owner is unavailable and the delegated team operates and recovers the service. | P7, P9, P10.1 | the team succeeds within the runbook and SLO without undocumented owner knowledge |
| P10.7 | Review public recognition, source offer, notices and any partner/referral disclosures for the release artifact. | P0.3, P0.4, P10.4 | shipped UI, docs, training and artifact metadata use only approved wording |

## The first four delivery slices

Do not start every epic at once. The first four independently demonstrable
slices are:

1. **Provenance and operation contracts:** P0.1–P0.6 plus P1.1–P1.4. This
   removes release ambiguity and freezes the API grammar.
2. **Read-only operational truth:** P1.5–P2.5 plus P4.1–P4.5. Clicking Tessera
   shows real, sourced health and access state without a dangerous write path.
3. **Repeatable private install:** P2.6 plus P3.1–P3.7 and P9.1–P9.2. A clean
   host reaches a delegated, restore-verified `ready` state.
4. **One real guide:** P5.1, P6.1–P6.5 and P5.3. “Use company sign-in” becomes
   the reference plan/apply/verify workflow before the other five guides copy
   its grammar.

Only after slice four should the team fan out across the remaining guide,
delegation, Zuul and maintenance work.

## Program-level done

The program is complete when:

- every work item is `done` or explicitly removed by a contract decision;
- every operational acceptance assertion passes from a clean environment;
- the two-person handoff drill passes;
- the worktree and shipped images contain no live secret;
- source provenance and public recognition match the released artifact;
- the Shippin dashboard can truthfully call the deployment `ready`.
