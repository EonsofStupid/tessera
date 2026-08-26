# Source provenance inventory

This directory records where Nomen's tracked source came from. Nomen's product
license is AngryVibes LLC and shippin.ai only (`LICENSE`, `NOTICE`,
`docs/23-release-governance-record.md`). The BOM must not assign AGPL, Apache,
or any other third-party product license to Nomen runtime paths.

`source-manifest.json` is reviewed policy. `source-bom.json` is generated
deterministically from that policy and full Git history. It groups every
tracked path by primary source, relationship, license expression and rule. It
contains no generated timestamp.

Run the complete local check from the repository root:

```sh
bash dev/check-source-provenance.sh
```

Regenerate after an approved source/path rule change:

```sh
go run ./dev/source-inventory
```

The generator needs full Git history because the baseline import commit is
part of the evidence. New paths under inherited source roots fail closed until
an explicit rule is reviewed. Dependency SBOMs, container packages, build
provenance and release artifacts are separate G1 controls.

