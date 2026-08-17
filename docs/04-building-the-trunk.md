# 04 — Building the trunk

**Status:** done and reproducible. `upstream/zitadel` builds and runs on this
box.

Zitadel does not build from a clone. Roughly a third of the Go source is
generated and none of it is checked in, so `go build ./...` on a fresh clone
fails with a wall of "no required module provides package". The recipe below is
what actually works, recorded because working it out took six separate
discoveries and nobody should pay that twice.

The authority is `apps/api/project.json` — their nx targets are the real build
documentation, more so than anything in the README.

## Toolchain

Fourteen binaries, **pinned**, installed into `.artifacts/bin/$GOOS/$GOARCH`
rather than onto the PATH. That isolation is deliberate and theirs: *"We avoid
using go tools so the dev tool dependencies don't interfere with the prod
dependencies."*

```bash
cd upstream/zitadel
python3 -c "
import json; d=json.load(open('apps/api/project.json'))
print('\n'.join(d['targets']['generate-install']['options']['commands']))
" | bash -e
```

That installs `buf`, the six protoc plugins, Zitadel's own two
(`protoc-gen-authoption`, `protoc-gen-zitadel`, built from `internal/protoc/`),
plus `statik`, `enumer`, `mockgen`, `stringer` and `gci`.

Also needed on the box: **Go 1.25+** (1.26 here), and **pnpm** — but only for
the login-v1 UI, which we are dropping. `npm install -g pnpm --prefix
~/.npm-global`, never into `/opt/node`, which is not ours.

## Three generators, in order

```bash
export PATH="${PWD}/.artifacts/bin/linux/amd64:$PATH"

# 1 · protobuf → 466 Go files
buf generate
mkdir -p pkg/grpc openapi/v2/zitadel
cp -r .artifacts/grpc/github.com/zitadel/zitadel/pkg/grpc/** pkg/grpc/
cp .artifacts/grpc/zitadel/*.swagger.json openapi/v2/zitadel/

# 2 · the assets service — run from the generator's own directory
(cd internal/api/assets/generator && go run . -directory ./)

# 3 · statik: embedded i18n, email templates, login assets
go generate internal/api/ui/login/statik/generate.go
go generate internal/notification/statik/generate.go
go generate internal/statik/generate.go

go build -o tessera .
```

## The four traps

**`pkg/grpc` is generated, not committed.** Every `internal/api/grpc/**` package
imports it, so nothing builds until `buf generate` has run and the output has
been *copied* out of `.artifacts/`. The copy is a separate step; generating
alone is not enough.

**`openapi/handler.go` embeds `v2/zitadel/*`.** A `go:embed` with no matching
files is a compile error, not a warning — the swagger JSON has to be copied too.

**The assets generator truncates before it validates.** It opens `authz.go`,
`router.go` and a docs file with `O_TRUNC` up front, then fails if the docs path
is missing — leaving two zero-byte Go files behind that break the build more
confusingly than the original error did. Its docs path is relative to the
*working directory*, not to `-directory`, so it must be run from
`internal/api/assets/generator/`.

**Statik is why a successful build still panics.** Without it the binary
compiles and then dies at startup with `statik/fs: no zip data registered` —
i18n translations are embedded, not read from disk. The first statik step needs
`pnpm` and `sass` to compile the login-v1 stylesheets; skip it and the other
three still work, which is the right trade when that UI is being dropped.

## Where this leaves us

```
$ ./tessera --help
Available Commands:
  init             initialize ZITADEL instance
  setup            setup ZITADEL instance
  start            starts ZITADEL instance
  start-from-init  cold starts zitadel
  mirror           mirrors all data from one database to another
  keys             manage encryption keys
```

The trunk stands. Next is Postgres and `start-from-init`, then the fork:
module path, strip what we will never ship, and the first Tessera-shaped change.
