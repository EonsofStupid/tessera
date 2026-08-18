# 00 — Charter

## One sentence

**Tessera** answers "who is this, and what may they do" once, for every service
in the umbrella, and hands the answer over as a short-lived signed token that
anyone can verify without calling it.

## The name

A *tessera hospitalis* was a token broken in two, each party keeping a half;
fitting the halves together proved identity and the bond between them. A
*tessera* was also the tablet carrying a watchword for a sentry to check.

Chosen over the obvious alternative for one reason worth writing down:
"connection" is already the most loaded word in this product — §7 is *Guided
connection*, Mesh Layer 2 is connectivity, MCP connectors ship — and all of that
weight sits on networking. A name from that family would read as the mesh
module to anyone arriving cold, including us in a year.

## Why it is its own project

ADR 06 puts identity and authorization in the Shippin Control Plane and requires
enforcement *at data boundaries*. That only works if there is one contract to
enforce — otherwise every consumer integrates a vendor SDK, and the vendor
quietly becomes the architecture.

`PROJECTS_OVERVIEW.md` rule 3 says to amend the product contract before adding a
second implementation authority. That amendment is
`shippin/docs/platform/PRODUCT-ARCHITECTURE.md` §6, which now names this module;
this repository is what it names.

## Boundaries

**Owns**

- Member, account and organization identity; sessions and their revocation.
- Entitlement scopes and the policy version that produced them.
- Seat tokens, the JWKS, and key rotation.
- The guided sign-in surfaces that belong to identity rather than to a product.
- Federation in both directions: upstream identity providers Tessera trusts and
  downstream applications for which Tessera is the identity provider.
- A provider-neutral management API for directory, sessions, federation,
  authentication flows, trust and audit. The Shippin panel projects that API;
  Tessera does not ship a second customer shell.

**Does not own**

- Billing, plans or pricing. Entitlement is *expressed* here as scopes and
  *decided* in the product domain — a scope cites a decision, it does not make
  one.
- Infrastructure inventory, workspace lifecycle, provider capability facts.
- Mesh topology, peer lifecycle and installation. Zuul owns those facts and
  uses Tessera for human, agent and installer identity.
- Conversation and orchestration. Automaton is a consumer.
- Live secrets of any kind under `W:` — the workspace rule, unchanged.

## The strategy, in one line

**Contract first; the provider behind it is swappable.**

Run a real identity provider now rather than writing one, keep every consumer
behind `01-seat-token-contract.md`, and replace the provider later if that still
makes sense. Then it is a swap, not a migration — and the decision your own
`CONFLICTS-DECISIONS.md` records as open stays genuinely open instead of being
settled by whatever got integrated first.

The surface is larger than it looks: token lifecycle, key rotation, session
revocation, MFA, and the one everybody underestimates — account recovery when a
customer has lost their second factor. Recovery is where identity systems
actually fail, and it is worth designing before anything else is built.


## The control surface

Tessera appears as a first-class module inside the Shippin member shell. A
customer clicks **Tessera** and the persistent Shippin chrome changes context
to identity management without navigating into a vendor console or a second
product. The module has three lenses over one identity system:

- **Infrastructure** — identity attachments to Shippin runtimes, services and
  Zuul enrollment. These are projections of other owners' inventory, never a
  second infrastructure database.
- **AI** — agent seats, service identities, delegated `act` chains, scopes and
  recent authentication activity.
- **Customers** — people, organizations, workspace membership, sessions, MFA,
  recovery and linked external identities.

Federation is the center of the module rather than an advanced settings page.
It distinguishes upstream providers Tessera trusts from downstream clients
that trust Tessera, and makes mappings, login policy, health and audit visible
without ever returning a client secret to the browser.

The complete panel and management contract is
`08-control-surface-and-federation.md`.

## Current state

Phases 1–4 are implemented: Tessera builds from its own source tree, mints the
seat-token contract Automaton verifies, converges declared identity state from
blueprints, and executes password, TOTP and recovery-code configurations on
one flow engine. The next product slice is the management/federation contract
and its Shippin panel projection; outposts and full account recovery follow it.
