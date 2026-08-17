# AGENTS.md — Shippin Identity

Read `../AGENTS.md` first (workspace house rules), then this.

## What this project is

Identity and authorization for the Shippin umbrella. Contract first: the
product boundary is `docs/01-seat-token-contract.md`, and the identity provider
behind it is an implementation detail that must stay swappable.

## Rules specific to here

- **Never put a live secret in this repository.** No private keys, client
  secrets, tokens, cookies, OAuth sessions or application databases — the
  workspace rule (`../../CLYFFY.md`), and it bites hardest in exactly this
  project because everything here looks like it wants one. Fixtures are
  generated at test time, never checked in.
- **Change the contract before the code.** A consumer is already verifying
  against `docs/01-seat-token-contract.md`. Editing behaviour without editing
  the contract is how two implementations start disagreeing quietly.
- **Record provenance at the moment of intake**, in
  `docs/02-provenance-and-licensing.md`. Studying an upstream is free; adapting
  or deriving from one is a row in that table and a licence check against the
  revision you actually read.
- **Asymmetric signatures only.** If a change would let a consumer verify with
  a shared secret, it is wrong — see the `HS*` note in the token contract.
- **A missing entitlement is a `403` with a typed body**, never a bare `401`.
  Not-signed-in and not-entitled have different fixes.

## Boundaries

This project does not own billing, plans, infrastructure inventory or
conversation. It expresses entitlement as scopes that cite decisions made
elsewhere. When a change here starts needing a pricing rule, the change belongs
in `../shippin`.

## Consumers to keep working

- **Automaton** — `../automaton/engine/serve/identity.mjs` (Stage 2) verifies
  seat tokens and gates routes on scopes. It is the first consumer and the
  fastest way to find out whether a contract change is real.
- DevForge and the Shippin panel follow.
