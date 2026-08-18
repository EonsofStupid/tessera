#!/usr/bin/env bash
# Phase 4's sentence, proven on this box: password, TOTP and recovery are
# three YAML files executed by one engine.
#
#   bash dev/up.sh && bash dev/seat-probe.sh && bash dev/flow-probe.sh
#
# Drives the login-password flow over HTTP as a real customer would — start,
# identify, wrong password (must fail closed and stay), right password, done —
# then verifies the resulting session in the eventstore carries the same
# factor events the session v2 API writes. No fixture stands in for anything.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
API="http://localhost:${TESSERA_PORT:-8088}"
PAT="$(cat "$ROOT/.artifacts/admin.pat")"
PSQL="psql -h 127.0.0.1 -p 5433 -U tessera -d zitadel -tAc"
HUMAN="flow-probe-human"
PASSWORD="Fl0w-probe-Passw0rd!"

step() { printf '\n\033[1m%s\033[0m\n' "$*"; }
ok()   { printf '  ✓ %s\n' "$*"; }
die()  { printf '  ✗ %s\n' "$*"; exit 1; }
jqp()  { python3 -c "import json,sys; d=json.load(sys.stdin); print(d$1)"; }

step "1 · a human with a password"
EXISTING="$($PSQL "select id from projections.users14 where username = '$HUMAN' limit 1")"
if [[ -z "$EXISTING" ]]; then
  curl -sf -X POST "$API/management/v1/users/human/_import" \
    -H "Authorization: Bearer $PAT" -H 'Content-Type: application/json' \
    -d "{\"userName\":\"$HUMAN\",
         \"profile\":{\"firstName\":\"Flow\",\"lastName\":\"Probe\"},
         \"email\":{\"email\":\"flow-probe@dev.invalid\",\"isEmailVerified\":true},
         \"password\":\"$PASSWORD\",\"passwordChangeRequired\":false}" >/dev/null
  ok "imported $HUMAN"
else
  ok "$HUMAN already exists"
fi

step "2 · the flows exist because a blueprint declared them"
for slug in login-password login-mfa recovery; do
  N="$($PSQL "select count(*) from tessera.flow_stages fs join tessera.flows f
      on f.instance_id = fs.instance_id and f.slug = fs.flow_slug
      where f.slug = '$slug'")"
  [[ "$N" -ge 2 ]] || die "$slug missing or empty — run dev/seat-probe.sh first (it applies the blueprints)"
  ok "$slug: $N stages"
done

step "3 · start login-password"
START="$(curl -s -X POST "$API/flows/v1/login-password/start")"
EXEC="$(printf '%s' "$START" | jqp '["execution_id"]')"
TOKEN="$(printf '%s' "$START" | jqp '["token"]')"
COMPONENT="$(printf '%s' "$START" | jqp '["challenge"]["component"]')"
[[ "$COMPONENT" == "tessera-stage-identify" ]] || die "first challenge = $COMPONENT"
ok "challenge: $COMPONENT"

answer() { # answer <json-answer> → response body
  curl -s -X POST "$API/flows/v1/executions/$EXEC" \
    -H 'Content-Type: application/json' \
    -d "{\"token\":\"$TOKEN\",\"answer\":$1}"
}

step "4 · a wrong token is a 404, not a hint"
CODE="$(curl -s -o /dev/null -w '%{http_code}' -X POST "$API/flows/v1/executions/$EXEC" \
  -H 'Content-Type: application/json' -d '{"token":"wrong","answer":{}}')"
[[ "$CODE" == "404" ]] || die "wrong token answered $CODE"
ok "wrong token → 404, indistinguishable from an unknown execution"

step "5 · identify"
R="$(answer "{\"identifier\":\"$HUMAN\"}")"
[[ "$(printf '%s' "$R" | jqp '["component"]')" == "tessera-stage-password" ]] || die "after identify: $R"
ok "identified; challenge: tessera-stage-password"

step "6 · the wrong password fails closed and stays"
R="$(answer '{"password":"not-the-password"}')"
[[ "$(printf '%s' "$R" | jqp '["component"]')" == "tessera-stage-password" ]] || die "wrong password moved the flow: $R"
printf '%s' "$R" | jqp '["errors"]["password"][0]' >/dev/null || die "no field error on wrong password: $R"
ok "re-asked with a field error, position unchanged"

step "7 · the right password completes the flow"
R="$(answer "{\"password\":\"$PASSWORD\"}")"
SESSION_ID="$(printf '%s' "$R" | jqp '["session_id"]')"
SESSION_TOKEN="$(printf '%s' "$R" | jqp '["session_token"]')"
[[ -n "$SESSION_ID" && -n "$SESSION_TOKEN" ]] || die "no session came back: $R"
ok "session $SESSION_ID, token delivered once"

step "8 · the execution is consumed"
CODE="$(curl -s -o /dev/null -w '%{http_code}' -X POST "$API/flows/v1/executions/$EXEC" \
  -H 'Content-Type: application/json' -d "{\"token\":\"$TOKEN\",\"answer\":{}}")"
[[ "$CODE" == "404" ]] || die "a completed execution answered $CODE"
ok "answering again → 404; the session token crossed the wire exactly once"

step "9 · the session carries the same factor events the session v2 API writes"
EVENTS="$($PSQL "select event_type from eventstore.events2
  where aggregate_type = 'session' and aggregate_id = '$SESSION_ID'
  order by \"sequence\"")"
printf '%s\n' "$EVENTS" | grep -q "session.user.checked"     || die "no user.checked event: $EVENTS"
printf '%s\n' "$EVENTS" | grep -q "session.password.checked" || die "no password.checked event: $EVENTS"
ok "eventstore: session.user.checked + session.password.checked"
# The wrong attempt lands on the USER aggregate — user.human.password.check.failed
# — which is where lockout policy counts it. The audit trail has both halves.
HUMAN_ID="$($PSQL "select id from projections.users14 where username = '$HUMAN'")"
FAILED="$($PSQL "select count(*) from eventstore.events2
  where aggregate_id = '$HUMAN_ID' and event_type = 'user.human.password.check.failed'")"
[[ "$FAILED" -ge 1 ]] || die "the wrong attempt left no user.human.password.check.failed event"
ok "and the wrong attempt is on the user aggregate ($FAILED × password.check.failed) — lockout counts it"

step "10 · the other two flows start on the same engine"
for slug in login-mfa recovery; do
  C="$(curl -s -X POST "$API/flows/v1/$slug/start" | jqp '["challenge"]["component"]')"
  [[ "$C" == "tessera-stage-identify" ]] || die "$slug: $C"
  ok "$slug → $C"
done

printf '\n  one engine, three files: blueprints/dev/flows.yaml\n\n'
