# 16 — Capability and compatibility discovery contract

**Status:** accepted v1 contract for P1.4.
**Builds on:** `14-operation-contract.md`.

## Purpose

`GET /tessera/v1/capabilities` is the only authority for deciding which paths
the Tessera management application offers. Neither the standalone browser nor
an optional host adapter infers support from a version string, a successful
health check, an inherited provider endpoint or a remembered deployment
profile.

This is a protected management read. Tessera requires
`tessera.capabilities.read`; missing authentication is a typed `401`, while an
authenticated caller without that permission receives a typed `403` naming
the permission. The same-origin browser uses the user's Tessera session;
external clients use a least-privilege management credential.

The response answers three different questions:

1. Which discovery schema is this response using?
2. Are Tessera and the optional operator, analytics, custody, mesh and host
   adapter components compatible?
3. Is each customer capability operational, degraded, preview-only or
   unsupported, and should its UI be enabled, disabled or hidden?

## Response

A discovery document contains:

- `schema_version`, currently `1`;
- an opaque `resource_revision` for cache/revalidation;
- `observed_at`, when the facts were assembled;
- optional SHA-256 `bundle_manifest_digest` binding a tested signed bundle;
- one fact for each relevant component;
- one fact for each advertised capability.

Component versions are opaque display values. Clients do not parse them or
construct compatibility ranges. Tessera reports the management API major and
the compatibility result already evaluated against its signed bundle policy.
This keeps compatibility policy in one tested place.

The only mandatory component role is `tessera`. `tessera_operator`,
`clickhouse`, `vaultix`, `zuul` and `shippin_adapter` are reported as
`not_present` when intentionally absent and are required only by capabilities
that name them. Vaultix is a custody dependency, ClickHouse is an asynchronous
analytics projection, and neither becomes an identity authority.

Component compatibility is one of:

- `compatible` — this combination is within the tested policy;
- `incompatible` — installed versions are known not to work together;
- `not_present` — the optional component is not installed;
- `unknown` — evidence is insufficient, so dependent mutations remain off.

## Capability facts

A capability has a stable id, support status, UI exposure, safe reason,
required components and supported operation kinds.

The proof contains `conformance_id`, `bundle_manifest_digest`, `result`,
`verified_at` and `evidence_digest`. Raw evidence remains in the protected
release-evidence store; discovery exposes only the immutable digest and safe
status needed to prevent an unproved UI claim.

Support status is `unsupported`, `preview`, `operational` or `degraded`. UI
exposure is `hidden`, `disabled` or `enabled`. The server decides both; clients
may demote exposure for safety but never promote it.

Required invariants:

- `unsupported` is never enabled;
- `preview` is disabled until a later explicit preview-enrollment contract;
- `operational` and `degraded` require passing conformance evidence bound to
  the advertised bundle manifest digest;
- hidden or disabled capabilities carry a safe reason;
- every required component must be present and `compatible` before a client
  may honor `enabled`;
- operation kinds use the P1.2 vocabulary and are unique;
- missing capability ids mean `hidden/not_advertised`, never “probably
  supported.”

Initial stable ids include `overview`, `guided_setup`,
`deployment_operations`, `analytics_olap`, `upstream_oidc`, `upstream_saml`,
`downstream_oidc`, `downstream_saml`, `ldap_outbound`, `ldap_inbound`,
`forward_auth`, `identity_aware_proxy`, `visual_flow_engine`, and
`vaultix_secret_custody`. New ids are additive.

## Client resolution

For each route or action, every Tessera client resolves in this order:

1. An unsupported `schema_version` disables it as `schema_incompatible`.
2. An invalid document using a supported schema disables it as
   `invalid_discovery`.
3. An absent capability is hidden as `not_advertised`.
4. A server-hidden capability remains hidden.
5. A missing, unknown or incompatible required component disables it as
   `component_unavailable`.
6. Otherwise the server-provided exposure is used unchanged.

An older client ignores unknown capabilities. A newer client facing an older
server treats absent capabilities as hidden. Neither case guesses support.

Discovery may be cached by revision for a short server-declared lifetime, but
the adapter revalidates before plan/apply for a mutation. A cached enabled
button does not override an apply-time compatibility or revision refusal.

## Signed bundle boundary

P1.4 defines the fact shape, not signature transport. P3.3 creates signed image
and bundle manifests; P9.4 adds upgrade policy. When a bundle digest is present
it uses `sha256:<lowercase hex>` and the component compatibility facts must have
been evaluated from that exact manifest.

Ad hoc or developer builds may omit the digest. They report compatibility as
`unknown` unless an explicit development policy provides evidence. The UI
labels that state; it does not present it as a production-tested bundle.

## Done when

- invalid documents and duplicate component/capability facts are rejected;
- tests cover absent, hidden, incompatible, unknown and supported paths;
- no client can promote a server-disabled or unsupported capability;
- the domain and protobuf vocabularies agree;
- the schema remains provider-neutral and contains no infrastructure
  inventory or secret material.
