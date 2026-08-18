# 05 — Minting a seat token

**Status:** done and reproducible. Tessera mints `shippin.seat-token.v1` and
Automaton's verifier accepts it.

`04-building-the-trunk.md` got the trunk to serve OIDC. This is the path from
there to a token a consumer will take, and the three places it does not work the
way you would guess.

## The whole path

```bash
bash dev/up.sh                       # postgres, init, setup, serve
PAT=$(cat .artifacts/admin.pat)      # gitignored; reissued whenever .pgdata is
```

A seat, as a machine user whose access tokens are JWTs:

```bash
curl -X POST localhost:8088/management/v1/users/machine \
  -H "Authorization: Bearer $PAT" -H 'Content-Type: application/json' \
  -d '{"userName":"seat-probe","name":"Seat probe",
       "accessTokenType":"ACCESS_TOKEN_TYPE_JWT"}'
```

Its facts, as user metadata (values base64, keys namespaced):

| key | value |
|---|---|
| `shippin:seat:workspaces` | `ws-0001 ws-0002` — which it may occupy |
| `shippin:seat:occupant` | `human` · `agent` |
| `shippin:seat:basis` | `subscription` · `usage` · `local` · `unknown` |
| `shippin:entitlement:scopes` | `hosting.active terminal:advanced chat.unified` |
| `shippin:entitlement:policy_version` | `pol_2026_08_17` |

Then a secret, and the token:

```bash
curl -X POST localhost:8088/oauth/v2/token -u "$CLIENT_ID:$SECRET" \
  -d grant_type=client_credentials \
  --data-urlencode 'scope=openid urn:shippin:audience:automaton:ws-0001'
```

```json
{ "iss": "http://localhost:8088", "sub": "3867…", "aud": ["automaton:ws-0001"],
  "exp": …, "jti": "V2_…",
  "schema": "shippin.seat-token.v1",
  "account_id": "3867…", "member_id": "3867…", "workspace_id": "ws-0001",
  "occupant": "agent", "basis": "subscription",
  "authorization": { "subject": "3867…", "policy_version": "pol_2026_08_17",
                     "scopes": ["chat:unified","hosting:active","terminal:advanced"] },
  "provider": { "access_class": "subscription" } }
```

Asking for a workspace this seat does not occupy is refused, and says so:

```json
{ "error": "invalid_target",
  "error_description": "seat: this member does not occupy that workspace: ws-0009 is not among \"ws-0001 ws-0002\"" }
```

## Where it lives

`backend/v1/domain` is the claim set and the rules, and imports nothing from
OIDC or the eventstore — the contract is the product boundary, and a boundary you can
only exercise by standing up a provider is one nobody exercises.
`internal/api/oidc/seat_claims.go` gathers the facts; `createJWT` calls it once,
immediately before the signature, so there is exactly one place to look for what
a consumer will see.

## The three traps

**A scope Tessera does not declare is not merely ignored — it is erased from the
audience too.** `op.ValidateAuthReqScopes` drops unrecognised scopes with
`slices.DeleteFunc`, which compacts the caller's backing array *in place*. The
client-credentials and jwt-profile paths then build their audience from
`r.Data.Scope` — the same array, after that compaction. So an undeclared
`urn:shippin:audience:…` does not fail: the token mints successfully, without
the workspace, and `aud` quietly falls back to the client id. It cost an hour.
`isScopeAllowed` in `client_converter.go` is where a Tessera scope gets declared,
and it is not optional.

**The token lifetime has exactly one key, and the two obvious ones both fail
silently.** `OIDC.DefaultAccessTokenLifetime` is read and accepted and changes
nothing — `setup` writes the instance's own OIDC settings, and once written they
override the runtime defaults for every token the instance ever mints.
`FirstInstance.OIDCSettings` in a steps file is worse: that struct has no such
field, so it parses cleanly, matches nothing, and reports nothing. The key that
works is `DefaultInstance.OIDCSettings` in the config passed to `--config`, with
all four durations given together or setup fails.

Both wrong answers were written down as correct here before the clean-slate run
caught them, which is the point: a 12-hour seat token looks exactly like a
15-minute one until you subtract `iat` from `exp`. `dev/seat-probe.sh` prints
the lifetime for that reason.

**An arbitrary audience cannot be requested at token exchange.** RFC 8693
refuses any audience not already in the subject token, so the workspace has to
enter at the *first* mint, from a scope. Reaching for token exchange first is
the natural move and it is a dead end.

## What proves it

`node dev/verify-seat-token.mjs` runs Automaton's real verifier
(`../automaton/engine/serve/identity.mjs`) against live tokens: it accepts one
minted for `ws-0001`, refuses one minted for `ws-0002` on the audience check,
refuses a payload tampered to add `hosting:unlimited`, and refuses the same
claims re-headed as `alg: none`.

A fake issuer would have proved none of it. The consumer was already written,
and pointing it at this is the only test that could fail for a real reason.
