#!/usr/bin/env bash
# Preflight: prove the environment can run the dev loop and the test suite
# BEFORE anything half-runs and fails in the middle.
#
#   bash dev/preflight.sh            all checks, remediation for each failure
#   bash dev/preflight.sh --quick    the fast subset up.sh runs before starting
#   bash dev/preflight.sh --fix      also perform the fixes that need no root
#   bash dev/preflight.sh --fetch-embedded   populate the embedded-PG cache only
#
# The rule for failures: every ✗ prints the exact command that fixes it, with
# sudo spelled out where root is genuinely required — never run by this script
# itself. Whether a person or an AI is driving, the output IS the runbook; a
# check that fails without naming its fix is a check that wasted the run.
set -uo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MODE="${1:-full}"
FAILS=0
FIX=0
[[ "$MODE" == "--fix" ]] && { FIX=1; MODE="full"; }

ok()   { printf '  \033[32m✓\033[0m %s\n' "$*"; }
warn() { printf '  \033[33m•\033[0m %s\n' "$*"; }
bad()  { printf '  \033[31m✗\033[0m %s\n' "$1"; printf '      fix: %s\n' "$2"; FAILS=$((FAILS+1)); }
step() { printf '\n\033[1m%s\033[0m\n' "$*"; }

# ---- fetch-embedded is its own mode ---------------------------------------
if [[ "$MODE" == "--fetch-embedded" ]]; then
  cd "$ROOT" && exec go run ./dev/prefetch-embedded
fi

step "toolchain"
if command -v go >/dev/null; then
  GOV="$(go env GOVERSION 2>/dev/null)"
  case "$GOV" in
    go1.2[5-9]*|go1.[3-9]*) ok "go $GOV" ;;
    *) bad "go is $GOV, need 1.25+" "install from https://go.dev/dl/ or: sudo apt-get install -y golang-1.25" ;;
  esac
else
  bad "go not on PATH" "sudo apt-get install -y golang-1.25, or install from https://go.dev/dl/"
fi

step "postgres server binaries"
PGBIN=""
for d in /usr/lib/postgresql/*/bin; do [[ -x "$d/initdb" ]] && PGBIN="$d"; done
if [[ -n "$PGBIN" ]]; then
  ok "initdb/pg_ctl/psql at $PGBIN"
elif command -v initdb >/dev/null; then
  ok "initdb on PATH ($(command -v initdb))"
else
  bad "no postgres server binaries (initdb) found" \
      "sudo apt-get install -y postgresql-18  # or postgresql-16+; dev/up.sh looks in /usr/lib/postgresql/*/bin"
fi

step "ports"
port_owner() { ss -tlnp 2>/dev/null | awk -v p=":$1" '$4 ~ p"$" {print $NF}' | head -1; }
for spec in "5433:the dev cluster (postgres)" "8088:nomen (OIDC)"; do
  PORT="${spec%%:*}"; WHAT="${spec#*:}"
  OWNER="$(port_owner "$PORT")"
  if [[ -z "$OWNER" ]]; then
    ok "port $PORT free — $WHAT can bind it"
  elif [[ "$OWNER" == *postgres* && "$PORT" == 5433 ]] || [[ "$OWNER" == *nomen* && "$PORT" == 8088 ]]; then
    ok "port $PORT held by ours ($WHAT already up)"
  else
    bad "port $PORT held by something else: $OWNER" \
        "sudo lsof -i :$PORT   # identify it, then stop it or export NOMEN_PORT for 8088"
  fi
done

step "writable state"
for d in "$ROOT/.artifacts" "$ROOT/.pgdata"; do
  PARENT="$d"; [[ -e "$d" ]] || PARENT="$(dirname "$d")"
  if [[ -w "$PARENT" ]]; then
    ok "$(basename "$d") writable"
  else
    bad "$(basename "$d") not writable ($PARENT)" \
        "sudo chown -R \"$(id -un)\" \"$PARENT\"   # something (likely a root run) took ownership"
  fi
done

if [[ "$MODE" == "--quick" ]]; then
  [[ $FAILS -eq 0 ]] && exit 0
  printf '\n%d preflight failure(s) — fix the above, then retry.\n' "$FAILS"; exit 1
fi

step "embedded postgres (the hermetic test harness)"
CACHE="$HOME/.embedded-postgres-go"
if compgen -G "$CACHE/*16*" >/dev/null 2>&1; then
  ok "PG 16 binary cached at $CACHE — tests need no network"
else
  # No cache yet: can we reach the binary source?
  if curl -sf -m 8 -o /dev/null -r 0-0 "https://repo1.maven.org/maven2/io/zonky/test/postgres/embedded-postgres-binaries-linux-amd64/maven-metadata.xml" 2>/dev/null; then
    if [[ $FIX -eq 1 ]]; then
      warn "cache empty, network fine — fetching now (~60s once)"
      (cd "$ROOT" && go run ./dev/prefetch-embedded) && ok "cache populated" || \
        bad "prefetch failed despite network" "re-run: bash dev/preflight.sh --fetch-embedded  # read its error"
    else
      bad "no cached PG 16 binary; first test run would hit the network" \
          "bash dev/preflight.sh --fetch-embedded   # one-time ~60s download to $CACHE"
    fi
  else
    bad "no cached binary AND repo1.maven.org unreachable — 3.3b's hermetic test cannot run here" \
        "from a network-enabled shell: bash dev/preflight.sh --fetch-embedded ; behind a proxy: HTTPS_PROXY=... bash dev/preflight.sh --fetch-embedded ; air-gapped: copy ~/.embedded-postgres-go from any machine that has it. Until then, dev/blueprint-probe.sh covers the same assertions on the 5433 cluster."
  fi
fi

step "optional (rebuild-only)"
if command -v pnpm >/dev/null; then ok "pnpm (login-v1 statik assets)"; else
  warn "pnpm missing — only matters for dev/up.sh --rebuild of login-v1 assets: npm install -g pnpm --prefix ~/.npm-global"
fi

printf '\n'
if [[ $FAILS -eq 0 ]]; then
  printf 'preflight clean — everything the dev loop and tests need is present.\n'
else
  printf '%d failure(s). Each ✗ above names its exact fix; run those, then re-run: bash dev/preflight.sh\n' "$FAILS"
  printf 'Fixes marked sudo genuinely need root; everything else: bash dev/preflight.sh --fix\n'
  exit 1
fi
