#!/usr/bin/env bash
# Install Tessera's pinned, repository-local protocol toolchain.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="$ROOT/.artifacts/bin"
mkdir -p "$BIN"

install_go_tool() {
  local binary="$1"
  local package="$2"
  if [[ -x "$BIN/$binary" ]]; then
    return
  fi
  printf '  installing %s\n' "$binary"
  GOBIN="$BIN" go install "$package"
}

install_go_tool buf github.com/bufbuild/buf/cmd/buf@v1.67.0
install_go_tool protoc-gen-go google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11
install_go_tool protoc-gen-go-grpc google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.1
install_go_tool protoc-gen-grpc-gateway github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.28.0
install_go_tool protoc-gen-openapiv2 github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.28.0
install_go_tool protoc-gen-validate github.com/envoyproxy/protoc-gen-validate@v1.3.3
install_go_tool protoc-gen-connect-go connectrpc.com/connect/cmd/protoc-gen-connect-go@v1.19.1
install_go_tool statik github.com/rakyll/statik@v0.1.8

printf '  protocol toolchain ready in %s\n' "$BIN"
