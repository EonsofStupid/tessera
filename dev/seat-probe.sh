#!/usr/bin/env bash
# From a running Tessera to a verified seat token.
#
#   bash dev/up.sh && bash dev/seat-probe.sh
#
# Provisions one seat, mints tokens for two workspaces, the Shippin panel, and
# one workspace it may not have, then hands them to Automaton's own verifier.
# Everything it writes lands in
# `.artifacts/`, which is gitignored — the client secret it creates is a live
# credential and never enters this repository (`AGENTS.md`).
#
# Re-runnable: the machine user is recreated only if it is missing, and the
# secret is reissued every time because Tessera returns it exactly once.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
API="http://localhost:${TESSERA_PORT:-8088}"
PAT="$(cat "$ROOT/.artifacts/admin.pat")"
WS_A=ws-0001
WS_B=ws-0002
WS_NONE=ws-0009

step() { printf '\n\033[1m%s\033[0m\n' "$*"; }
ok()   { printf '  ✓ %s\n' "$*"; }
api()  { curl -sf -X "$1" "$API$2" -H "Authorization: Bearer $PAT" -H 'Content-Type: application/json' "${@:3}"; }
jqp()  { python3 -c "import json,sys; d=json.load(sys.stdin); print(d$1)"; }

step "0 · the API answers"
# On a first boot the management API exists before its projections have caught
# up, and a request in that window fails as if the server were broken. Waiting
# on a real management call — not on OIDC discovery, which comes up earlier —
# is the difference between a probe that works and one that works usually.
for _ in $(seq 60); do
  api POST /management/v1/users/_search -d '{"queries":[]}' >/dev/null 2>&1 && break
  sleep 0.5
done
api POST /management/v1/users/_search -d '{"queries":[]}' >/dev/null 2>&1 || {
  printf '  ✗ management API did not become ready — see .artifacts/run.log\n'; exit 1; }
ok "management API ready"

step "1 · the seat"
USER_ID="$(api POST /management/v1/users/machine -d '{
  "userName":"seat-probe","name":"Seat probe",
  "description":"proves the mint path; docs/05-minting-a-seat-token.md",
  "accessTokenType":"ACCESS_TOKEN_TYPE_JWT"}' 2>/dev/null | jqp '["userId"]' 2>/dev/null || true)"
if [[ -z "${USER_ID:-}" ]]; then
  # Already there from a previous run — find it rather than fail.
  USER_ID="$(api POST /management/v1/users/_search -d '{"queries":[{"userNameQuery":{"userName":"seat-probe"}}]}' \
    | jqp '["result"][0]["id"]')"
  ok "seat-probe already exists ($USER_ID)"
else
  ok "created ($USER_ID)"
fi

step "2 · its facts — declared, not typed"
# Seats arrive as a blueprint now: rendered from the reviewed template, checked
# by `blueprint validate` (no database), applied atomically, and applied AGAIN —
# because a blueprint that is not a no-op the second time is not declarative,
# and the probe should be the first thing to notice.
PSQL="psql -h 127.0.0.1 -p 5433 -U tessera_app -d tessera -tAc"
INSTANCE="$($PSQL "select id from zitadel.instances limit 1")"
ACCOUNT="$($PSQL "select id from zitadel.organizations limit 1")"
TESSERA="$ROOT/.artifacts/tessera"
BPDIR="$ROOT/.artifacts/blueprints"

mkdir -p "$BPDIR"
sed -e "s/@MEMBER_ID@/$USER_ID/" -e "s/@ACCOUNT_ID@/$ACCOUNT/" \
  "$ROOT/blueprints/dev/seats.yaml.tmpl" > "$BPDIR/seats.yaml"
cp "$ROOT/blueprints/dev/flows.yaml" "$BPDIR/flows.yaml"

"$TESSERA" blueprint validate --dir "$BPDIR" >/dev/null 2>&1 && ok "blueprint validates (no database needed)"
"$TESSERA" blueprint apply --config "$ROOT/dev/dev.yaml" --dir "$BPDIR" --instance "$INSTANCE" 2>/dev/null | sed 's/^/  ✓ /'
SECOND="$("$TESSERA" blueprint apply --config "$ROOT/dev/dev.yaml" --dir "$BPDIR" --instance "$INSTANCE" 2>/dev/null)"
if printf '%s' "$SECOND" | grep -q "converged: nothing to change"; then
  ok "second apply converged — the blueprint is a no-op, as declared state must be"
else
  printf '  ✗ second apply was not a no-op:\n%s\n' "$SECOND"; exit 1
fi

# Prove the old source is gone rather than merely unused: if a single seat fact
# were still readable from metadata, every assertion below would pass for the
# wrong reason.
LEFTOVER="$($PSQL "select count(*) from zitadel.user_metadata where key like 'shippin%'")"
if [[ "$LEFTOVER" != "0" ]]; then
  printf '  ✗ %s seat facts still in user metadata — the old path is still live\n' "$LEFTOVER"
  exit 1
fi
ok "no seat facts remain in user metadata"

step "3 · credentials"
# Returned exactly once, so it is always reissued rather than cached.
SECRET_JSON="$(api PUT "/management/v1/users/$USER_ID/secret" -d '{}')"
CID="$(printf '%s' "$SECRET_JSON" | jqp '["clientId"]')"
CSE="$(printf '%s' "$SECRET_JSON" | jqp '["clientSecret"]')"
ok "client_id $CID, secret reissued"

step "4 · minting"
mint() {
  mint_for automaton "$1"
}
mint_for() {
  local consumer="$1" workspace="$2"
  curl -s -X POST "$API/oauth/v2/token" -u "$CID:$CSE" \
    -d grant_type=client_credentials \
    --data-urlencode "scope=openid urn:shippin:audience:${consumer}:${workspace}"
}
for ws in "$WS_A" "$WS_B"; do
  mint "$ws" | jqp '["access_token"]' > "$ROOT/.artifacts/tok-${ws/ws-/ws}.txt"
  python3 - "$ROOT/.artifacts/tok-${ws/ws-/ws}.txt" <<'PY'
import base64, json, sys
p = open(sys.argv[1]).read().strip().split(".")[1]
d = json.loads(base64.urlsafe_b64decode(p + "=" * (-len(p) % 4)))
print(f"  ✓ {d['workspace_id']}: {d['schema']}, aud={d['aud']}, "
      f"{d['occupant']}/{d['basis']}, {d['exp']-d['iat']}s, "
      f"scopes={','.join(d['authorization']['scopes'])}")
PY
done
mint_for shippin "$WS_A" | jqp '["access_token"]' > "$ROOT/.artifacts/tok-shippin-ws0001.txt"
ok "minted short-lived Shippin panel handoff token for $WS_A"
REFUSAL="$(mint "$WS_NONE")"
if printf '%s' "$REFUSAL" | grep -q invalid_target; then
  ok "$WS_NONE refused: $(printf '%s' "$REFUSAL" | jqp '["error"]')"
else
  printf '  ✗ %s was not refused: %s\n' "$WS_NONE" "$REFUSAL"; exit 1
fi

step "5 · the consumer decides"
COMMON_GIT_DIR="$(git -C "$ROOT" rev-parse --path-format=absolute --git-common-dir)"
CANONICAL_REPO="$(dirname "$COMMON_GIT_DIR")"
AUTOMATON_IDENTITY_MODULE="${AUTOMATON_IDENTITY_MODULE:-$(dirname "$CANONICAL_REPO")/automaton/engine/serve/identity.mjs}" \
  node "$ROOT/dev/verify-seat-token.mjs"
