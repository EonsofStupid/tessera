# 03 — Roadmap: Tessera identity core and portable edges

**Status:** build plan. Both trees are cloned under `upstream/`.

Tessera incorporates two design references; neither is operated alongside it.
The shipped product, process names, APIs and operator experience are Tessera.

## The trunk is Tessera's compatibility core

The evented Go compatibility core is the trunk because the architecture is the
part worth preserving and Go matches the rest of the estate.

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


## The portable identity edges come from Authentik — and most are Go

`upstream/authentik` — Go 1.26 module `goauthentik.io` **plus** a Django core.
The split matters, and it is better than it looks: the parts we want are mostly
Go and portable directly.

| path | language | what it gives us |
|---|---|---|
| `internal/outpost/proxyv2/` | **Go** | forward-auth proxy, session store, handlers — 150 Go files under `internal/` |
| `internal/outpost/ldap/`, `radius/`, `rac/` | **Go** | protocol frontends for things that will never speak OIDC |
| `blueprints/` + `schema.json` | **YAML/JSON** | declarative config as data — language-neutral, lift as-is |
| `authentik/flows/` | Python | the flow/stage *model* (`planner.py`, `challenge.py`, `markers.py`) — ported, not copied |

So: evented Go core, portable Go identity edges, YAML blueprint
schema lifted whole, and one Python subsystem (the flow engine) that gets
reimplemented in Go because it is a model rather than a library. These are
identity capabilities, not Tessera's secrets infrastructure. Vaultix is the
separate Infisical-derived product that owns secrets, certificates and
privileged-access custody and integrates with Tessera through protected
references.

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

Done. Seat facts left compatibility metadata for `tessera.seats`, in Tessera's own
schema with its own migration series starting at one — separate from the
compatibility schema next door, whose 001–018 belong to the imported migration series and whose next number
would collide with ours on any sync.

The shape is a scale decision rather than a storage chore. Workspaces are a
child table because the relation is read in both directions and only one of them
is the token path: `seat → workspaces` on every mint, and `workspace → seats`
for the panel, which an array column cannot answer without scanning the
instance. Scopes stay an array because they are only ever read *with* the seat.

`tessera seat set|show|list` is the operator's way in, and drives the same
repository blueprints will. **Done when** the mint path reads from the table and
no seat fact remains in metadata — `dev/seat-probe.sh` now fails if one does.

One trap worth the ink: the compatibility core runs its schema migration as a *setup step*,
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
attaches to is still owned by the compatibility core, created by setup or the management API — "known
state from files alone" covers users only when a user applier joins the
registry, which is the natural next model.

### Phase 4 — flows ✅

Done, and the done-when is literal on this box: `blueprints/dev/flows.yaml`
declares login-password, login-mfa and recovery as three YAML entries, one
engine executes all three, and `dev/flow-probe.sh` drives login-password over
HTTP as a customer would — start, identify, wrong password (fails closed,
stays, re-asks with a field error), right password, done. The resulting
session carries `session.user.checked` and `session.password.checked` in the
eventstore — the same factor events the session v2 API writes, because stages
delegate to the same SessionCommands and never verify anything themselves.
The wrong attempt lands as `user.human.password.check.failed` on the user
aggregate, where lockout counts it.

The engine's own rules, each with a test: a wrong answer re-asks and moves
nothing; infrastructure failure is never a field error; every identify
failure answers with one vague sentence (account enumeration is harvested on
login pages); unknown execution and wrong token answer identically; the
session token crosses the wire exactly once because completion consumes the
execution. The executor's in-process calls run as TESSERA_FLOWS with exactly
session.read + session.write — what a login client holds, not SYSTEM_OWNER.

Design and the planner seam (where Authentik's policies land when needed):
`docs/07-flows.md`.

- **Done when** password + TOTP + recovery are three configurations of one
  engine rather than three code paths. ✓

### Phase 5 — control surface and federation

Make Tessera a first-class module inside the persistent Shippin member shell,
not a vendor console beside it. The first slice is a provider-neutral overview
API and `/member/tessera`: three identity lenses (infrastructure attachments,
AI/agent identities and customers), upstream providers and downstream clients,
sign-in posture, sessions, trust and audit.

Federation is the product center. Tessera both trusts upstream OIDC/OAuth2,
SAML and LDAP providers and acts as the OIDC/SAML identity provider for
Shippin, Zuul, Automaton, DevForge and customer applications. Zuul uses Tessera
for PKCE/device or guided one-time enrollment, then owns the mesh it creates;
Tessera never becomes a peer or infrastructure inventory.

The control-surface contract and vertical slice are
`docs/08-control-surface-and-federation.md`; the outcome-first guide and its
normalized API mapping are `docs/09-guided-identity-experience.md`.

### Operationalization program

Phase 5 becomes a maintainable private-community product through
`docs/11-operational-blueprint.md`. That document defines the lifecycle,
reconciliation, delegation, recovery and release quality bar. The complete
dependency-ordered delivery queue is `docs/12-execution-worklist.md`; its first
four slices deliberately establish provenance and operation contracts, then
read-only truth, repeatable installation and one real guided federation path
before the team fans out.

The service ships in this phase as one cloud-neutral, OCI-runtime-compatible
container image. Its database, ingress and secret injection remain deployment
concerns; the exact runtime and Podman boundary is
`docs/10-container-runtime.md`.

- **Done when** clicking Tessera changes the Shippin shell into a coherent
  identity dashboard, upstream/downstream trust are visibly distinct, and a
  Zuul installer can authenticate and enroll one workspace without an embedded
  reusable secret.

### Phase 6 — outposts

Bring `internal/outpost/proxyv2` across as the forward-auth proxy. It already
speaks OIDC to a core and configures itself over websockets; it needs pointing
at ours.

Gives workspaces authenticated ingress without each service implementing
anything, and gives LDAP/RADIUS to things that will never speak OIDC.

- **Done when** a workspace sits behind the proxy and Automaton still
  authenticates through it unchanged.

This phase is not complete when code merely compiles. Inbound and outbound
LDAP, forward-auth, identity-aware proxy behavior and their isolated failure
modes must pass the conformance proofs in
`18-identity-edge-and-vaultix-contract.md`.

### Phase 6.1 — visual flow control surface

The execution engine in Phase 4 is the runtime foundation, not the finished
operator experience. Add a visual graph editor over the same versioned flow
schema: templates, typed stage configuration, validation, simulation, diff,
publish, rollback and audit. YAML and the visual editor round-trip through one
canonical model so they cannot become competing sources of truth.

- **Done when** a non-specialist can customize a login and recovery journey,
  prove every reachable outcome in simulation, review the exact diff, publish
  a revision and roll it back without shell or YAML work.

### Phase 7 — recovery

Left last deliberately, and it is the one that decides whether this is real.
Account recovery when a customer has lost their second factor is where identity
systems actually fail, and it is a flow like any other once Phase 4 lands.

## Provenance

`upstream/` stays as reference and is gitignored. The code we take lives here,
under our module path.


## The two backends

`backend/v3` is the imported compatibility architecture. It is not
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
