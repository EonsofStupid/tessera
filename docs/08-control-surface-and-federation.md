# 08 — Control surface and federation contract

**Status:** accepted build contract; implementation follows this document.
**Product surface:** Tessera inside the persistent Shippin member shell.
**Service id:** `shippin.tessera`.

## The promise

Clicking **Tessera** in the Shippin panel opens a complete identity control
surface. The Shippin shell stays mounted and changes module context; customers
do not enter a repainted Zitadel console, an Authentik deployment, or a second
navigation system.

Tessera folds in the durable qualities of both sources:

- Zitadel's evented identity core, organization/project model, sessions,
  standards support and provider federation;
- Authentik's declarative blueprints, guided flow/stage model and portable
  outposts;
- Shippin's seat-token boundary, human/agent distinction, workspace audience
  and typed entitlement failures.

They are source and design provenance, not customer-visible dependencies. The
provider behind the contract remains swappable.

## One shell, three identity lenses

The panel route is `/member/tessera`. Selecting it updates the persistent shell
with the Tessera module name, its navigation and an identity-scoped account
control. The shell remains Shippin-owned; Tessera owns the data and actions.

### Shell module contract

Tessera is a first-class Shippin module, not a standalone application embedded
after login. Shippin publishes a versioned `shippin.tessera` module descriptor
from its service catalog. The descriptor supplies the mount path, route ids,
sublinks, required capabilities and member permissions; it does not download
arbitrary administration code from the identity service.

The initial route tree is:

| route id | member path | shell label |
|---|---|---|
| `tessera.start` | `/member/tessera` | Start |
| `tessera.directory` | `/member/tessera/directory` | Directory |
| `tessera.access` | `/member/tessera/access` | People & access |
| `tessera.infrastructure` | `/member/tessera/infrastructure` | Infrastructure |
| `tessera.ai` | `/member/tessera/ai` | AI |
| `tessera.customers` | `/member/tessera/customers` | Customers |
| `tessera.federation.upstreams` | `/member/tessera/advanced/upstreams` | Identity providers |
| `tessera.federation.clients` | `/member/tessera/advanced/clients` | Connected applications |
| `tessera.flows` | `/member/tessera/advanced/flows` | Sign-in flows |
| `tessera.trust` | `/member/tessera/advanced/trust` | Trust & keys |
| `tessera.audit` | `/member/tessera/advanced/audit` | Audit |

Shippin keeps its header, account/workspace context, global navigation and
member session mounted while the content outlet changes. Tessera owns the
module title, local subnavigation, breadcrumbs and content beneath that
outlet. An iframe, full-page vendor-console redirect, second login shell or
independently drifting navigation tree violates this contract.

Every sublink names a capability-discovery id and optional permission. The
server-side adapter resolves those facts before returning the member module
descriptor. Unsupported routes are absent, unavailable routes are disabled
with a typed remedy, and neither the browser nor a cached descriptor promotes
a route that Tessera did not advertise.

Direct URLs, refresh, forward/back history and copied deep links resolve to the
same shell and selected sublink. A route-level error boundary contains Tessera
failures inside the content outlet: the Shippin shell and account controls
remain usable, and the content renders loading, empty, degraded, unauthorized,
not-entitled or unavailable states from typed responses instead of a blank
page.

### Reliability and quality bar

Visual polish cannot substitute for identity-system reliability. A route is
production-ready only when its provider-neutral contract, authorization,
loading/empty/degraded behavior, audit result and recovery path are tested.
The inherited ZITADEL and authentik suites remain valuable implementation
evidence, but Shippin/Tessera conformance tests prove the product boundary.

For every released module version:

- service-catalog validation rejects duplicate paths, unknown parents and
  routes without capability gates;
- browser tests cover selection, direct entry, refresh, history, expired
  member authentication, typed entitlement denial and dependency loss;
- adapter tests prove timeouts, bounded retries, cancellation and redaction;
- contract tests run against fixtures and a live Tessera deployment without
  changing the component model;
- upgrade tests prove the oldest supported Shippin descriptor against the
  newest supported Tessera API and the reverse supported combination;
- route availability and operation success are measured separately, so a
  rendered dashboard cannot hide broken sign-in, federation or recovery.

A capability stays preview, disabled or degraded until those checks pass. It
does not become production-ready because the equivalent inherited vendor page
works.

| lens | shows | authority |
|---|---|---|
| **Infrastructure** | identity attachments for workspaces, services, runtimes and Zuul installers; issuer/client health | Tessera for identity; Shippin/Zuul for referenced inventory |
| **AI** | agent seats, service identities, delegation chains, scopes, session activity | Tessera |
| **Customers** | people, organizations, workspace membership, MFA, recovery posture and external links | Tessera |

Infrastructure is deliberately an identity lens, not an inventory screen.
Tessera may say “this Zuul installer authenticated as this workspace service
identity”; it may not decide where the peer runs or whether a machine exists.

## Dashboard information architecture

The first viewport starts with the outcome-oriented guide defined in
`09-guided-identity-experience.md`. A customer chooses what they want to set up
before seeing identity-domain vocabulary. Alongside that guide, the overview
answers four questions without requiring a settings hunt:

1. **Can people and agents sign in?** Issuer, JWKS, flow and federation health.
2. **Who has access?** Human seats, agent seats and workspace bindings.
3. **What trusts what?** Upstream providers and downstream clients, separated.
4. **What just changed?** Session, federation, key and policy audit activity.

The module navigation leads with customer jobs while retaining an advanced
operator layer:

| area | purpose |
|---|---|
| Start | guided setup and readiness |
| Directory | humans, agents, organizations, workspace bindings and service identities |
| People & access | directory, sign-in posture, recovery and active sessions |
| Advanced | upstream/downstream federation, flows, keys, policy and immutable audit |

## Federation is two directional

The UI and API must never flatten “IdP” into one ambiguous list.

### Upstream — Tessera trusts them

Organizations may authenticate through OIDC/OAuth2, SAML, LDAP and the
supported social/enterprise templates already present in the inherited core.
Every provider exposes:

- protocol and issuer/metadata identity;
- organization scope and enabled login flows;
- account creation/linking policy;
- domain and attribute/role mappings;
- last successful discovery or metadata validation;
- a typed readiness state and a safe diagnostic;
- audit history for configuration and mapping changes.

Secrets are accepted only through a protected write path and are never
returned by reads, logs, browser fixtures or blueprint exports. A read may say
that a credential is configured and when it was rotated, never what it is.

### Downstream — they trust Tessera

Shippin, Zuul, Automaton, DevForge and future customer applications are
relying parties. Every client exposes:

- stable client id and display name;
- protocol, redirect origins and allowed grant types;
- workspace/audience policy and allowed scopes;
- human, agent or both occupant classes;
- public-client/PKCE versus confidential-client posture;
- last authentication and readiness;
- rotation status without returning credentials.

OIDC discovery and JWKS remain the default integration. SAML service-provider
support is a standards adapter; it does not change Tessera's internal model.

## Zuul boundary

Zuul is the dynamically configured installer and private-mesh product. A
customer can install it, authenticate, and receive a working mesh without
learning network control-plane details. The identity handshake is Tessera's:

1. A browser or desktop client uses authorization code + PKCE.
2. A headless installer uses device authorization or a one-time guided
   enrollment flow; it never embeds a reusable shared secret.
3. Tessera binds the authenticated human or service identity to one Shippin
   account and workspace and issues only the enrollment scopes Zuul needs.
4. Zuul creates/configures peers, routes and policy in its own domain.
5. Audit links the Zuul enrollment id to Tessera's subject, actor, workspace,
   client and policy revision without copying mesh inventory into Tessera.

Dex remains a development adapter only. Production Zuul federates to Tessera;
Zuul does not become another account database.

## Management API boundary

The Shippin web app talks to a Shippin server-side adapter, which calls the
Tessera management API. Browsers do not receive operator credentials or call
inherited vendor-shaped administration endpoints directly.

The provider-neutral resource groups are:

```text
GET  /tessera/v1/overview
GET  /tessera/v1/directory/subjects
GET  /tessera/v1/directory/seats
GET  /tessera/v1/federation/upstreams
GET  /tessera/v1/federation/clients
GET  /tessera/v1/flows
GET  /tessera/v1/sessions
GET  /tessera/v1/trust/keys
GET  /tessera/v1/audit
```

Mutations live under the same resources and must be idempotent where a stable
desired state exists. Every mutation records subject, actor (including `act`),
account, workspace when applicable, policy revision and result.

Every response declares provenance:

- `G` for protocol/capability definitions;
- `T` for organization providers and policy;
- `W` for workspace seats, clients and attachments;
- `M` for a live instance check;
- `P` for dashboard summaries and joined views.

Projection responses name their source revision/check time. A projection never
silently promotes a missing live fact into healthy.

## Authorization and failure contract

Management permissions are namespaced and narrow:

```text
tessera:overview:read
tessera:directory:read       tessera:directory:write
tessera:federation:read      tessera:federation:write
tessera:flows:read           tessera:flows:write
tessera:sessions:read        tessera:sessions:revoke
tessera:trust:read           tessera:trust:rotate
tessera:audit:read
zuul:mesh:enroll
```

Not signed in is `401`. Signed in but missing one of these permissions is `403`
with the existing typed body naming the missing permission. The panel renders
the different remedies and never turns a `403` into a login loop.

## First vertical slice

The first implementation is intentionally read-first:

1. publish `shippin.tessera` in the Shippin service catalog with link relations
   for the member route and management adapter;
2. implement `GET /tessera/v1/overview` as a provider-neutral projection;
3. add `/member/tessera` and a Tessera-aware shell context;
4. render the six-outcome non-mutating setup guide, overview, the three
   identity lenses, upstream/downstream trust and
   recent activity from a typed fixture behind an integration boundary;
5. replace each fixture lane with the real management projection without
   changing domain components;
6. add federation writes only after read masking, typed `403`s, audit and
   protected-secret transport are proven.

## Done when

- Selecting Tessera in the Shippin panel changes the persistent shell and opens
  a coherent identity dashboard without exposing a vendor console.
- The service catalog, server-side adapter and browser agree on one versioned
  route tree; deep links and browser history preserve shell and sublink state.
- Capability and permission changes hide or disable the correct sublinks from
  server facts without a frontend deployment or a guessed version rule.
- Loss of Tessera data or one route never unmounts the Shippin shell, traps the
  member in a spinner or breaks unrelated Shippin modules.
- Humans, agents and infrastructure attachments are distinguishable views over
  one authority, not three databases.
- Upstream IdPs and downstream clients are visibly different and safely
  manageable.
- A Zuul installer authenticates through Tessera and enrolls one workspace
  without a reusable embedded secret.
- Automaton continues verifying the unchanged seat-token contract.
- No browser response, fixture, log, blueprint or repository file contains a
  client secret, private key, session credential or enrollment token.
