# 17 — Management overview projection contract

**Status:** accepted v1 contract for P4.1.
**Builds on:** `15-management-error-contract.md` and
`16-capability-discovery-contract.md`.

## Purpose

`GET /tessera/v1/overview` is the first provider-neutral read model consumed by
the Shippin server-side adapter. It summarizes identity facts Tessera owns; it
does not expose an inherited administration response, infrastructure inventory,
billing policy, credentials or raw audit payloads.

The browser never calls this endpoint directly. Shippin authenticates the
member, checks product entitlement and calls Tessera with a protected
server-side credential. Tessera separately requires
`tessera:overview:read`. Missing authentication is typed `401`; missing
permission is typed `403` and names that permission.

## Wire response

The management API uses snake-case JSON. The Shippin adapter deliberately maps
this into its browser-facing camel-case projection rather than leaking the
Tessera wire shape into components.

```json
{
  "schema_version": 1,
  "service_id": "shippin.tessera",
  "resource_revision": "sha256:<lowercase hex>",
  "observed_at": "2026-08-19T00:00:00Z",
  "readiness": {
    "status": "ready",
    "issuer": "https://id.shippin.ai",
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
      "detail": "Workspace identity attachments managed by Tessera.",
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
      "detail": "Human seats managed by Tessera.",
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
- `ready` requires at least one usable signing key and one configured Tessera
  flow. Otherwise status is `degraded` and `reasons` names every missing fact.
- A storage or signing-key read failure returns typed `503 service_unavailable`
  with a safe diagnostic reference. It never returns a healthy-looking zero.
- Lens values come from Tessera-owned seat and workspace-attachment tables.
  They do not claim Shippin infrastructure inventory or billing truth.

## Failure and caching

Successful responses are member-sensitive and use `Cache-Control: private,
no-store`. Typed errors follow contract 15 and are also non-cacheable. The
handler emits JSON only, rejects non-GET methods through routing, and never
serializes upstream errors, SQL, secrets, tokens or stack traces.

The Shippin adapter has a bounded timeout and performs no automatic retry for
the member request. A timeout or malformed success becomes Shippin's matching
typed `503`; `401` and `403` pass through their stable structured meaning.

## Done when

- domain validation rejects an incomplete or falsely-ready overview;
- a storage-backed source reports seats, attachments, flows and policy facts;
- the HTTP handler enforces `tessera:overview:read` and emits typed failures;
- the running Tessera router owns `/tessera/v1/overview`;
- Shippin adapter fixtures and a live Tessera handler satisfy the same mapping
  tests without changing the panel component model.
