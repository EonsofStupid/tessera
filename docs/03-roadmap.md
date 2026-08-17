# 03 — Roadmap: Zitadel's architecture, Authentik's infrastructure

**Status:** proposed sequence. Phase 0 is in flight; nothing past it is committed.

Two intakes, and they are different in kind. From **Zitadel**, a *data and audit
model* — ideas, written from scratch. From **Authentik**, *running
infrastructure* — a service that already works, configured properly. Keeping
those two straight is what makes this legal and cheap; confusing them is how a
weekend becomes a licensing problem.

## The licence line, checked at source

**Zitadel is AGPL-3.0-only.** It moved off Apache 2.0 at v3. AGPL §13 is a
network clause: run a modified version that users reach over a network and you
owe them the source. "We are not reselling it" does not avoid this — the trigger
is *network interaction by users*, not sale, and Shippin workspaces are exactly
network interaction by users.

So: **study Zitadel, copy nothing from its core.**

There is one precise exception worth knowing, because it is the useful part.
`LICENSING.md` carves out directories:

| path | licence | usable? |
|---|---|---|
| repository default | **AGPL-3.0-only** | architecture study only |
| `proto/` | **Apache 2.0** | yes — API definitions, with attribution |
| `apps/docs/` | Apache 2.0 | yes |
| `apps/login/`, `packages/zitadel-client/`, `packages/zitadel-proto/` | **MIT** | yes |

Their *API shape* is Apache 2.0 while their *server* is AGPL. If anything is
ever taken from Zitadel, it is a protobuf definition — never an implementation.

**Authentik core is MIT**, with enterprise features under a separate licence.
Blueprints, config and deployment shape can be adapted with attribution. Check
which features are enterprise before depending on one.

Every intake gets a row in `02-provenance-and-licensing.md`, with the revision
actually read. That table is still empty and should stay empty until something
is genuinely taken.

## Where Authentik actually stands today

From `shippin-mesh/authentik/docker-compose.yml`, not from a wish:

- `ghcr.io/goauthentik/server:2025.10`, server + worker, Postgres 16, Dragonfly
  for Redis, published on 9000/9443.
- **No `/blueprints` volume is mounted.** All configuration is therefore made by
  hand in the UI, and exists only in that database. This is the gap Phase 1
  closes and the single highest-value thing to take from Authentik.
- **The worker runs `user: root` with `/var/run/docker.sock` mounted.** That is
  how it manages outpost containers, and it is also host root by another name: a
  compromise of the worker is a compromise of the host. Phase 2 decides whether
  to keep it.

## Phase 0 — run it behind the contract *(in flight)*

Authentik is the identity provider. Tessera owns
`01-seat-token-contract.md`; Automaton's Stage 2 verifies against it.

- An OIDC provider in Authentik whose property mappings emit
  `shippin.seat-token.v1` claims — `authorization.scopes`, `account_id`,
  `workspace_id`, and an `aud` naming exactly one workspace.
- Automaton verifies it with `engine/serve/identity.mjs` — generic OIDC
  discovery and JWKS, no vendor knowledge.

**Done when** a token minted by the real Authentik is accepted by Automaton, and
one minted for another workspace is refused on the audience check.

## Phase 1 — configuration becomes data *(Authentik: blueprints)*

Blueprints are Authentik's infrastructure-as-code: YAML applied on a 60-minute
loop, discovered from `/blueprints`, an OCI registry or the database, and each
applied "within a single **atomic database transaction**" that rolls the whole
blueprint back if any entry fails.

That transactional, repeatedly-applied shape is what turns identity config from
something clicked once into something reviewed, diffed and reproduced. It is the
piece that makes fifty workspaces identical rather than fifty hand-built ones.

- Mount `/blueprints`; move today's hand-made config into `tessera/blueprints/`.
- Provider, scope mappings, flows and groups become reviewed files.
- A fresh Authentik on an empty database reaches known state unattended.

**Done when** the running instance can be destroyed and rebuilt from the
repository, and Phase 0's test still passes against it.

## Phase 2 — harden the deployment *(Authentik: outposts, and the socket)*

An outpost is "a single deployment of an authentik component… that can be
deployed anywhere that allows for a connection to the authentik API",
configured over websockets and needing no internet access.

- **Decide on the docker socket.** Either accept it and treat that host as
  identity-critical with nothing else on it, or deploy outposts declaratively
  and drop the mount. It should be a decision with a date, not an inheritance.
- Evaluate a **forward-auth proxy outpost** in front of workspaces as defence in
  depth — it authenticates in front of a service that does not have to change,
  which is useful independently of who mints tokens.

**Done when** the identity host's blast radius is written down and matches how
it is actually deployed.

## Phase 3 — adopt the tenancy model *(Zitadel: architecture, no code)*

Zitadel's hierarchy is **Instance → Organization → Project**. An Instance gives
"full data isolation and independent settings"; an Organization "is the vessel
where your projects and users live… comparable to a tenant in a SaaS system".
**Project Grants** let one organization delegate role assignment to another,
which is the B2B case: a partner manages their own users against your roles.

Tenancy in the data model rather than a column on a user is the thing that is
painful to retrofit at five hundred customers and nearly free to decide now.

- Map it: Shippin **account** ≙ Organization, **workspace** ≙ the audience a
  token is scoped to, **member** ≙ user. Write the mapping into
  `01-seat-token-contract.md` so `account_id` has a defined meaning.
- Decide whether multi-org membership is a v1 claim. It is cheap as a claim now
  and a migration later.

**Done when** the contract states the tenancy model and Automaton's entitlement
gate reads it.

## Phase 4 — event-sourced authorization *(Zitadel: architecture, no code)*

Zitadel stores changes "as events in an Event Store… a ledger that gets new
entries over time", with **projections** as "computed objects, that will be used
on the query side", giving "a built-in audit trail that tracks all changes over
an unlimited period of time".

That is the property §8's reliability law wants and the seam's
`audit/{correlation_id}` assumes: *who granted this, when, and what revoked it*
becomes answerable by construction rather than by having remembered to log it.

**Start narrow.** Not all of identity — only the **authorization aggregate**:
grants, revocations, scope changes, policy versions. It is the lowest-volume
part of the system and the part where "we cannot reconstruct what happened" is
least acceptable. Sessions and credentials stay wherever the provider keeps
them.

**Done when** a grant and its revocation can be replayed from events, and a
projection rebuilt from scratch matches the live read model.

## Phase 5 — the swap, if it is still wanted

Nothing above requires replacing Authentik, and that is the design. If a
Tessera-native provider is ever built, it emits the same tokens against the same
JWKS discovery, and consumers change nothing. In Automaton it is one flag
(`--auth-strict`), not a migration.

The honest cost of going native, so it is a real decision when it arrives: token
lifecycle, key rotation, session revocation, MFA, and account recovery when a
customer has lost their second factor. Recovery is where identity systems
actually fail.

## What this roadmap will not do

- Copy Zitadel server code. AGPL, network clause, and Shippin is a network
  service. Architecture only.
- Fork Authentik. Adapt configuration, not the codebase — a fork owns every CVE
  in it forever.
- Adopt either project's token claims. `shippin.seat-token.v1` follows the
  Shippin seam so no vendor is visible to a consumer.
- Run two identity authorities at once, which `CONFLICTS-DECISIONS.md` already
  warns against. Authentik is the one that runs; Zitadel is read, not deployed.

## Sources

- Zitadel eventstore — https://zitadel.com/docs/concepts/eventstore/overview
- Zitadel organizations — https://zitadel.com/docs/concepts/structure/organizations
- Zitadel licensing — https://github.com/zitadel/zitadel/blob/main/LICENSING.md
- authentik blueprints — https://docs.goauthentik.io/customize/blueprints/
- authentik outposts — https://docs.goauthentik.io/add-secure-apps/outposts/
- The deployed stack — `shippin-mesh/authentik/docker-compose.yml`
