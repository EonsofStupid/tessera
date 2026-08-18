# Source provenance inventory

This directory records where Tessera's tracked source came from. It is an
engineering control, not a substitute for the top-level license and notices
required by P0.3 in `docs/12-execution-worklist.md`.

`source-manifest.json` is the reviewed policy:

- commit `31c07a1` is the bulk trunk import recorded by Tessera's own history;
- paths added by that commit are classified as ZITADEL-derived;
- `proto/` receives the Apache-2.0 directory exception documented by the pinned
  ZITADEL reference revision;
- later blueprint and flow implementations remain Tessera-native while naming
  authentik as design inspiration;
- later files under inherited source roots require an explicit path rule. This
  makes a new unreviewed `internal/`, `cmd/`, `backend/` or `proto/` intake fail
  rather than silently appear native.

`source-bom.json` is generated deterministically from the manifest and Git
history. It groups every tracked path by primary source, relationship, license
expression and rule. It deliberately contains no generated timestamp.

Run the check from the repository root:

```sh
./dev/check-source-provenance.sh
```

Regenerate after a reviewed path or rule change:

```sh
go run ./dev/source-inventory
```

The generator needs full Git history because the baseline import commit is part
of the evidence. Dependency SBOMs, container packages and generated build
artifacts are separate release controls under P10.4.
