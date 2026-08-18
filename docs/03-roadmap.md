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

### Phase 1 — the code is ours ✅

Done. `internal/`, `pkg/`, `cmd/`, `backend/`, `proto/` and `main.go` are in
this repository under `github.com/EonsofStupid/tessera`, and it builds.

Still to strip, when it stops being useful rather than on principle: the console
UI, the login-v1 assets, the SaaS onboarding and billing paths.

### Phase 2 — it mints `shippin.seat-token.v1` ✅

Done. Tessera mints seat tokens and Automaton's own verifier accepts them.

- `backend/v1/domain` holds the claim set and the rules, with no OIDC import —
  so "`unknown` is never promoted" and "`aud` names exactly one workspace" are
  unit tests rather than integration hopes.
- `internal/api/oidc/seat_claims.go` gathers the facts and stamps them in
  `createJWT`, the one call before the signature.
- A workspace reaches `aud` through `urn:shippin:audience:<entry>`, and the
  member's stored entitlement decides whether they may have it. See the token
  contract, *How one is asked for*.
- **Done:** `docs/05-minting-a-seat-token.md` walks the whole path, and
  Automaton accepts a token for `ws-0001`, refuses one minted for `ws-0002`,
  refuses a tampered payload and refuses `alg: none`.

Still Phase 3's to fix: seat facts live in user metadata, which is where they
are *stored*, not where they should be *authored*. Blueprints are what will
write them.

### Phase 3.1 — seats get a table ✅

Done. Seat facts left Zitadel user metadata for `tessera.seats`, in Tessera's own
schema with its own migration series starting at one — separate from the
`zitadel` schema next door, whose 001–018 belong upstream and whose next number
would collide with ours on any sync.

The shape is a scale decision rather than a storage chore. Workspaces are a
child table because the relation is read in both directions and only one of them
is the token path: `seat → workspaces` on every mint, and `workspace → seats`
for the panel, which an array column cannot answer without scanning the
instance. Scopes stay an array because they are only ever read *with* the seat.

`tessera seat set|show|list` is the operator's way in, and drives the same
repository blueprints will. **Done when** the mint path reads from the table and
no seat fact remains in metadata — `dev/seat-probe.sh` now fails if one does.

One trap worth the ink: Zitadel runs its own schema migration as a *setup step*,
and a setup step is recorded once and skipped forever. Ours runs on every start
instead. Had it been a step, migration 002 would never reach a database that had
already been set up, every deployment in the fleet would silently keep the old
schema, and the failure would surface as a missing column a long way from here.

### Phase 3 — blueprints ✅ (seats; more models plug into the same engine)

Take Authentik's blueprint model — YAML applied on a loop inside one atomic
transaction, rolled back whole on any failure — and implement it over Tessera's
command layer. The schema is JSON Schema and lifts directly.

This is what makes fifty workspaces identical rather than fifty hand-built ones:
identity config becomes reviewed files instead of something somebody clicked.

- **Done when** a fresh database reaches known state from `blueprints/` alone,
  and re-applying is a no-op.

Done, for everything Tessera owns: `tessera blueprint validate|apply`, one
transaction per file, advisory-locked per instance, `Blueprints.Dir` applied to
every instance on every start — a deleted seat comes back on the next boot as
`1 created`, and the boot after that reports `converged: nothing to change`.
The atomicity proof runs against an embedded real PostgreSQL 16: a blueprint
whose last entry violates a real constraint leaves the database byte-identical,
timestamps included. Design and traps: `docs/06-blueprints.md`.

The honest boundary: `blueprints/` declares Tessera's own state (seats today;
each new model is one applier registered in `cmd/blueprint`). The *user* a seat
attaches to is still Zitadel's, created by setup or the management API — "known
state from files alone" covers users only when a user applier joins the
registry, which is the natural next model.

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

`upstream/` stays as reference and is gitignored. The code we take lives here,
under our module path.


## The two backends

`backend/v3` is Zitadel's third backend architecture, inherited whole. It is not
dead weight and it is not a parallel tree — it is already the substrate under the
legacy one: `internal/` imports its logging (67 files), its database and SQL
dialect layers (67), its domain objects (32) and its repositories (31).

`backend/v1` is ours, and it starts at one because it is the first architecture
*this* project has rather than the third somebody else had. The number is the
whole point: it says which of the two a file belongs to without anybody having
to remember.

New Tessera-owned code goes in `v1`, in v3's shape, because pre-release is
exactly when structure is cheap and there is no legacy to keep working:

| layer | holds | seat |
|---|---|---|
| `v1/domain` | entities, rules, and the ports they need as interfaces | `Seat`, `Seat.Token`, `SeatRepository` |
| `v1/storage` | adapters implementing those ports | `storage/seat` over user metadata |
| the API surface | nothing but translation | `internal/api/oidc/seat_claims.go` |

The rule that matters is which way the arrows point. `v1/domain` imports no
storage and no OIDC; storage imports the domain; the API layer imports both and
decides nothing. So "may this member occupy this workspace" is answered in one
method on one type, and every mint path — auth code, client credentials, jwt
profile, token exchange — reaches the same answer rather than four adapters each
remembering to ask.

Seat tokens went first because they are the first thing Tessera owns outright
rather than inherits: nothing upstream has a concept of
`shippin.seat-token.v1`.

What has *not* moved, and why: `SeatAudienceScope` stays in `internal/domain`
next to `ProjectIDScope`, because it is not domain vocabulary — it is OIDC scope
plumbing that has to sit beside the scope machinery it extends, and it uses that
package's unexported audience append.
