# 02 — Provenance and licensing

**Status:** operating rule. Applies before any code, schema or blueprint moves
into this repository.

## The distinction that matters

Not reselling widens what may be **studied and adapted**. It does not remove a
licence obligation, and it does not make copying and reimplementing the same
act.

| | what it is | what it needs |
|---|---|---|
| **Study** | read the architecture, understand the model, write our own | attribution in a design note; no licence obligation |
| **Adapt** | take a schema, blueprint, migration or protocol detail and rework it | check the licence; record where it came from |
| **Derive** | copy source, port it line by line, vendor a module | the upstream licence terms apply in full, including notice and attribution requirements |

The workspace already has the rule: *"Do not move code between projects without
an explicit absorption/split plan"* and *"Preserve immutable IDs and upstream
provenance."* `PRODUCT-ARCHITECTURE.md` §1 adds that upstream notices, licences
and security advisories stay visible to operators. This document is where that
gets recorded for identity.

**Verify the licence of each upstream at the moment you take something.** Do not
take a version from memory — including mine. Read the `LICENSE` in the revision
you are actually looking at, and note the revision.

## What is worth taking from each

Recorded as design intent, not as a claim that anything has been taken.

### Zitadel — the architecture

- **Organizations as a first-class primitive.** Tenancy in the data model rather
  than a column on a user. This is the one that matters at fifty customers and
  is painful to retrofit at five hundred.
- **Event-sourced core.** Every grant, revocation and membership change is an
  event, so "who allowed this, when, and what took it back" is answerable by
  construction instead of by having remembered to log it. That lands directly on
  §8's reliability law and the seam's `audit/{correlation_id}`.
- **Projections.** Read models rebuilt from the event log, which is also how a
  bad migration becomes recoverable rather than terminal.

### Authentik — the operations

- **Blueprints: identity configuration as declarative data.** This is how fifty
  workspaces get identical identity config instead of fifty hand-built ones, and
  it is the piece that makes the fleet story work.
- **Forward-auth outposts.** A proxy that authenticates in front of a service
  the service does not have to change. Worth keeping available as a deployment
  shape in front of workspaces regardless of who mints tokens.
- **The application/provider split** — one identity, many relying parties, each
  with its own audience. Which is exactly what `aud` does in
  `01-seat-token-contract.md`.

### What is deliberately not taken from either

Their token *shapes*. `shippin.seat-token.v1` follows
`SHARED-SEAM-DRAFT.md` because the seam is the thing every Shippin client
already shares. Adopting a vendor's claim names would make the vendor visible in
every consumer, which is the coupling this whole approach exists to avoid.

## Recording an intake

Every time something is adapted or derived, append a row. An empty table is the
honest state today.

| date | upstream | revision | licence | what was taken | where it landed |
|---|---|---|---|---|---|
| — | — | — | — | — | — |
