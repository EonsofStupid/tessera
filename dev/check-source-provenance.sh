#!/usr/bin/env bash
set -euo pipefail

repo_root=$(git rev-parse --show-toplevel)
cd "$repo_root"

go test ./dev/source-inventory
go run ./dev/source-inventory -check
