# 01 — The seat token contract

**Status:** draft v1, and the thing to get right before any implementation.
**Schema:** `shippin.seat-token.v1`
**Verified by:** Automaton (`engine/serve/identity.mjs`), DevForge, and every
future consumer.

## Why this document is first

Every service in the umbrella needs one question answered the same way: *is this
caller who they say they are, and what are they allowed to do?* If each one
learns the answer from a vendor's SDK, the vendor becomes load-bearing and
swapping it is a migration across every repo.

So the boundary is a **token**, not a product. Anything that can mint one of
these correctly can be the identity provider — Authentik today, Zitadel,
something Shippin-native later. Consumers never find out which.

This is also why the contract is written before the service: Automaton is
already building the verifier, and a verifier is the most honest specification
there is. It either accepts your token or it does not.

## Shape

A seat token is a JWT signed with an asymmetric key published in the issuer's
JWKS. Consumers verify the signature and read claims; they never call the
issuer on the request path.

```json
{
  "iss": "https://id.shippin.example/",
  "sub": "mem_01J8…",
  "aud": ["automaton:ws-0001"],
  "exp": 1786930000,
  "iat": 1786929100,
  "nbf": 1786929100,
  "jti": "01J8…",

  "schema": "shippin.seat-token.v1",
  "account_id": "acc_01J8…",
  "member_id": "mem_01J8…",
  "workspace_id": "ws-0001",

  "occupant": "human",
  "basis": "subscription",

  "authorization": {
    "subject": "mem_01J8…",
    "scopes": ["hosting:active", "terminal:advanced", "chat:unified"],
    "policy_version": "pol_2026_08_17"
  },

  "provider": { "access_class": "subscription" }
}
```

A token minted for an agent seat differs in exactly two claims:

```json
{ "sub": "clyffy", "occupant": "agent", "basis": "subscription" }
```

And a **delegated** one — Clyffy acting for a customer — carries RFC 8693's
`act`:

```json
{ "sub": "mem_01J8…", "occupant": "human",
  "act": { "sub": "clyffy", "occupant": "agent" } }
```

The `authorization` and `provider` blocks are lifted verbatim from
`shippin/docs/foundation/SHARED-SEAM-DRAFT.md` rather than invented, so a token
and a seam envelope describe entitlement with the same words.

## The rules that matter

**`aud` names one workspace, and consumers must check it.** This is the whole
multi-tenant boundary in one claim: a token minted for `ws-0001` presented to
`ws-0002` is a forgery attempt, and it fails on an audience check rather than on
someone remembering to compare a tenant id later. At fifty workspaces this is
the difference between isolation and hope.

**Short `exp`.** Fifteen minutes or less. These are handoff and session tokens,
not API keys; a long-lived one is a credential in a cookie jar with no way to
take it back. Renewal is the control plane's job, not the consumer's.

**`jti` on every token.** Not used at v1, and required anyway: revocation before
`exp` needs an id to revoke, and adding one later means reissuing everything.

**Asymmetric signatures only.** Consumers hold a *public* key. `alg: none` is
refused; so is any `HS*` — a token signed with the public key as an HMAC secret
verifies against a naive implementation, and the public key is public. Allowed:
`RS256`, `ES256`. (`ES256` carries a raw `r‖s` signature, so verifiers need
`dsaEncoding: "ieee-p1363"` — without it a valid token silently fails and looks
like an issuer misconfiguration.)

**Scopes are the entitlement, and they are namespaced.**

| scope | means |
|---|---|
| `hosting:active` | a live hosting subscription — required for anything at all |
| `terminal:advanced` | the workspace terminal and multi-pane surfaces |
| `chat:unified` | the same providers as panel chat, through brokers |
| `workflows:guided` | guided workflow surfaces and deep links |

`03-control-plane-contract.md` in DevForge names these as dotted booleans
(`hosting.active`). Consumers accept both spellings; the colon form is
canonical, because it is what OAuth scope syntax and the seam draft already use.

**A seat is a licensed occupant, and the token says which kind and on what
basis.** The canonical definition is `shippin/docs/platform/NAMING-AND-SSOT.md`
§1.1; this contract is where it becomes wire format.

| claim | values | means |
|---|---|---|
| `occupant` | `human` · `agent` | who is in the seat |
| `basis` | `subscription` · `usage` · `local` · `unknown` | how its capacity is paid for |

`occupant: agent` is the customer-legible half of *non-human identity*. It is
not a lesser seat — an agent seat authenticates, carries scopes and appears in
an audit trail exactly as a human one does.

**`unknown` is never promoted to `subscription`.** A basis nobody measured is
not a subscription, and guessing is how a per-usage bill arrives that nobody
chose. Automaton enforces this on the consuming side
(`engine/abstract/seat.mjs`) and will refuse a token that asserts otherwise
without evidence.

`provider.access_class` in the seam envelope carries the same axis and the same
values; Automaton's older internal spelling (`subscription_oauth`, `api_key`)
is read as an alias so neither side has to migrate before the other.

**Delegation is `act`, and it is not impersonation.** When an agent seat acts
for a human one, `sub` stays the human and RFC 8693's `act` names the actor.
The distinction is the RFC's: under impersonation the actor is
"indistinguishable" from the subject, under delegation it "still has its own
identity". `PRODUCT-ARCHITECTURE.md` §7 requires a persistent visible indicator
and a complete audit trail for act-as-you mode — both are only possible under
the second, because an impersonated session has nothing left to indicate.
Nested `act` claims express a chain, outermost being the current actor.

**A missing entitlement is not a failed login.** Consumers answer `403` with a
typed body naming what is missing, never a bare `401`. "You are not signed in"
and "your plan does not include this" have different fixes, and a customer told
the wrong one is told nothing.

## Discovery

Standard OIDC, so nothing bespoke has to be taught to a consumer:

```
GET <issuer>/.well-known/openid-configuration   → { "jwks_uri": … }
GET <jwks_uri>                                  → { "keys": [ … ] }
```

Consumers cache the JWKS with a TTL and refetch on an unknown `kid` **at most
once per cooldown** — refetching on every miss turns key rotation into an
outbound-request amplifier that anyone holding a made-up `kid` can pull.

## Open

1. Refresh: does the control plane mint a new seat token on a schedule, or does
   the consumer redirect to a handoff again? (Automaton's Stage 2 assumes the
   second, which needs no new endpoint.)
2. Revocation transport — a published list, a short-TTL check, or `exp` alone.
3. Whether `workspace_id` and `aud` can ever disagree, and what that would mean.
4. Organizations: `account_id` is the tenant today; multi-org membership needs a
   claim before it needs an implementation.
