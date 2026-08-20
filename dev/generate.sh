#!/usr/bin/env bash
# Generate every ignored protocol/runtime artifact required by build and test.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="$ROOT/.artifacts/bin"

bash "$ROOT/dev/toolchain.sh"
export PATH="$BIN:$PATH"
cd "$ROOT"

# The custom protocol plugin consumes its own option type. Bootstrap that
# option with the standard Go generator before compiling the custom plugins.
buf generate \
  --template buf.gen.bootstrap.yaml \
  --path proto/zitadel/protoc_gen_zitadel/v2/options.proto
mkdir -p pkg/grpc/protoc/v2
cp .artifacts/grpc/github.com/EonsofStupid/tessera/pkg/grpc/protoc/v2/*.go pkg/grpc/protoc/v2/

GOBIN="$BIN" go install \
  ./internal/protoc/protoc-gen-authoption \
  ./internal/protoc/protoc-gen-zitadel

buf generate
mkdir -p pkg/grpc openapi/v2/zitadel apps/docs/content/apis/assets
cp -R .artifacts/grpc/github.com/EonsofStupid/tessera/pkg/grpc/. pkg/grpc/
cp .artifacts/grpc/zitadel/*.swagger.json openapi/v2/zitadel/

(cd internal/api/assets/generator && go run . -directory ./)
go generate internal/api/ui/login/statik/generate.go
go generate internal/notification/statik/generate.go
go generate internal/statik/generate.go

printf '  generated protocol, API and embedded runtime artifacts\n'
