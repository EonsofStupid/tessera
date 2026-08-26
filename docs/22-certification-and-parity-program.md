# 22 — Capability parity and certification program

**Status:** active; product-owner scope approved 2026-08-21.
**Last external-requirement review:** 2026-08-21.
**Product boundary:** `02-standalone-product-contract.md`.
**Promotion standard:** `20-capability-delivery-standard.md`.
**Governance record:** `23-release-governance-record.md`.
**Work preservation:** `24-work-preservation-inventory.md`.

## Outcome

Nomen becomes a standalone identity and access management product that can
truthfully demonstrate the capabilities selected in this program. Each promoted capability is backed by a
Nomen-owned contract, persistence boundary, API, guided browser workflow,
authorization model, audit evidence, protocol or real-target conformance,
negative tests, and supported operational lifecycle.

AngryVibes LLC owns the commercial product. Nomen is offered privately to its
users as either a standalone deployment or a Shippin-managed deployment. A
public commercial entry page may explain the product and begin authorized
sign-in, but it does not create public enrollment, expose management facts or
make Shippin a dependency of the standalone product.

The program distinguishes four different claims:

1. **implemented** — implementation material exists;
2. **conformant** — a named standards or interoperability suite passes;
3. **operational** — the complete Nomen evidence manifest passes for a
   declared release and deployment profile; and
4. **certified or attested** — a named external body has issued or published
   the applicable certificate, certification, validation, or report.

These terms are not interchangeable. A running inherited endpoint is not a
Nomen capability. A protocol test is not an operational release. An
organizational audit is not a product protocol certification.

## Current baseline

The development bundle is Nomen `1.0.0-alpha` (frozen). It has a real PostgreSQL-backed runtime,
a Nomen-owned React/TypeScript application, OIDC discovery, a PKCE sign-in
handoff, capability discovery, and Playwright running against the production
container boundary. The current capability service advertises every named
capability as `preview` and disabled. No capability is currently promoted as
`operational`.

Inherited implementation material is inventory. It is absorbed
capability-by-capability; Nomen does not run two identity authorities or
bulk-merge stale branches.

## Non-negotiable program controls

- Change the product or protocol contract before behavior.
- Keep PostgreSQL as the identity, policy, operation and audit authority.
- Use asymmetric signatures; never add an `HS*` verification profile.
- Return typed `401` for missing authentication and typed `403` for missing
  permission or entitlement.
- Keep every capability disabled until its release-bound evidence passes.
- Never store live secrets, private correspondence, OAuth sessions, private
  keys, databases or waiver documents in Git.
- Use one Nomen-owned management command for API, browser, CLI and AI action
  surfaces.
- Exercise promoted capabilities against real disposable targets and retain
  sanitized failure evidence.
- Preserve upstream provenance and required notices without leaking upstream
  product identity into Nomen's runtime surface.
- Do not claim certification, validation, partnership or endorsement without
  the exact issuing body, scope, release/version, date and evidence identifier.

## Source, permission and distribution gate

Hosted production and controlled managed-container distribution are blocked
until this gate passes. A public informational page is not evidence that the
software artifact is legally or operationally releasable.

| ID | Work | Required evidence |
|---|---|---|
| L0.1 | Inventory all tracked paths and generated artifacts by origin, source revision and license expression | deterministic source BOM and CI failure for unmapped imported paths |
| L0.2 | Archive inbound source permissions or waivers outside Git | protected record identifiers, grantor authority, date, exact scope, modification/distribution rights and approved public wording |
| L0.3 | Confirm no third-party product license is absorbed; counsel reviews trademarks and any protected permission evidence outside Git | signed legal disposition that Nomen's product license is AngryVibes LLC and shippin.ai only; dependency SBOMs remain a G1 control |
| L0.4 | Keep `LICENSE` and `NOTICE` as AngryVibes LLC and shippin.ai only | release image and source archive contain those files; no ZITADEL, Authentik, AGPL, or Apache product license is shipped as Nomen |
| L0.5 | Pin clean upstream snapshots and review every retained patch | reproducible source mapping from an immutable upstream revision plus Nomen patch history |
| L0.6 | Harvest existing branches without bulk merging | per-path decisions for `agent/nomen-provenance-inventory`, `agent/nomen-operational-plan`, LDAP branches and other existing work |
| L0.7 | Approve public product and provenance language | no unapproved partnership, endorsement or trademark claim |

The product-owner license is recorded: **AngryVibes LLC and shippin.ai only**.
Nomen absorbs no third-party product license. Git stores `LICENSE`, `NOTICE`,
and provenance identifiers. Git does not store waiver text. Counsel still
archives protected permission evidence outside Git (`docs/23-release-governance-record.md`).

## Union capability ledger

Each stable ID below becomes an independently discoverable capability or a
documented sub-capability of the named parent. An omitted capability is not
supported.

| Family | Stable capability IDs and required outcomes | Current Nomen class |
|---|---|---|
| Standalone lifecycle | `deployment_preflight`, `guided_setup`, `owner_enrollment`, `deployment_operations`, `backup_restore`, `upgrade_rollback`, `key_rotation`, `high_availability` | baseline or preview; no complete operational evidence |
| Directory and tenancy | `organizations`, `users`, `groups`, `memberships`, `attributes_metadata`, `service_accounts`, `invitations`, `delegated_administration`, `tenancy_isolation` | inherited inventory; not independently promoted |
| Authentication | `password_authentication`, `passkeys_webauthn`, `totp`, `email_otp`, `sms_otp`, `static_recovery_codes`, `duo`, `multi_factor_chains`, `account_recovery`, `step_up_authentication` | inherited/flow inventory; not operational |
| Sessions and consent | `session_management`, `session_revocation`, `refresh_family_rotation`, `consent_management`, `security_events`, `impersonation_controls` | inherited inventory; not operational |
| Authorization | `projects`, `applications`, `roles`, `user_grants`, `project_grants`, `application_entitlements`, `policy_engine`, `access_reviews` | inherited and Nomen domain inventory; not operational |
| Downstream protocols | `downstream_oidc`, `oauth_device_authorization`, `oauth_token_exchange`, `private_key_jwt`, `downstream_saml`, `ws_federation`, `token_introspection`, `token_revocation`, `front_back_channel_logout` | OIDC/SAML preview; remaining items not advertised |
| Upstream federation | `upstream_oidc`, `upstream_saml`, `upstream_ldap`, `upstream_kerberos`, `upstream_jwt`, `social_identity_sources`, `account_linking`, `federation_mapping`, `emergency_local_owner` | OIDC/SAML preview; remaining items not advertised |
| Provisioning and sync | `scim_service_provider`, `scim_source`, `google_workspace_sync`, `microsoft_entra_sync`, `provisioning_preview`, `deprovisioning` | not advertised |
| Access edge | `ldap_outbound`, `ldap_inbound`, `radius`, `forward_auth`, `identity_aware_proxy`, `remote_access_control`, `edge_outposts` | LDAP/proxy/forward-auth preview; remaining items not advertised |
| Flows and extensibility | `visual_flow_engine`, `authentication_flows`, `authorization_flows`, `enrollment_flows`, `recovery_flows`, `conditional_policies`, `expression_actions`, `token_claim_actions`, `saml_response_actions`, `declarative_blueprints` | implementation inventory or preview; not operational |
| Security extensions | `mutual_tls`, `device_trust`, `shared_signals`, `password_history`, `reputation_policy`, `geoip_policy`, `captcha`, `certificate_lifecycle`, `fips_profile` | not advertised; FIPS profile requires separate module evidence |
| Product surfaces | `public_entry`, `management_ui`, `end_user_account`, `management_api`, `cli`, `action_discovery`, `localization`, `accessibility` | public/UI baseline; complete management outcomes missing |
| Audit and notification | `immutable_audit`, `audit_search_export`, `notification_rules`, `email_delivery`, `sms_delivery`, `webhook_delivery`, `support_bundles` | partial inherited/Nomen inventory; not operational |
| Supply chain | `signed_oci`, `sbom`, `source_provenance`, `build_provenance`, `vulnerability_response`, `release_verification` | planned or partial; no signed release gate |

Legacy compatibility profiles such as OAuth implicit or hybrid response types
may be supported for migration, but they are disabled by default, isolated from
the secure default profile, and tested against the current OAuth Security Best
Current Practice. Resource Owner Password Credentials is not a Nomen target.

## Delivery sequence

### Gate G0 — Establish authority and preserve work

1. Complete L0.1–L0.7.
2. Capture the dirty worktree and all existing branches in a non-destructive
   inventory: commit, purpose, unique files, tests, conflicts and disposition.
3. Extract useful source-BOM, operational-plan and LDAP work as reviewed path
   patches; do not merge entire historical branches.
4. Create the canonical capability ledger and machine-readable schema.
5. Add a program dashboard generated from evidence, never edited into green.

**Exit:** distribution rights are documented, no known work is lost, and every
capability has a stable ID, current class, dependency and evidence owner.

### Gate G1 — Build the release evidence factory

1. Create a clean disposable PostgreSQL plus production-image harness.
2. Add CI lanes for domain, migration/RLS, API, browser, protocol, hostile
   input, restart, backup/restore and supported upgrade testing.
3. Generate SPDX or CycloneDX source/container SBOMs.
4. Produce SLSA v1.2 provenance from a hosted isolated builder.
5. Sign OCI images and attestations with Cosign/Sigstore; verify by digest and
   expected CI identity before deployment.
6. Add dependency, container, secret, license and provenance policies.
7. Define vulnerability disclosure, supported-version and security-update
   policies.

**Exit:** one digest-addressed development image carries verifiable signature,
SBOM, provenance and test manifests. This proves the factory, not IAM parity.

### Gate G2 — Promote the first operational IAM slice

Execute `21-first-iam-vertical-slice.md` without adding unrelated dashboards:

1. preflight and resumable initialization;
2. passkey first-owner enrollment and independent recovery;
3. organization, role and public OIDC application creation;
4. member invitation and enrollment;
5. Authorization Code with PKCE, consent and asymmetric tokens;
6. typed `401`/`403`, tenant denial and hostile redirect/PKCE/replay tests;
7. session and refresh-family revocation;
8. immutable audit and equivalent API/browser/CLI/action commands; and
9. restart plus destructive disposable restore.

Integrate the OpenID Foundation conformance suite into CI during development.
Do not pay or submit for certification until the same release candidate passes
the complete Nomen operational evidence gate.

**Exit:** the first signed release digest may advertise only its proven OIDC,
owner, organization, application, invitation, session and audit capabilities as
operational.

### Gate G3 — Complete the IAM core

Deliver vertical slices for:

1. human and service-account lifecycle;
2. groups, metadata, projects, roles, user grants and delegated project grants;
3. TOTP, email/SMS OTP, recovery codes, step-up and authentication policy;
4. end-user profile, authenticators, sessions, consent and recovery;
5. upstream OIDC/SAML/JWT/social identity brokering and safe account linking;
6. downstream SAML with metadata, signing, encryption, SLO and rollover;
7. device authorization, client credentials/private-key JWT, refresh rotation
   and token exchange;
8. SCIM User service-provider lifecycle, explicitly retaining any Group
   limitation until implemented and proven;
9. actions for internal/external authentication, token claims and SAML
   responses;
10. branding, custom domains/text, SMTP/SMS and organization self-service; and
11. exhaustive audit exploration and stable REST/gRPC/CLI automation.

**Exit:** every IAM-core capability in the ledger is operational or has a
named, explicitly accepted exclusion. Inherited code alone cannot close a row.

### Gate G4 — Complete federation and the access edge

Deliver independently supervised slices for:

1. LDAP and Kerberos upstream sources with mapping and failover;
2. LDAP provider with bind/search/groups/paging/StartTLS/LDAPS and isolation;
3. identity-aware proxy and single-application/domain-level forward-auth;
4. header, cookie, websocket, logout and bypass-resistance behavior;
5. RADIUS authentication and attribute mappings;
6. remote access control for SSH, RDP and VNC with endpoint policy;
7. portable proxy, LDAP, RADIUS and RAC outposts with least-privilege service
   identity, live configuration, health and air-gapped operation;
8. SCIM source and outbound provisioning;
9. Google Workspace and Microsoft Entra lifecycle synchronization;
10. WS-Federation and Shared Signals Framework;
11. visual authentication/authorization/enrollment/recovery flow editing,
    simulation, versioning and rollback; and
12. reputation, GeoIP, CAPTCHA, event and expression policies.

**Exit:** each edge can fail or upgrade without corrupting identity state or
disabling emergency owner access. Linux and Windows/AD claims are promoted
separately.

### Gate G5 — Enterprise security and commercial operations

1. mTLS/client-certificate authentication and certificate lifecycle.
2. Device-trust signals and continuous-access/shared-signal processing.
3. Password history and organization security policy inheritance.
4. High availability, capacity, load, soak and dependency-failure evidence.
5. Verified backups, RPO/RTO measurement, upgrades and rollback boundaries.
6. Redacted support bundles, audit export and customer data/config export.
7. Managed-access approval, expiry, revocation and customer offboarding.
8. Dedicated and community tenancy conformance with PostgreSQL RLS.
9. Optional FIPS deployment profile using an identified FIPS 140-3 validated
   module in its approved operational environment.

**Exit:** the commercial release gate in `02-standalone-product-contract.md`
passes from a clean host for one immutable image digest.

### Gate G6 — Independent assurance and public certification

External submission happens only after the corresponding internal gate is
stable. Fixes are made in the ordinary product branch and the entire affected
evidence manifest is rerun.

| Track | Nature | When | Required result |
|---|---|---|---|
| OpenID Connect OP | formal OpenID Foundation self-certification | after G2 release candidate | applicable Basic/Config and logout profiles published for the Nomen deployment; add Dynamic/other profiles only when implemented |
| OpenID Connect RP | formal certification only if Nomen ships a generic RP/source implementation worth certifying separately | after upstream OIDC in G3 | applicable RP profiles published |
| FIDO2 server | formal FIDO Alliance functional certification | after passkey registration/authentication/recovery is stable | self-validation, official interoperability testing, submission and certificate |
| SAML 2.0 | standards conformance and independent interoperability evidence; no assumed universal product certificate | after G3 SAML | OASIS conformance matrix plus Shibboleth, Microsoft Entra ID, AD FS, Okta and another independent implementation |
| SCIM 2.0 | RFC conformance and interoperability evidence | after SCIM slice | RFC 7643/7644 matrix, Okta and Entra provisioning, bulk/error/patch/filter tests; limitations published |
| LDAP/AD/Kerberos | protocol and platform interoperability evidence | after G4 directory slices | OpenLDAP plus supported Microsoft AD/Windows matrix, TLS, paging, groups, failover and hostile input |
| RADIUS | protocol/client interoperability evidence | after G4 RADIUS | FreeRADIUS plus supported VPN/network clients and negative packet tests |
| Proxy/forward-auth | real reverse-proxy/application interoperability evidence | after G4 proxy | NGINX, Traefik, Caddy and Envoy; websocket, streaming, cookies, headers and bypass tests |
| Web application security | independent assessment | after G5 hardening | OWASP ASVS 5.0 target matrix, external penetration test, remediation and clean retest |
| NIST digital identity | implementation profile, not a NIST product certificate | during G2–G5 | documented SP 800-63-4 AAL/federation mappings and explicit non-applicable identity-proofing claims |
| ISO/IEC 27001:2022 | organization/ISMS certification | managed-service operations are stable | accredited certification body issues certificate for the defined organizational/service scope |
| SOC 2 | independent CPA attestation, not a product certificate | after controls have operated for the selected observation period | Type I if commercially necessary, followed by Type II; report scope names the managed Nomen service |
| FIPS 140-3 | cryptographic-module validation or `FIPS 140-3 Inside` claim | only for a selected government profile | exact active module certificate, version, operational environment and approved-mode evidence; never call Nomen itself validated unless it undergoes CMVP |
| Accessibility | standards audit and customer-facing report | before general availability | WCAG 2.2 AA evidence and current VPAT/ACR if public-sector procurement requires it |
| Supply chain | verifiable artifact assurance | from G1 onward | signed OCI, SBOM, vulnerability report and SLSA provenance verified before promotion |

As of the review date, OpenID's official fee page lists OpenID Connect
certification at USD 700 for members or USD 3,500 for non-members per new
deployment, with multiple OIDC profiles allowed for that deployment during the
calendar year and separate OP/RP payments. Open-source implementations may be
eligible for a waiver under the Foundation's policy. The FIDO Alliance lists
FIDO2 server functional certification at USD 6,000 for members or USD 9,000
for non-members, excluding internal labor and any additional testing expense.
Fees are planning figures only and must be rechecked before purchase.

## Evidence manifest

Every promoted capability publishes one immutable manifest bound to the source
revision and OCI digest. It contains digests or stable references for:

- product contract and threat model;
- source-origin and license disposition;
- database migrations, rollback and supported PostgreSQL matrix;
- API/protobuf/OpenAPI schemas and typed error fixtures;
- authorization and cross-tenant negative results;
- protocol standards/profile selection and conformance report;
- real-target interoperability matrix;
- Playwright production-image journeys and accessibility results;
- API/browser/CLI/action equivalence results;
- audit redaction and correlation evidence;
- restart, dependency-loss, backup/restore and upgrade results;
- SBOM, vulnerability results, image signature and build provenance;
- external report or certificate identifier, scope and expiry when applicable;
- known limitations and exact operator remediation; and
- approving release-policy revision.

The runtime accepts `operational` only when the manifest validates and every
required component is compatible. Revoked, expired or mismatched evidence
degrades or disables the capability.

The normative wire shape is
`backend/v1/domain/evidence-manifest.schema.json`; validation and proof
derivation live beside the capability domain. Evidence artifacts remain in a
protected release-evidence store. The Git object contains only its contract,
validator, generated test material and immutable digests—never waiver text,
credentials, private keys, customer data or raw assessment payloads.

## Collaboration checkpoints

The product owner and engineering work together at explicit decision points.
Private values are entered directly into the relevant provider or protected
business record, never pasted into source files or chat logs.

| Checkpoint | Product owner supplies or decides | Engineering returns |
|---|---|---|
| C0 — legal identity | AngryVibes LLC and shippin.ai as the only product licensors; Jesse Hall product authority; private standalone/Shippin-managed distribution; protected permission evidence IDs stay outside Git | source BOM with no absorbed third-party product license, `LICENSE`/`NOTICE`, counsel question set |
| C1 — release estate | source host, CI identity, OCI registry, public domain/DNS and security contact | signed development artifact and verification instructions |
| C2 — market scope | private commercial users; standalone and Shippin-managed profiles; regulated-market claims require a later explicit decision | ordinary commercial assurance baseline and explicit regulated-market exclusions |
| C3 — communications | sender domain, SMTP provider and whether SMS is required | verified SPF/DKIM/DMARC/delivery evidence and recovery limitations |
| C4 — custody | cloud/KMS/HSM choice, backup custody, recovery contacts and FIPS requirement | secret-reference, rotation, restore and approved-mode design |
| C5 — certification spend | OIDF membership/payment, FIDO membership/payment, audit and penetration-test budgets | submission-ready evidence packs and vendor/auditor question sets |
| C6 — release approval | accepted known limitations and public claims | immutable release manifest, runbooks and go/no-go report |

## Planning estimate and staffing reality

The complete union is a product program, not a two-week feature. Before the
branch-by-branch absorption audit, the honest order-of-magnitude estimate is
approximately **120–220 senior engineer-weeks**, plus product/security review,
legal counsel, independent penetration testing and organizational audit work.
Existing inherited implementation can reduce coding time, but it does not
remove Nomen API/UI, tenancy, recovery, conformance or operational evidence.

A practical core team is:

- one IAM/protocol lead;
- two backend/security engineers;
- one frontend/product engineer;
- one platform/release engineer;
- fractional product design/accessibility, security assurance and legal; and
- independent penetration tester/auditor at the applicable gates.

With fewer people, preserve the gate order and extend the calendar. Do not run
all capability families in parallel before G2 establishes the reference
vertical-slice pattern.

## First execution tranche

The first tranche contains no speculative feature implementation:

1. preserve and classify the current dirty worktree;
2. inventory all existing branches and produce per-path harvest decisions;
3. port and regenerate the source-provenance inventory against current `main`;
4. resolve top-level license/notice and private-permission questions with
   counsel;
5. add the machine-readable union capability ledger;
6. create the evidence-manifest schema and validator;
7. establish hosted CI, OCI registry identity, SBOM, provenance and signing;
8. expand the production Playwright harness into the eight projects in
   `21-first-iam-vertical-slice.md`;
9. integrate the local OpenID conformance suite; and
10. implement G2 in contract order until one release digest is genuinely
    operational.

Only then does execution fan out into G3 and G4.

### Execution record — 2026-08-22

| Tranche item | Current result | Next acceptance boundary |
|---|---|---|
| 1–2 — preserve work | Complete in the current worktree: `24-work-preservation-inventory.md` records every branch and prohibits bulk merge/deletion | refresh before any branch is archived |
| 3 — source provenance | Deterministic manifest, generated BOM and fail-closed validator pass for 3,962 currently tracked paths | regenerate after these worktree files become part of a release commit |
| 4 — legal disposition | Product/distribution facts and counsel questions are recorded; private grant text remains outside Git | Jesse Hall supplies protected evidence record IDs and counsel supplies the signed disposition at C0 |
| 5 — capability ledger | 112 union capabilities and 3 supporting runtime capabilities are schema-described, embedded and contract-drift tested | every capability stays `preview` until its own evidence passes |
| 6 — evidence factory contract | The v1 schema, validator and proof-digest derivation are implemented and tested across all 18 evidence layers | G1 must produce actual digest-addressed artifacts; generated test values are not release evidence |
| 7 — signed supply chain | Not started; hosted CI/registry identity is intentionally not guessed | select the C1 source host, builder identity and private OCI registry |
| 8 — browser evidence | Nine production-runtime Chromium journeys have passed: the original six, typed public denial, deployment workbench, and a real one-time bootstrap-owner journey using Chromium's virtual authenticator; the protected live-check case remains intentionally skipped without a same-instance runtime token | supply a same-instance least-privilege E2E identity and move destructive/repeatable owner tests to disposable PostgreSQL, then implement the remaining G2 projects |
| 9 — OpenID suite | Not started | stand up a local conformance harness after the G2 OIDC behavior is contract-complete |
| 10 — G2 slice | Active — preflight is wired through domain/service/API/UI and the worktree image; owner enrollment now has runtime-only bootstrap authorization, deterministic resume, verified WebAuthn public-material persistence, recovery confirmation, consumption and restart proof (`complete`, revision 3) | create/link the canonical owner identity, role and authenticated session; add explicit reset/lost-artifact handling and disposable restore/replay evidence before promotion |

The immediate engineering slice is now canonical owner activation: bind the
verified Nomen credential to the inherited identity core's human user,
instance-owner role and phishing-resistant session without introducing a
second identity truth. Add lost-artifact/reset review, repeat the complete
bootstrap project against disposable PostgreSQL, and supply a same-instance
least-privilege credential for the protected preflight case. Until those
journeys and a release-bound evidence manifest pass, `deployment_preflight` and
`owner_enrollment` remain disabled `preview` facts.

## Official external references

- OpenID certification process and conformance suite:
  <https://openid.net/certification/how-to-certify-your-implementation/> and
  <https://openid.net/certification/about-conformance-suite/>
- OpenID certification fees:
  <https://openid.net/certification/fees/>
- FIDO2 server certification and fees:
  <https://fidoalliance.org/certification/functional-certification/functional-certification-servers/>
  and <https://fidoalliance.org/fido-certification-fees/>
- OASIS SAML 2.0 conformance requirements:
  <https://docs.oasis-open.org/security/saml/v2.0/saml-conformance-2.0-os.pdf>
- OAuth 2.0 Security Best Current Practice, RFC 9700:
  <https://datatracker.ietf.org/doc/rfc9700/>
- NIST SP 800-63-4 Digital Identity Guidelines:
  <https://csrc.nist.gov/pubs/sp/800/63/4/final>
- OWASP ASVS 5.0:
  <https://owasp.org/www-project-application-security-verification-standard/>
- ISO/IEC 27001:2022 overview:
  <https://www.iso.org/standard/27001>
- NIST CMVP and FIPS claim rules:
  <https://csrc.nist.gov/Projects/cryptographic-module-validation-program/FAQs>
- Sigstore/Cosign signing and verification:
  <https://docs.sigstore.dev/cosign/signing/overview/> and
  <https://docs.sigstore.dev/cosign/verifying/verify/>
- SLSA v1.2 build requirements:
  <https://slsa.dev/spec/v1.2/build-requirements>
- Public TLS certificate automation:
  <https://letsencrypt.org/docs/>
- Inbound license and trademark disposition is recorded privately under L0;
  product docs do not republish third-party legal pages.
