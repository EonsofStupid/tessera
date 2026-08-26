# AGENTS.md — Nomen

Read `../AGENTS.md` first (workspace house rules), then this.

## What this project is

Nomen is a standalone identity and access management product for independent
self-hosted and managed-customer deployments. It must install, bootstrap,
operate, upgrade, recover and expose its complete management experience without
Shippin being installed or reachable.

Contract first: the product boundary is
`docs/02-standalone-product-contract.md`. The Shippin seat-token profile in
`docs/01-seat-token-contract.md` is one optional consumer integration, not the
Nomen product boundary.

## Rules specific to here

- **Never put a live secret in this repository.** No private keys, client
  secrets, tokens, cookies, OAuth sessions or application databases — the
  workspace rule (`../../CLYFFY.md`), and it bites hardest in exactly this
  project because everything here looks like it wants one. Fixtures are
  generated at test time, never checked in.
- **Absorb zero third-party product licenses.** Nomen is licensed only by
  AngryVibes LLC and shippin.ai (`LICENSE`, `NOTICE`, `docs/23`). Do not add
  AGPL, Apache, MIT, or BSD product `LICENSE` files to runtime trees. Do not
  classify Nomen as ZITADEL or Authentik licensed. Dependency licenses stay
  in the SBOM, not as Nomen's product license. Waiver text stays out of Git.
- **Change the contract before the code.** A consumer is already verifying
  against `docs/01-seat-token-contract.md`. Editing behaviour without editing
  the contract is how two implementations start disagreeing quietly.
- **Asymmetric signatures only.** If a change would let a consumer verify with
  a shared secret, it is wrong — see the `HS*` note in the token contract.
- **A missing entitlement is a `403` with a typed body**, never a bare `401`.
  Not-signed-in and not-entitled have different fixes.
- **Product version is `1.0.0-alpha` and frozen.** Do not invent 1.1, v2
  product APIs, or another name until Nomen Vault, Nomen Mesh, and the first
  IAM slice are built. `backend/v1` / `backend/v3` / gRPC `v2` are internal
  layout, not releases.
- **Public vs enterprise** is `docs/26-editions-and-demo-tier.md`. One binary.
  Default `NOMEN_EDITION=public`. Redis, Vault, and Mesh adapters are withheld
  on public. The hosted demo is public plus `NOMEN_DEMO_CAPS`.
- **Project hooks** in `.grok/hooks/` (and `.claude/settings.json`) are the
  runtime events that can deny a tool or block a stop. Package-manager scripts
  cannot. They stay inert until `/hooks-trust` (or `--trust`). They fail open
  on crash; CI remains the merge gate.

## Boundaries

Nomen owns its standalone web application, management API, protocol surfaces,
identity lifecycle, federation, policy and flow configuration, audit evidence,
deployment lifecycle and operator guidance. It does not own billing, pricing,
general infrastructure inventory, mesh networking, secret custody or
conversation. Optional adapters may cite decisions made by those systems but
Nomen must remain useful when every adapter is absent.

A standalone product feature must not depend on a Shippin route, account,
token, adapter or shell. Shippin-specific pricing or entitlement behavior
belongs in `../shippin`; a generic authorization capability belongs here.

## Consumers to keep working

- **Nomen standalone UI and conformance harness** — the first consumers of
  Nomen's management and protocol contracts. They may not use Shippin-only
  assumptions.
- **Automaton** — `../automaton/engine/serve/identity.mjs` (Stage 2) verifies
  the optional seat-token profile and gates routes on scopes.
- DevForge and the Shippin panel follow through adapters after the standalone
  release gate passes.
