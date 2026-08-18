# 06 — Blueprints: the plan

**Status:** built. Every step below landed as its own commit, each proven by
its done-when before the next began; this document is now the design record.

## What a blueprint is here

A YAML file that declares what should be true, applied in one transaction,
rolled back whole on any failure, and a no-op when re-applied. Identity
configuration stops being something somebody clicked and becomes something
somebody reviewed.

The model is Authentik's (`upstream/authentik/blueprints/`), reimplemented over
our own domain — the flow-engine treatment from the roadmap, applied to
configuration. Nothing here shells out to Authentik or parses its schema.

```yaml
schema: tessera.blueprint.v1
name: dev — the probe seat

entries:
  - model: tessera/seat
    id: probe                     # local handle, for refs from later entries
    identifiers:
      member: "386740…"
    attrs:
      occupant: agent
      basis: subscription
      account: "386740…"
      workspaces: [ws-0001, ws-0002]
      scopes: [hosting:active, terminal:advanced, chat:unified]
      policy_version: pol_2026_08_17
```

## Decisions, and why they are not preferences

**Applied in file order; refs reach backward only.** No topological sort. A
sort hides authoring mistakes and makes apply order change when an unrelated
entry is edited — with a sort, the same file can succeed on Tuesday and
deadlock on Wednesday. A forward reference is an error naming *both* entries,
which is a review comment, not a runtime surprise.

**References are strings, not YAML tags.** `${keyof:probe}` in an attr value,
resolved by the engine, instead of Authentik's `!KeyOf` custom tag. Two
reasons. Custom tags need `gopkg.in/yaml.v3` unmarshalers on every entry type;
`sigs.k8s.io/yaml` (already a direct dependency) parses YAML *as JSON*, so a
string convention keeps the whole model JSON-compatible — and the panel will
*generate* blueprints, which means emitting JSON and validating against a JSON
Schema. A format only writable by a YAML library is a format the panel fights.

**Unknown attrs are errors, not ignored.** A typo'd `workspace:` (singular)
that is silently dropped is a blueprint that lies — it applies clean and grants
nothing. Strict decoding (`DisallowUnknownFields`) makes the typo a validate
failure naming the field.

**`unchanged` is detected by read-and-compare, not claimed.** Each applier
reads current state and skips the write when equal. That is what makes "apply
twice, second run reports all-unchanged" a real assertion rather than an
UPSERT that happens to not error, and it keeps `updated_at` honest.

**One advisory lock per apply.** `pg_advisory_xact_lock` on a constant, inside
the transaction. Two applies interleaving — the panel and an operator, or two
panel workers — is otherwise a race where both succeed and one silently wins.
The lock serialises them for the cost of one line.

**States are `present` and `absent`.** `absent` of something already gone is
success — a desired state reached twice is reached. Authentik's `must_created`
guard is deferred until something needs it; a state nobody uses is a state
nobody tests.

**Instance is an apply-time argument, not file content.** The same blueprint
provisions any instance; which instance is the operator's (later, the panel's)
decision at apply time. Files with instance ids baked in cannot be promoted
from staging to production without editing, and edited-on-promote files are not
the reviewed ones.

## Where it lives

| piece | path | layer rule it obeys |
|---|---|---|
| `Blueprint`, `Entry`, `DesiredState`, validation | `backend/v1/domain/blueprint.go` | rules with no IO |
| `Applier` port, registry, apply engine | `backend/v1/domain/blueprint_apply.go` | domain defines the port |
| seat applier | `backend/v1/storage/seat/applier.go` | storage implements it |
| YAML loader (dir → domain types) | `backend/v1/storage/blueprint/` | filesystem is an adapter too |
| `tessera blueprint validate\|apply` | `cmd/blueprint/` | surface translates, decides nothing |
| the dev state, as files | `blueprints/dev/` | config is reviewed, not clicked |

The engine's applier port takes a `database.Transaction` (which is a
`QueryExecutor`); the seat repository grows tx-scoped variants of its writes so
the engine's transaction is the only one — `SetSeat` keeps its own wrapper for
callers that arrive without one.

## Sequence

Each step is one commit, each with its own "done when".

**3.2a — model and validation.** Domain types and `Validate`: schema version,
known model, known state, duplicate local ids, ref syntax, refs reaching
backward. *Done when* every refusal is a table-driven test and every ref error
names both entries.

**3.2b — engine over fakes.** Registry, engine, one transaction via
`Beginner`, ref substitution, per-entry outcomes
(`created|updated|unchanged|removed`). Tested with a fake applier and a fake
transaction that records commit/rollback. *Done when* idempotence, ordering,
and failure-at-entry-N-rolls-back are proven without a database.

**3.3a — the seat applier.** Repository refactor to `QueryExecutor` helpers;
attr → `domain.Seat` mapping with strict decoding; read-compare for
`unchanged`. *Done when* the OIDC path is untouched and `dev/seat-probe.sh` is
still green.

**3.3b — atomicity on real Postgres.** Embedded-PG test (the harness
`migrations_test.go` already uses): apply with a deliberately broken last
entry, prove the database identical to before; apply twice, prove the second
run all-unchanged. *Fallback* if the embedded binary cannot download in this
sandbox: the same assertions in `dev/blueprint-probe.sh` against the dev
cluster on 5433. *Done when* the broken-final-entry run leaves no trace.

**3.4a — loader and command.** Directory → blueprints in lexicographic order,
`.yaml`/`.yml`; `tessera blueprint validate` (parse + resolve, no writes) and
`apply --dir --instance`. `blueprints/dev/seats.yaml` enters the repo;
`seat-probe.sh` step 2 becomes a blueprint apply and asserts the second apply
reports no changes. *Done when* `rm -rf .pgdata` → `up.sh` → `blueprint apply`
→ probe all green, and re-apply changes nothing.

**3.4b — apply on start.** `Blueprints.Dir` in config (empty = off), applied
in `start.go` immediately after our migration. *Done when* the probe needs no
explicit apply step — a fresh database reaches known state from files on boot,
which is the roadmap's Phase 3 exit criterion.

## Out of scope, on purpose

Conditions, `!Env`/`!File`/`!Format` tags, OCI-fetched blueprints, a
blueprint-instances status table, hot-reload, and multi-instance fan-out (the
CLI takes one `--instance`; iterating a fleet is the panel's job, one apply per
tenant, each atomic on its own).

## Preflight, instead of a risk list

A risk you can check for is not a risk, it is a missing check.
`dev/preflight.sh` proves the environment before anything half-runs: Go 1.25+,
postgres server binaries, ports 5433/8088 free or ours, writable state dirs,
and the embedded-postgres PG 16 cache. `dev/up.sh` runs the quick subset before
touching anything.

The rule for failures: every ✗ prints the exact command that fixes it — `sudo`
spelled out where root is genuinely required, never run by the script itself.
Whether a person or an AI is driving, the output is the runbook.

The one formerly-listed risk — embedded-postgres downloading its binary on
first use — is now the preflight's last check, with three remediations by
cause: `--fetch-embedded` for a one-time download, `HTTPS_PROXY=…` behind a
proxy, and copying `~/.embedded-postgres-go` from any machine that has it when
air-gapped. On this box the cache is already populated, so 3.3b's test runs
with no network at all. Until a blocked environment is fixed,
`dev/blueprint-probe.sh` covers the same assertions against the dev cluster.

What remains an actual trade-off: **read-compare cost** is one SELECT per
entry. Fine at panel scale (entries per tenant are tens, not thousands);
revisit only if a blueprint ever grows past that.
