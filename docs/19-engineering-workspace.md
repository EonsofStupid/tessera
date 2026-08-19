# 19 — Tessera engineering workspace

**Status:** accepted development contract for the capability-gap program.
**Builds on:** `12-execution-worklist.md` and
`18-identity-edge-and-vaultix-contract.md`.

## Purpose

The workspace must make the safe path the easy path. A contributor can create
one isolated worktree, initialize generated-only local credentials, select a
named conformance profile, and know exactly which contracts and evidence apply.
No capability is developed as an untracked experiment and no local convenience
becomes a production secret path.

The canonical entry point is:

```text
go run ./dev/workspace init
go run ./dev/workspace doctor --profile core
go run ./dev/workspace sync-references
go run ./dev/workspace doctor --profile workspace
```

The machine-readable source of truth is `dev/workspace/manifest.json` and its
schema is `dev/workspace/manifest.schema.json`.

## Worktree and pull-request topology

Capability work is stacked on the accepted private-cloud foundation instead of
mixed into `main` or the existing container/control-surface branch.

```text
agent/tessera-private-cloud
└── agent/tessera-gap-foundation
    ├── secret-reference + Vaultix adapter contract
    ├── canonical visual-flow graph contract
    ├── LDAP outbound connector
    ├── LDAP inbound edge
    └── forward-auth / identity-aware proxy edge
```

The foundation branch contains workspace mechanics and shared contracts only.
Each implementation slice gets its own child branch and draft PR with one
capability proof target. A child may not mark an unrelated capability
`available`.

## Authority and source roots

| project | authority | workspace relationship |
|---|---|---|
| Tessera | identity, authorization, federation, journeys and identity-edge policy | writable project under development |
| Vaultix | secrets, certificates and privileged-access custody | read-only contract dependency; protected values never enter Tessera state |
| Shippin | persistent shell, tenant/product context and commercial policy | read-only consumer contract; UI implementation lives there |
| Zuul | installer and private-mesh lifecycle | read-only enrollment consumer contract |

Sibling projects are resolved from the shared `projects/` directory even when
Tessera runs inside a nested Git worktree. CI profiles that do not check out
sibling repositories use only the `core` profile.

Design-study source is treated separately from sibling product contracts. The
workspace manifest pins each reference to a full commit, installs only declared
sparse paths beneath ignored `upstream/`, and refuses a dirty or drifted
checkout. `sync-references` never overlays Tessera source and never silently
discards a contributor's changes. The pinned Authentik reference must also
match `provenance/source-manifest.json`.

## Code ownership map

New implementation follows these boundaries:

| path | purpose |
|---|---|
| `backend/v1/domain/` | provider-neutral entities, policy and ports for secret references, connectors, edges and journey revisions |
| `backend/v1/storage/` | Tessera persistence adapters; never a Vaultix value cache |
| `internal/api/` | protocol translation only; no policy decisions |
| `internal/edge/ldap/` | inbound LDAP protocol adapter over Tessera authority |
| `internal/edge/proxy/` | forward-auth and identity-aware proxy adapters |
| `conformance/` | black-box capability proofs and test profiles |
| `dev/workspace/` | workspace initialization, diagnostics and machine-readable topology |
| `../shippin/apps/web/` | visual editor and guided shell projection over Tessera contracts |

The visual editor, blueprint loader, management API and runtime executor share
one canonical graph schema owned by Tessera. Shippin renders that schema; it
does not create a second flow model.

## Named validation profiles

| profile | runs on | proves |
|---|---|---|
| `core` | Linux or Windows | contracts, domain rules, schemas, provenance, product language and generated-secret policy |
| `workspace` | Linux operator workstation | local tools, all four project boundaries and pinned design-study references |
| `linux-integration` | `tessera-linux` | PostgreSQL, OpenLDAP reference profile, proxy reference application and Vaultix contract adapter |
| `windows-ad` | `tessera-windows` | supported Active Directory profile, nested groups, disabled users, certificate trust and failover behavior |
| `release` | both runners | core plus every capability-specific conformance result bound to the exact signed bundle digest |

An available feature requires its own proof on the applicable profile. Linux
OpenLDAP success cannot substitute for the Windows Active Directory profile,
and proxy success cannot promote LDAP.

## Local state and secret policy

All mutable state is under `.artifacts/workspace/`, which is already ignored:

```text
.artifacts/workspace/
├── secrets/       # mode 0700; generated values, individual files 0600
├── evidence/      # raw local conformance evidence, never committed
├── logs/          # redacted runtime logs
└── state/         # disposable process/container state
```

`dev/workspace init` generates the development Tessera master key with the OS
cryptographic random source and never prints it. The dev loop consumes the file
through `--masterkeyFile`; neither YAML nor a process argument contains the
value. Connector passwords, Vaultix workload credentials, TLS private keys,
tokens and cookies follow the same generated/write-only rule.

Committed fixtures contain identities, schemas and expected outcomes only.
Tests create protected values at runtime and erase their temporary directory on
completion. A fixture value that can authenticate outside its own disposable
test process is a live secret and is forbidden.

## Ports and isolation

The manifest allocates loopback-only development ports. Each integration
profile uses a private Podman network and named, disposable state beneath the
workspace state root. Release proofs use a fresh network and state directory;
they do not reuse the operator's development database or the live Vaultix
instance.

The live `vaultix.shippin.cloud` service is never an automatic test target.
Tests use a contract fake first and a dedicated Vaultix sandbox only when an
operator explicitly supplies its endpoint and workload identity through the
protected environment.

## Evidence lifecycle

Every conformance run writes a result containing:

- capability and conformance ids;
- source and installed bundle revisions;
- runner/profile and dependency versions;
- start/end time and pass/fail result;
- redacted assertion summary;
- SHA-256 digest of the immutable raw evidence.

Raw evidence stays in the protected artifact store. Only the proof envelope and
digest may enter capability discovery. Failed, stale, unsigned or mismatched-
bundle evidence cannot enable a UI capability.

## First implementation sequence

1. **Workspace foundation:** doctor/init, generated credentials, manifest and
   schema, core CI on both runners.
2. **Vaultix reference boundary:** provider-neutral `SecretReference`, workload
   authentication port, fake adapter, negative leak tests, then sandbox wire
   verification.
3. **Canonical journey graph:** schema, validation and simulator before any
   visual editor component.
4. **Outbound LDAP:** OpenLDAP first, then the Windows AD profile; separate
   authentication and provisioning lifecycle modes.
5. **Portable edges:** common signed configuration/session contract, inbound
   LDAP, forward-auth, then identity-aware proxy.
6. **Shippin projection:** guided plans and the visual editor consume only
   stable Tessera management contracts and capability facts.

## Workspace ready gate

The foundation is ready when:

- it works from the nested worktree and the repository root;
- `init` creates only ignored, permission-restricted state and prints no value;
- `doctor --profile core` passes on Linux and Windows;
- `doctor --profile workspace` verifies all sibling contracts and the exact,
  clean design-reference revision on an operator workstation;
- the manifest contains all six mandatory capability ids from contract 18;
- sibling boundaries resolve without reading a credential or live database;
- the existing dev loop contains no checked-in master key;
- source inventory and product-language checks remain green;
- no production service is started as part of setup or validation.
