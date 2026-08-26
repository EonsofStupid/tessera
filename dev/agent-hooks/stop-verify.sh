#!/usr/bin/env bash
# Stop gate: targeted process tests only. Fail-open if stdin is unreadable.
set -u
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"

payload="$(cat || true)"
stop_active="$(python3 -c 'import json,sys
try:
    ev=json.loads(sys.argv[1] or "{}")
except Exception:
    ev={}
print("true" if ev.get("stopHookActive") or ev.get("stop_hook_active") else "false")' "$payload" 2>/dev/null || echo false)"

log="$(mktemp)"
fail=0
{
  go test ./dev/product-language ./internal/cache/connector ./backend/v1/domain ./backend/v1/management ./cmd/setup -count=1
} >"$log" 2>&1 || fail=1

if [[ "$fail" -eq 0 && -d web/node_modules ]]; then
  (cd web && pnpm check:process) >>"$log" 2>&1 || fail=1
fi

if [[ "$fail" -eq 0 ]]; then
  rm -f "$log"
  exit 0
fi

reason="$(tail -n 20 "$log" | tr '\n' ' ' | cut -c1-800)"
rm -f "$log"
if [[ "$stop_active" == "true" ]]; then
  printf '%s\n' "Process tests still failing after a continuation. Stop and ask a human. $reason" >&2
  exit 0
fi
python3 -c 'import json,sys; print(json.dumps({"decision":"block","reason":sys.argv[1]}))' "Process tests failed. Fix credentials, edition, or landing language before finishing. $reason"
exit 0
