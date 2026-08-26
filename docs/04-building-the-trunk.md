# 04 — Building Nomen cleanly

**Status:** reproducible from a clean child worktree.

Nomen contains generated protocol clients, OpenAPI documents and embedded
runtime assets. The supported build path pins every generator under
`.artifacts/bin`, bootstraps the protocol option needed by Nomen's custom
generator, generates the complete graph, then creates the embedded assets.

## Canonical build

Requirements are Go 1.25 or newer, PostgreSQL client tools and the standard
archive/download utilities used by `dev/toolchain.sh`.

```bash
bash dev/generate.sh
go test ./...
go build -o .artifacts/nomen .
```

`dev/generate.sh` is the single authority for ordering. It:

1. installs the pinned generators into `.artifacts/bin`;
2. generates the bootstrap protocol option;
3. builds Nomen's custom protocol plugins;
4. generates Go, Connect, gRPC and OpenAPI outputs;
5. copies the outputs into their runtime package locations;
6. generates the assets and embedded statik archives.

No global tool install is required, and the script does not read a parent
checkout. A child worktree must produce the same generated result.

## Runnable development instance

```bash
bash dev/up.sh --rebuild
```

The command creates ignored workspace state, starts PostgreSQL on loopback,
applies migrations, generates an asymmetric signing key and waits for OIDC
discovery plus JWKS readiness. PostgreSQL's Unix socket is disabled so deeply
nested worktree paths cannot exceed the platform socket-path limit.

Live credentials are generated only under ignored `.artifacts` state and are
never printed. Integration signing fixtures are generated in memory unless a
protected external integration estate explicitly supplies key-file paths.

## Acceptance gates

- `nomen --help` and the runtime banner use Nomen product language;
- discovery names the configured Nomen issuer;
- JWKS publishes only allowed asymmetric signing algorithms;
- `GET /nomen/v1/overview` and `/capabilities` enforce typed authorization;
- `bash dev/seat-probe.sh` applies declared seats and flows twice, proves the
  second apply is a no-op, mints a 15-minute token and passes Automaton's real
  verifier;
- no generated credential or application database is tracked by Git.

These are build facts. Capability availability is reported separately by the
conformance-backed discovery contract and is never inferred from a successful
compile or a listening process.
