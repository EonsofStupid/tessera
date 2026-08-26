# 17 — Management overview projection contract

**Status:** accepted v1 contract for P4.1.
**Builds on:** `15-management-error-contract.md` and
`16-capability-discovery-contract.md`.

## Purpose

`GET /nomen/v1/overview` is the first provider-neutral read model consumed by
the standalone Nomen management application. It summarizes identity facts
Nomen owns; it does not expose an inherited administration response,
infrastructure inventory, billing policy, credentials or raw audit payloads.

The same-origin browser calls this endpoint with the user's Nomen session.
External panels may call through a server-side adapter with a least-privilege
credential. Nomen requires `nomen.overview.read` in either case. Management
permissions use dot-separated action names because a colon is reserved for an
optional resource context suffix. Missing authentication is typed `401`;
missing permission is typed `403` and names that permission.

## Wire response

The management API uses snake-case JSON. Nomen's data client validates and
maps this wire shape into typed view models. Optional adapters may map it into
their host convention without changing the Nomen contract.

```json
{
  "schema_version": 1,
  "service_id": "nomen",
  "resource_revision": "sha256:<lowercase hex>",
  "observed_at": "2026-08-19T00:00:00Z",
  "readiness": {
    "status": "ready",
    "issuer": "https://id.nomen.test",
    "signing_keys": 1,
    "flows": 3,
    "policy_revision": "sha256:<lowercase hex>",
    "reasons": []
  },
  "lenses": [
    {
      "id": "infrastructure",
      "label": "Infrastructure",
      "value": 7,
      "unit": "attachments",
      "detail": "Workspace identity attachments managed by Nomen.",
      "status": "ready"
    },
    {
      "id": "ai",
      "label": "AI",
      "value": 4,
      "unit": "agent seats",
      "detail": "Agent seats with explicit identity and scope.",
      "status": "ready"
    },
    {
      "id": "customers",
      "label": "Customers",
      "value": 18,
      "unit": "human seats",
      "detail": "Human seats managed by Nomen.",
      "status": "ready"
    }
  ],
  "federation": { "upstreams": [], "clients": [] },
  "activity": []
}
```

Federation and activity are present even when empty. P4.1 does not invent
provider health or audit events merely to fill the dashboard; later read
resources populate those arrays from their own projections.

## Source and readiness rules

- `resource_revision` is a digest of the facts rendered, not a build version or
  current timestamp. The same source facts produce the same revision.
- `observed_at` is when the projection was assembled. It is never used as the
  resource revision.
- `policy_revision` is a digest of the distinct seat policy revisions in the
  instance. Missing or mixed policy facts therefore cannot masquerade as one
  named live policy.
- `signing_keys` counts usable active asymmetric signing keys. A shared-secret
  verifier is not counted.
- `ready` requires at least one usable signing key and one configured Nomen
  flow. Otherwise status is `degraded` and `reasons` names every missing fact.
- A storage or signing-key read failure returns typed `503 service_unavailable`
  with a safe diagnostic reference. It never returns a healthy-looking zero.
- Lens values come from Nomen-owned seat and tenant-attachment tables. They
  do not claim host infrastructure inventory or billing truth.

## Failure and caching

Successful responses are member-sensitive and use `Cache-Control: private,
no-store`. Typed errors follow contract 15 and are also non-cacheable. The
handler emits JSON only, rejects non-GET methods through routing, and never
serializes upstream errors, SQL, secrets, tokens or stack traces.

The Nomen browser client has a bounded timeout and performs no automatic
retry for a user's mutation. A timeout or malformed success renders the typed
service-unavailable remedy; `401` and `403` preserve their distinct structured
meaning. External adapters follow the same rule.

## Done when

- domain validation rejects an incomplete or falsely-ready overview;
- a storage-backed source reports seats, attachments, flows and policy facts;
- the HTTP handler enforces `nomen.overview.read` and emits typed failures;
- the running Nomen router owns `/nomen/v1/overview`;
- standalone browser fixtures and a live Nomen handler satisfy the same
  mapping tests; optional adapters prove the same wire mapping independently.
