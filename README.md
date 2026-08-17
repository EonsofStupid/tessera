# Tessera

**Workspace ID:** `shippin.tessera` · **Umbrella:** Shippin · **Status:** new,
contract-first

Identity and authorization for the Shippin umbrella: who a member is, which
account they belong to, what they are entitled to, and the short-lived tokens
that carry all three to every other service.

A Roman **tessera hospitalis** was a small token broken in two — each party kept
a half, and fitting them back together proved who you were and what bond you
held. A *tessera* was also the tablet a sentry checked for the watchword. The
name is not a metaphor for a seat token; it is the same object, older.

It exists because ADR 06 puts identity and authorization in the Shippin Control
Plane and asks that it be enforced *at data boundaries* — which means every
consumer needs one contract to enforce against, not one integration per vendor.

## What this owns

- Member, account and organization identity.
- Sessions, and their revocation.
- Entitlement scopes, and the policy version that produced them.
- **Seat tokens** — the short-lived, audience-scoped JWTs every other service
  verifies. `docs/01-seat-token-contract.md` is the load-bearing document here.
- The JWKS every consumer fetches, and the key rotation behind it.

## What this does not own

- Billing and plans. Entitlement is *expressed* here and *decided* by the
  product domain; a scope in a token cites a decision made elsewhere.
- Infrastructure inventory, workspace lifecycle, provider facts.
- Conversation, intent, orchestration — that is Automaton, which is a consumer.

## Contract first, implementation second

The order is deliberate and is the whole strategy: **the token contract is the
product boundary, and the identity provider behind it is an implementation
detail.** Get the contract right and the provider is swappable — run Authentik
today, run something Shippin-native later, and no consumer changes a line.

Get it backwards and every consumer learns one vendor's quirks, which is the
migration nobody wants to do twice.

Automaton is the first consumer and is already building the verifier
(`engine/serve/identity.mjs`, Stage 2). A verifier is the most honest way to
specify a token: it either accepts yours or it does not.

## Reading order

| doc | what it settles |
|---|---|
| `docs/00-charter.md` | boundaries, and what "identity" means here |
| `docs/01-seat-token-contract.md` | the claims every consumer verifies |
| `docs/02-provenance-and-licensing.md` | what may be studied, what may be copied |
| `docs/03-roadmap.md` | the sequence: Authentik infrastructure, Zitadel architecture |
| `docs/04-building-the-trunk.md` | how the Zitadel source actually builds |

House rules: `../AGENTS.md` first, then `AGENTS.md` here. Product law starts at
`../shippin/docs/platform/PRODUCT-ARCHITECTURE.md`.
