# 21 — First operational IAM vertical slice

**Status:** accepted implementation contract; capability remains `preview`
until every acceptance row passes.

## Outcome

From a clean PostgreSQL database, a private operator can deploy Nomen, enroll
the first owner, create an organization and OIDC application, invite a member,
complete Authorization Code with PKCE, enforce an entitlement, inspect and
revoke the resulting session, and recover the deployment without another
product shell.

This slice establishes the implementation pattern for every later capability.
It is deliberately deeper than the dashboard and narrower than the complete
IAM release.

## Actors

- `deployment_operator` initializes and recovers the Nomen deployment but
  cannot silently impersonate an organization member;
- `organization_owner` configures the organization, application and invitation;
- `organization_member` authenticates and grants consent where required;
- `relying_party` uses a public OIDC client and Authorization Code with PKCE;
- `clyffy_operator` plans and applies only the same authorized application and
  invitation commands exposed to the owner.

## Required journey

1. Preflight reports database, issuer, TLS, notification and custody readiness
   without changing state.
2. Initialization creates a Nomen-owned schema and a single-use bootstrap
   ceremony without printing credential material.
3. The first owner enrolls a passkey using WebAuthn. Development may offer a
   clearly marked password bootstrap, but operational promotion requires the
   phishing-resistant factor and independent recovery proof.
4. The owner creates an organization and receives an explicit tenant context.
5. The owner creates an OIDC public application from a guided form. Nomen
   validates exact redirect URIs, selects Authorization Code, requires PKCE
   `S256`, and explains each choice.
6. The owner creates an invitation with role and expiry. The invitation is
   single-use, revocable and never stores a reusable bearer value in audit or
   analytics.
7. The member accepts the invitation, enrolls authentication and completes the
   application's OIDC flow.
8. Nomen issues asymmetrically signed tokens with issuer, audience, subject,
   tenant and authorized scopes. Unsupported `alg`, redirect mismatch, missing
   PKCE and code replay fail closed.
9. The relying party requests a protected route. No session returns typed
   `401`; a valid identity missing the entitlement returns typed `403`.
10. Owner and member can inspect applicable sessions and security events. The
    owner revokes the session and refresh-token family; subsequent use fails.
11. Every human action and the equivalent Clyffy action produces correlated,
    tenant-scoped audit and operation evidence.
12. Nomen restarts without state drift. A backup is restored into a disposable
    deployment and preserves issuer, application, identities, policy, audit and
    revocation outcome.

## Public resources

The browser and Clyffy use versioned Nomen management resources for:

- preflight and deployment state;
- owner enrollment and recovery state;
- organizations and memberships;
- applications and redirect URIs;
- invitations;
- sessions and revocation;
- audit search; and
- immutable operation plan/apply/verify.

Every JSON response crossing into the browser is parsed by an ArkType schema.
Every request is independently validated and authorized by the Go service.
The browser never supplies trusted tenant, actor, entitlement or assurance
context.

The exact preflight and bootstrap authority behavior is specified in
`25-deployment-preflight-contract.md`. That contract is changed before either
resource shape or ceremony behavior changes.

## Playwright acceptance projects

The suite runs against the production OCI image and disposable PostgreSQL:

| Project | Required evidence |
|---|---|
| `bootstrap-owner` | clean initialization, WebAuthn virtual authenticator, recovery export and resume after interruption |
| `guided-application` | terminology help, safe defaults, redirect validation, preview, confirmation and successful creation |
| `member-invitation` | expiry, single use, revocation, tenant ownership and successful acceptance |
| `oidc-pkce` | discovery, authorize, consent, code exchange, refresh rotation, logout and session revocation |
| `authorization-errors` | typed `401` versus typed `403`, missing scope and cross-tenant denial |
| `clyffy-equivalence` | human and AI paths invoke the same command and produce equivalent authorization and audit outcomes |
| `accessibility` | keyboard-only journey, focus order, accessible names, error association and reduced motion |
| `restart-restore` | runtime restart and destructive disposable restore preserve the stated outcomes |

Playwright traces, screenshots and videos are retained only for failed CI runs,
are scrubbed of credential fields, and are not product audit records.

## Promotion gate

The slice remains `preview` until all acceptance projects, PostgreSQL tenant
tests, protocol negative tests, Automaton consumer verification, secret scan,
image hardening checks and backup/restore drill pass for one release digest.
No navigation label, seeded fixture or manual login can override this gate.
