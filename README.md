# Tessera

**Product ID:** `tessera` · **Deployment:** standalone or managed · **Status:**
pre-release, contract-first

Tessera is a standalone identity and access management platform: users,
organizations, applications, authentication, federation, authorization,
sessions, recovery and the audit trail connecting them. It is built to run for
one independent customer, as a managed customer deployment, or behind an
optional product adapter.

A Roman **tessera hospitalis** was a small token broken in two — each party kept
a half, and fitting them back together proved who you were and what bond you
held. A *tessera* was also the tablet a sentry checked for the watchword. The
name is not a metaphor for a seat token; it is the same object, older.

Tessera is the identity authority, not a panel component. Its own web
application and APIs are complete product surfaces. A host product may embed
those surfaces later, but it cannot become a runtime dependency.

## What this owns

- Users, organizations, projects, applications and service identities.
- Sessions, authenticators, recovery, revocation and delegated administration.
- OIDC/OAuth, SAML, federation, directory and identity-aware access edges.
- Authorization policy, visual flows, blueprints and lifecycle automation.
- The standalone management UI, API, CLI, audit evidence and guided setup.
- Signing keys, JWKS publication and safe rotation.
- Optional integration profiles such as the Shippin seat token.

## What this does not own

- Billing, plans and pricing decisions.
- General infrastructure inventory and mesh networking.
- Secret custody; Tessera accepts references to a secret manager and does not
  turn its database into one.
- Conversation or application orchestration.

## Contract first, implementation second

The order is deliberate: **standalone product contract, standalone deployment,
managed operation, then integrations.** Tessera must be installable,
configurable, observable, upgradeable, recoverable and usable without a
Shippin service, route, account or token.

The standalone contract is `docs/02-standalone-product-contract.md`. The seat
token remains a tested optional integration profile for Automaton and Shippin;
it does not define Tessera's product limits.

## Reading order

| doc | what it settles |
|---|---|
| `docs/00-charter.md` | ownership and dependency direction |
| `docs/02-standalone-product-contract.md` | the release contract every deployment must satisfy |
| `docs/03-roadmap.md` | the evidence-driven worklist to reach that release |
| `docs/01-seat-token-contract.md` | the optional Shippin token profile |
| `docs/05-minting-a-seat-token.md` | how that optional profile is exercised |

House rules: `../AGENTS.md` first, then `AGENTS.md` here. Tessera product law
lives in this repository. Host-product documents govern only their adapters.
