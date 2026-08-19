# 20 — Capability gap ledger

**Status:** evidence-based baseline for the child implementation branches.
**Measured at:** `agent/tessera-gap-foundation` fork point
`bf01af99ebd0a5843731e9f7d8b62f4b53b4b6f1`.

## Why this exists

The source tree contains substantial inherited and Tessera-native foundations,
but foundation is not finished product behavior. This ledger prevents source
presence from being confused with a supported capability and identifies the
smallest honest branch that advances each gap.

Statuses are:

- **foundation** — useful runtime code exists but the Tessera contract,
  management lifecycle or conformance proof is incomplete;
- **absent** — no Tessera-owned implementation exists in the expected boundary;
- **contracted** — provider-neutral contract exists but runtime behavior does
  not;
- **verified** — bundle-bound conformance evidence exists. None of the six gap
  capabilities is verified at this baseline.

## Baseline

| capability | current evidence | status | missing before implementation proof |
|---|---|---|---|
| LDAP outbound | `internal/idp/providers/ldap/` contains an inherited LDAP provider with TLS, bind/search and mapping code; inherited admin/management endpoints configure it. | foundation | provider-neutral Tessera resource/API; Vaultix secret reference; explicit authentication versus provisioning modes; mapping preview; group/deprovision behavior; AD profile; safe health/audit; Linux and Windows conformance |
| LDAP inbound | no `internal/edge/ldap/` or Tessera LDAP server package exists | absent | tenant-scoped virtual directory model; bind/search policy; TLS and certificate custody; revocation propagation; edge supervision/configuration; standard-client conformance |
| Forward auth | no `internal/edge/proxy/` or forward-auth Tessera adapter exists | absent | signed/versioned edge configuration; request decision endpoint; header allowlist and spoof stripping; original-target contract; revocation and fail-closed behavior; proxy conformance |
| Identity-aware proxy | no Tessera proxy edge exists | absent | browser session/callback model; cookie and CSRF policy; upstream allowlist; websocket behavior; logout/revocation; independent supervision; proxy conformance |
| Visual flow engine | `backend/v1/domain/flow*.go`, `backend/v1/storage/flow/` and `internal/api/flows/` implement a linear stage executor and three blueprint-defined flows | foundation | canonical graph schema; branches and terminal outcomes; dependency-failure modeling; simulation; version/diff/publish/rollback; management contract; Shippin visual editor; accessibility and stale-revision conformance |
| Vaultix secret custody | component/capability vocabulary and contract 18 exist; no Tessera `SecretReference` domain type or Vaultix adapter exists | contracted | opaque reference schema; purpose/tenant binding; short-lived workload auth; write-only enrollment; resolve/use boundary; rotation/expiry/revocation events; fake then sandbox adapter; exhaustive leak tests |

## Cross-cutting foundation already present

- `CapabilityProof` rejects available/degraded facts without a passing proof
  whose bundle digest matches the installed bundle.
- Operation, management-error, lifecycle and capability contracts provide the
  plan/apply/verify and typed-failure grammar.
- The blueprint reconciler and linear flow executor are working foundations,
  not replacements for the missing desired-state and graph behavior.
- The private-cloud container generates protocol stubs and builds a non-root
  static image.
- Product-language CI prevents ordinary CLI/help/log surfaces from reverting
  to inherited branding.

## Child branch sequence and exit evidence

| branch | scope | exit evidence |
|---|---|---|
| `agent/tessera-vaultix-reference` | `SecretReference`, custody port, fake adapter and leak-safe operation envelope | unit/contract tests plus seeded-value scan; no real Vaultix credential |
| `agent/tessera-vaultix-sandbox` | native Vaultix adapter and workload policy against a dedicated sandbox | denial, expiry, rotation, outage and audit-correlation evidence |
| `agent/tessera-flow-graph` | canonical graph, validation, simulator and revision semantics | round-trip, unsafe-graph, all-terminal simulation and stale-publish tests |
| `agent/tessera-ldap-outbound` | normalize inherited provider behind Tessera resources and Vaultix binding | OpenLDAP suite, then Windows AD suite; authentication/provisioning separation |
| `agent/tessera-edge-contract` | signed/versioned edge config, session and health contract | contract tests shared by all portable edges |
| `agent/tessera-ldap-inbound` | virtual directory edge | standard client, tenant isolation, revoke and outage suites |
| `agent/tessera-forward-auth` | reverse-proxy decision adapter | header, target, tenant, revoke and fail-closed suites |
| `agent/tessera-identity-proxy` | browser-session proxy adapter | callback, cookie, CSRF, redirect, upstream and websocket suites |
| Shippin child branch | visual editor and outcome-first guides | browser/a11y tests consuming the canonical Tessera schema only |

Each branch changes its contract first, implements one boundary, adds negative
tests, and leaves the capability preview until the release proof is produced.
