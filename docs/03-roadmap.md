# 03 — Roadmap: building Tessera from Zitadel and Authentik

**Status:** build plan. Both trees are cloned under `upstream/`.

Tessera is **built from** these two codebases, not operated alongside them.
Nothing here deploys Authentik or Zitadel as a service; both are source we take
from.

## The trunk is Zitadel

`upstream/zitadel` — Go 1.25, `github.com/zitadel/zitadel`. It is the trunk
because the architecture is the part worth having and because Go matches the
rest of the estate (GoClyffy, and Authentik's own outposts).

What we are here for, in `internal/`:

| path | what it gives us |
|---|---|
| `eventstore/` | the ledger: aggregates, events, push/query, `read_model.go`, the `handler/` projection machinery |
| `command/` | the write side — how an intent becomes events |
| `query/` | the read side, built from projections |
| `org/`, `project/`, `iam/` | Instance → Organization → Project, the tenancy model |
| `api/` | gRPC + REST surface, and `authz/` behind it |
| `crypto/`, `id/`, `migration/` | key handling, id generation, schema migration |
| `idp/` | upstream identity provider federation |

`proto/` is where the API shape lives and is separately Apache-2.0 licensed.

## The infrastructure is Authentik's — and most of it is Go too

`upstream/authentik` — Go 1.26 module `goauthentik.io` **plus** a Django core.
The split matters, and it is better than it looks: the parts we want are mostly
Go and portable directly.

| path | language | what it gives us |
|---|---|---|
| `internal/outpost/proxyv2/` | **Go** | forward-auth proxy, session store, handlers — 150 Go files under `internal/` |
| `internal/outpost/ldap/`, `radius/`, `rac/` | **Go** | protocol frontends for things that will never speak OIDC |
| `blueprints/` + `schema.json` | **YAML/JSON** | declarative config as data — language-neutral, lift as-is |
| `authentik/flows/` | Python | the flow/stage *model* (`planner.py`, `challenge.py`, `markers.py`) — ported, not copied |

So: Go core from Zitadel, Go outposts from Authentik, YAML blueprint schema
lifted whole, and one Python subsystem (the flow engine) that gets reimplemented
in Go because it is a model rather than a library.

## Phases

### Phase 1 — the trunk stands up

Fork Zitadel's core into `tessera/`, strip what we will never ship, and get it
building and running against Postgres under our own module path.

- Rename the module; keep `internal/eventstore`, `command`, `query`, `org`,
  `project`, `iam`, `crypto`, `id`, `migration`.
- Drop: their console UI, their marketing site, their SaaS-specific billing and
  onboarding paths, i18n we do not need.
- **Done when** it builds, migrates a fresh Postgres, and creates an
  organization through the API.

### Phase 2 — it mints `shippin.seat-token.v1`

The contract already exists and Automaton already verifies against it
(`engine/serve/identity.mjs`). This is the first thing that makes Tessera real.

- An OIDC provider path that emits our claims: `authorization.scopes`,
  `account_id`, `workspace_id`, and `aud` naming exactly one workspace.
- JWKS published; discovery at `/.well-known/openid-configuration`.
- **Done when** Automaton accepts a Tessera-minted token and refuses one minted
  for a different workspace. That test exists and currently runs against a fake
  issuer; point it at this.

### Phase 3 — blueprints

Take Authentik's blueprint model — YAML applied on a loop inside one atomic
transaction, rolled back whole on any failure — and implement it over Tessera's
command layer. The schema is JSON Schema and lifts directly.

This is what makes fifty workspaces identical rather than fifty hand-built ones:
identity config becomes reviewed files instead of something somebody clicked.

- **Done when** a fresh database reaches known state from `blueprints/` alone,
  and re-applying is a no-op.

### Phase 4 — flows

Port Authentik's flow/stage engine: a flow is an ordered set of stages, a
planner decides which apply to this request, and a stage returns a challenge the
client renders. Identity, MFA, recovery and consent all become the same object.

`planner.py`, `challenge.py` and `markers.py` are the model to read; the
implementation is ours, in Go, over the eventstore.

- **Done when** password + TOTP + recovery are three configurations of one
  engine rather than three code paths.

### Phase 5 — outposts

Bring `internal/outpost/proxyv2` across as the forward-auth proxy. It already
speaks OIDC to a core and configures itself over websockets; it needs pointing
at ours.

Gives workspaces authenticated ingress without each service implementing
anything, and gives LDAP/RADIUS to things that will never speak OIDC.

- **Done when** a workspace sits behind the proxy and Automaton still
  authenticates through it unchanged.

### Phase 6 — recovery

Left last deliberately, and it is the one that decides whether this is real.
Account recovery when a customer has lost their second factor is where identity
systems actually fail, and it is a flow like any other once Phase 4 lands.

## Provenance

`upstream/` is reference and is gitignored — we do not vendor two whole trees
into this repo. Code that comes across gets a row in
`02-provenance-and-licensing.md` naming the file, the upstream revision and the
licence, because a rename in six months should still be traceable to where it
came from.

Authentik is MIT: keep the notice. Zitadel is AGPL-3.0: if people other than us
reach it over a network, they can ask for the source of our modified version.
Both are recorded once in `02` and are not re-litigated here.
