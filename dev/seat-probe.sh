#!/usr/bin/env bash
# From a running Tessera to a verified seat token.
#
#   bash dev/up.sh && bash dev/seat-probe.sh
#
# Provisions one seat, mints tokens for two workspaces and one it may not have,
# then hands them to Automaton's own verifier. Everything it writes lands in
# `.artifacts/`, which is gitignored — the client secret it creates is a live
# credential and never enters this repository (`AGENTS.md`).
#
# Re-runnable: the machine user is recreated only if it is missing, and the
# secret is reissued every time because Zitadel returns it exactly once.
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

step "1 · the seat"
USER_ID="$(api POST /management/v1/users/machine -d '{
  "userName":"seat-probe","name":"Seat probe",
  "description":"proves the mint path; docs/05-minting-a-seat-token.md",
  "accessTokenType":"ACCESS_TOKEN_TYPE_JWT"}' 2>/dev/null | jqp '["userId"]' || true)"
if [[ -z "${USER_ID:-}" ]]; then
  # Already there from a previous run — find it rather than fail.
  USER_ID="$(api POST /management/v1/users/_search -d '{"queries":[{"userNameQuery":{"userName":"seat-probe"}}]}' \
    | jqp '["result"][0]["id"]')"
  ok "seat-probe already exists ($USER_ID)"
else
  ok "created ($USER_ID)"
fi

step "2 · its facts"
# Seats live in `tessera.seats` now, not in Zitadel user metadata, and they are
# written through the same repository blueprints will use — so this script
# cannot pass a test the real path would fail.
#
# The ids come from the database because a dev box has one instance and one
# organization and there is no API that hands them over more cheaply.
PSQL="psql -h 127.0.0.1 -p 5433 -U tessera -d zitadel -tAc"
INSTANCE="$($PSQL "select id from zitadel.instances limit 1")"
ACCOUNT="$($PSQL "select id from zitadel.organizations limit 1")"
TESSERA="$ROOT/.artifacts/tessera"

"$TESSERA" seat set --config "$ROOT/dev/dev.yaml" \
  --instance "$INSTANCE" --member "$USER_ID" --account "$ACCOUNT" \
  --occupant agent --basis subscription \
  --workspaces "$WS_A,$WS_B" \
  --scopes 'hosting.active,terminal:advanced,chat.unified' \
  --policy-version pol_2026_08_17 2>/dev/null | sed 's/^/  ✓ /'

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
  curl -s -X POST "$API/oauth/v2/token" -u "$CID:$CSE" \
    -d grant_type=client_credentials \
    --data-urlencode "scope=openid urn:shippin:audience:automaton:$1"
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
REFUSAL="$(mint "$WS_NONE")"
if printf '%s' "$REFUSAL" | grep -q invalid_target; then
  ok "$WS_NONE refused: $(printf '%s' "$REFUSAL" | jqp '["error"]')"
else
  printf '  ✗ %s was not refused: %s\n' "$WS_NONE" "$REFUSAL"; exit 1
fi

step "5 · the consumer decides"
node "$ROOT/dev/verify-seat-token.mjs"
