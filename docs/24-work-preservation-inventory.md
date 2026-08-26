# 24 — Work preservation and branch harvest inventory

**Status:** active non-destructive inventory.
**Observed:** 2026-08-21 at `main` commit `a60746e`.
**Rule:** no branch is deleted, force-updated or bulk-merged from this record.

## Current worktree

The primary worktree contains uncommitted standalone UI, landing-page,
Playwright, contract and generated-asset work. It is user work and remains
untouched except for the additive planning/governance documents explicitly
created during this program.

Before implementation harvesting begins, the current modifications and
untracked files must be preserved in a named commit or operator-approved patch
bundle. Generated asset replacement pairs are reviewed together; deleted hash
assets are not restored independently from their new build outputs.

## Branch topology

Eighteen named `agent/*` branches exist locally. Every branch is also attached
to a worktree except `agent/phase-5-control-surface`. The two standalone
branches are already contained in current `main`; the other branches diverged
before the standalone product/runtime work and cannot be treated as safe
fast-forwards.

| Branch | Relationship to current `main` | Harvest classification |
|---|---|---|
| `agent/nomen-standalone-runtime` | same commit as `main` | contained; retain as historical integration point |
| `agent/nomen-standalone-product` | ancestor of `main` | contained; retain as product-boundary checkpoint |
| `agent/nomen-provenance-inventory` | diverged | high-priority path harvest: provenance generator, manifest, BOM and CI concepts; regenerate against current history |
| `agent/nomen-operational-plan` | diverged | documentation reference only; older Shippin-first assumptions are superseded by standalone contracts and program 22 |
| `agent/nomen-ldap-outbound` | diverged composite | capability candidate: extract LDAP contract/client/tests and relevant provenance path-by-path after G2 pattern exists |
| `agent/nomen-ldap-lifecycle` | diverged composite | capability candidate layered on outbound LDAP; harvest only after outbound design review |
| `agent/nomen-vaultix-reference` | diverged composite | optional-adapter reference; defer until standalone security gate, extract value-blind secret-reference work only after review |
| `agent/nomen-runnable-iam` | diverged composite | implementation/reference inventory; compare pinned container and conformance work with current runtime, never bulk merge |
| `agent/nomen-gap-foundation` | diverged composite | aggregation checkpoint; use as an index to earlier domain/API/provenance work, not a merge source |
| `agent/nomen-private-cloud` | diverged composite | product-language and workspace reference; standalone boundary supersedes private-cloud coupling |
| `agent/nomen-capability-discovery` | diverged early slice | current `main` contains later capability discovery; inspect only for tests/contracts missing from current implementation |
| `agent/nomen-container` | diverged early slice | current runtime has later container work; compare hardening/config tests only |
| `agent/nomen-deployment-lifecycle` | diverged early slice | current `main` contains lifecycle domain; compare exhaustive transition tests only |
| `agent/nomen-management-errors` | diverged early slice | current `main` contains typed management errors; compare fixtures and negative cases only |
| `agent/nomen-operation-contract` | diverged early slice | current `main` contains operation grammar; compare idempotency/conflict tests only |
| `agent/nomen-overview-api` | diverged early slice | current `main` contains overview projection; compare authorization/projection tests only |
| `agent/phase-5-control-surface` | diverged documentation | harvest terminology and guided-outcome research only where compatible with standalone UI |
| `agent/nomen-shippin-shell-contract` | diverged documentation | optional future adapter reference; explicitly outside the standalone release gate |

Commit-count comparisons show many unmatched historical commits, but those
counts are not a delivery measure: later standalone commits may reimplement or
supersede equivalent behavior with different patches. Harvesting is based on
current contract compatibility and per-path behavioral evidence.

## Harvest procedure

For each branch, create a read-only harvest record before applying anything:

1. identify its exact source commit and merge base;
2. list unique contracts, implementation paths, migrations, fixtures and tests;
3. compare each path with current `main` and current standalone terminology;
4. classify it `already_absorbed`, `superseded`, `candidate`, `blocked_legal`,
   `blocked_dependency` or `reject_with_reason`;
5. record origin/license/permission evidence for candidate code;
6. port the smallest coherent capability patch onto current `main`;
7. change the current contract before changed behavior;
8. run the complete affected evidence gate; and
9. retain the old branch until the new release manifest proves absorption.

Do not cherry-pick aggregation merges or generated-code floods. Do not resolve a
conflict by selecting an entire side. Database migrations, generated protocol
code, built web assets and provenance manifests receive dedicated review.

## Harvest order

1. Preserve the primary worktree.
2. Regenerate source provenance from `agent/nomen-provenance-inventory`.
3. Compare early domain/API branches against current `main` for missing tests.
4. Complete the first operational vertical slice on current `main`.
5. Harvest outbound LDAP and lifecycle work as the first Nomen-class edge
   only after the vertical-slice evidence pattern exists.
6. Defer Vaultix, Shippin shell and private-cloud coupling until their optional
   program gates.

## Deletion policy

No worktree or branch is removed merely because it is old, diverged or partly
superseded. Removal requires:

- a completed harvest record;
- confirmation that every candidate path is absorbed or rejected with reason;
- an immutable release/source-provenance reference where applicable; and
- explicit product-owner approval for the exact branch/worktree names.

