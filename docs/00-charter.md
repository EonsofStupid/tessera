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

**Does not own**

- Billing, plans or pricing. Entitlement is *expressed* here as scopes and
  *decided* in the product domain — a scope cites a decision, it does not make
  one.
- Infrastructure inventory, workspace lifecycle, provider capability facts.
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


## Current state

Nothing is implemented. What exists is the contract, and one consumer
(Automaton, Stage 2) building a verifier against it — which is the fastest way
to find out whether the contract is real.
