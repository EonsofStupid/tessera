#!/usr/bin/env bash
# Tessera's trunk, from nothing to serving OIDC.
#
#   bash dev/up.sh            build if needed, start Postgres, init, setup, run
#   bash dev/up.sh --rebuild  regenerate and rebuild first
#   bash dev/up.sh --down     stop the server and the database
#
# Its own Postgres on 5433, never the system cluster on 5432 — whose purpose on
# this box is not ours to assume. The data directory is `.pgdata` and is
# disposable: delete it and this script rebuilds the world.
#
# The generator chain and the four traps in it are docs/04-building-the-trunk.md.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TRUNK="$ROOT"
PGDATA="$ROOT/.pgdata"
CONFIG="$ROOT/dev/dev.yaml"
STEPS="$ROOT/dev/steps.yaml"
BIN="$ROOT/.artifacts/tessera"
PORT="${TESSERA_PORT:-8088}"
# 32 characters exactly, and a dev value on purpose: a real one never lives in
# a repository (../CLYFFY.md, and AGENTS.md here).
MASTERKEY="MasterkeyNeedsToHave32Characters"

export PATH="/usr/lib/postgresql/18/bin:$ROOT/.artifacts/bin:$PATH"

step() { printf '\n\033[1m%s\033[0m\n' "$*"; }
ok()   { printf '  ✓ %s\n' "$*"; }

# Fail before half-running: the quick preflight proves binaries, ports and
# writability, and every failure it prints names its exact fix (sudo included
# where root is really needed). Full version: bash dev/preflight.sh
if [[ "${1:-}" != "--down" ]]; then
  bash "$ROOT/dev/preflight.sh" --quick || exit 1
fi

if [[ "${1:-}" == "--down" ]]; then
  pkill -f "[t]essera --config $CONFIG" 2>/dev/null || true
  pg_ctl -D "$PGDATA" stop -m fast 2>/dev/null || true
  ok "stopped"
  exit 0
fi

# ---- 1 · the database ------------------------------------------------------
step "1 · postgres on 5433"
if [[ ! -d "$PGDATA" ]]; then
  initdb -D "$PGDATA" -U tessera --auth-local=trust --auth-host=trust >/dev/null
  cat >> "$PGDATA/postgresql.conf" <<EOF

# Our own TCP-loopback cluster. Unix sockets are disabled because nested
# worktree paths can exceed the operating system's socket path limit.
port = 5433
listen_addresses = '127.0.0.1'
unix_socket_directories = ''
EOF
  ok "cluster created"
fi
# Migrate clusters created by an older launcher that used a worktree-relative
# Unix socket. This changes only the disposable ignored development cluster.
sed -i "/^unix_socket_directories = /c\\unix_socket_directories = ''" "$PGDATA/postgresql.conf"
pg_ctl -D "$PGDATA" -l "$PGDATA/server.log" start >/dev/null 2>&1 || true
database_ready=false
for _ in $(seq 30); do
  if psql -h 127.0.0.1 -p 5433 -U tessera -d postgres -c 'select 1' >/dev/null 2>&1; then
    database_ready=true
    break
  fi
  sleep 0.3
done
if [[ "$database_ready" != true ]]; then
  printf '\n  postgres did not become ready — database log follows\n\n' >&2
  tail -20 "$PGDATA/server.log" >&2
  exit 1
fi
psql -h 127.0.0.1 -p 5433 -U tessera -d postgres -tAc \
  "select 1 from pg_roles where rolname='zitadel'" | grep -q 1 ||
  psql -h 127.0.0.1 -p 5433 -U tessera -d postgres -c "CREATE ROLE zitadel LOGIN SUPERUSER" >/dev/null
psql -h 127.0.0.1 -p 5433 -U tessera -d postgres -tAc \
  "select 1 from pg_database where datname='zitadel'" | grep -q 1 ||
  psql -h 127.0.0.1 -p 5433 -U tessera -d postgres -c "CREATE DATABASE zitadel" >/dev/null
ok "listening, role and database present"

# ---- 2 · the binary --------------------------------------------------------
step "2 · the trunk"
if [[ ! -x "$BIN" || "${1:-}" == "--rebuild" ]]; then
  mkdir -p "$ROOT/.artifacts"
  cd "$TRUNK"
  if [[ ! -d pkg/grpc || "${1:-}" == "--rebuild" ]]; then
    bash "$ROOT/dev/generate.sh"
    ok "generated"
  fi
  go build -o "$BIN" .
  ok "built $(du -h "$BIN" | cut -f1)"
else
  ok "already built"
fi

# ---- 3 · schema and instance ----------------------------------------------
step "3 · init and setup"
cd "$TRUNK"
if ! psql -h 127.0.0.1 -p 5433 -U tessera -d zitadel -tAc \
     "select 1 from information_schema.tables where table_name='events2'" | grep -q 1; then
  "$BIN" init --config "$CONFIG" >/dev/null
  "$BIN" setup --config "$CONFIG" --steps "$STEPS" --masterkey "$MASTERKEY" --init-projections >/dev/null
  ok "eventstore, projections and the first instance"
  ok "admin PAT at .artifacts/admin.pat (gitignored, reissued whenever .pgdata is)"
else
  ok "already initialised"
fi

# ---- 4 · run ---------------------------------------------------------------
step "4 · serving"
nohup "$BIN" start --config "$CONFIG" --masterkey "$MASTERKEY" \
  > "$ROOT/.artifacts/run.log" 2>&1 &
for _ in $(seq 40); do
  curl -sf "http://localhost:$PORT/.well-known/openid-configuration" >/dev/null 2>&1 && break
  sleep 0.5
done

if curl -sf "http://localhost:$PORT/.well-known/openid-configuration" >/dev/null 2>&1; then
  ok "OIDC discovery at http://localhost:$PORT/.well-known/openid-configuration"
  ok "JWKS at $(curl -s "http://localhost:$PORT/.well-known/openid-configuration" |
        python3 -c 'import json,sys; print(json.load(sys.stdin)["jwks_uri"])')"
  printf '\n  Automaton verifies against it with:\n\n    AUTOMATON_OIDC_ISSUER=http://localhost:%s npm run serve\n\n' "$PORT"
else
  printf '\n  did not come up — see .artifacts/run.log\n\n'
  tail -5 "$ROOT/.artifacts/run.log"
  exit 1
fi
