# 09 — Guided identity experience

**Status:** accepted production experience contract. A non-mutating prototype
may validate composition, but it is not a completion milestone.
**Depends on:** `08-control-surface-and-federation.md`.

## Product rule

Tessera asks for the customer's outcome before it asks for an identity
primitive. A customer should be able to protect a team, application, agent or
private service without first learning OIDC, SAML, SCIM, PKCE, device grants,
flows, stages, projects or outposts.

Those primitives remain visible in an optional review for operators who need
them. Hiding vocabulary must never mean hiding consequences: before a write,
the panel shows who will gain access, what will trust Tessera, what Tessera
will trust, which workspace is affected and how the change can be undone.

## First-run outcomes

The first control in the Tessera panel is **What do you want to set up?** It
offers six stable jobs:

| customer outcome | plain-language questions | provider-neutral result |
|---|---|---|
| Invite my team | who, which workspace, what can they use? | people, membership, human seats and a sign-in flow |
| Use company sign-in | which company and verified domain? | upstream provider, domain routing and account-link policy |
| Connect an app | browser, server or device; who uses it? | downstream client, redirect/origin policy, grants, audience and scopes |
| Add an AI agent | acts alone or for a person; which workspace? | agent seat, service identity, delegation policy and narrow scopes |
| Set up private access | laptop, server or both; which workspace? | one-time Tessera enrollment handed to Zuul; mesh state remains in Zuul |
| Secure or recover access | passkey/MFA, recovery or active-session concern? | authenticators, recovery flow or session revocation plan |

The panel recommends the safest compatible posture. Protocol choice is only
shown when it changes an external requirement or the customer opens advanced
details.

## Guided sequence

Every setup follows the same five-part grammar:

1. **Choose an outcome.** Start with the customer's desired result.
2. **Answer only discriminating questions.** Ask for information that changes
   the resulting resources or policy. Use existing Shippin account and
   workspace context rather than asking twice.
3. **Recommend a plan.** Explain it in terms of people, applications and
   recovery; name technical protocols secondarily.
4. **Review effects.** Show the trust direction, affected workspace, created or
   changed resources, permissions and rollback before applying.
5. **Apply and verify.** Use a server-side adapter, record audit provenance and
   finish with an outcome test such as a discovery check, test sign-in or Zuul
   enrollment check.

A prototype may implement steps one through four against typed fixtures only
when it is explicitly labeled preview. Production work does not stop there:
the same components must consume live provider-neutral projections, submit the
shared P1.2 operation contract, resume durable progress and verify the outcome.
Fixture interactions must never imply that live identity state changed.

## Progressive disclosure

The default layer uses these customer terms:

| say first | reveal under “What Tessera will create” |
|---|---|
| company sign-in | upstream IdP, OIDC/SAML/LDAP source, domain routing |
| connected app | downstream relying party, client, grants, redirect URIs |
| sign-in steps | flow, stage, authenticator and login policy |
| AI agent | service identity, agent seat, `act` delegation and scopes |
| private access | device authorization or PKCE enrollment handed to Zuul |
| active access | session, factor evidence, token and revocation |

Health and errors follow the same rule: lead with the consequence and remedy,
then expose a safe diagnostic. Secret values are never repeated in review,
fixtures, exports, logs or browser responses.

## API provenance and normalized resources

The guided layer is a composition surface, not a second identity model. It
normalizes capabilities proven in both inherited APIs:

| Tessera resource | Zitadel capability provenance | Authentik capability provenance |
|---|---|---|
| people and teams | users, organizations, memberships | users, groups and group membership |
| human sign-in | login policy, factors, sessions | flows, stages and authenticators |
| company sign-in | generic OIDC, SAML and LDAP identity providers | OAuth/OIDC, SAML, LDAP and social sources |
| connected apps | projects; OIDC, SAML and API applications | applications and OAuth/OIDC, SAML, proxy, LDAP and RADIUS providers |
| AI and service identities | machine users, keys, grants and memberships | service accounts, tokens and application access |
| active access and evidence | sessions, factor checks and events | sessions, tokens and events |
| portable access edge | provider adapters | outposts, endpoints, agents and enrollment |
| desired state | evented commands and projections | blueprint-driven configuration |

The browser never calls these inherited vendor-shaped administration APIs.
The Shippin server-side adapter submits one provider-neutral intent to Tessera,
receives a deterministic plan, then applies that plan through Tessera's command
layer.

```text
POST /tessera/v1/guides/plan
POST /tessera/v1/guides/{plan_id}/apply
GET  /tessera/v1/guides/{plan_id}
POST /tessera/v1/guides/{plan_id}/verify
```

`plan` is non-mutating and contains no credential values. `apply` requires the
narrow write permissions for every affected resource, is idempotent against
the plan's desired-state revision and returns a typed conflict when the source
state changed after review. A missing permission remains a typed `403`.

## Done when

- A new customer can reach a correct recommended plan from any of the six
  outcomes without identity-domain vocabulary.
- The production path plans, applies, resumes and verifies through Tessera; a
  polished fixture-only or vendor-console-backed screen does not satisfy this
  contract.
- The review states the trust direction, affected workspace, access granted,
  recovery posture and rollback before a mutation.
- An expert can inspect the exact provider-neutral resources and standards
  without leaving the guide.
- A successful apply is followed by an outcome verification, not merely a
  green API response.
- Any prototype is visibly non-mutating; the production path is server-side,
  authorized, audited, secret-safe and idempotent.
