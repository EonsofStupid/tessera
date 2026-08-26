# 15 — Typed management error contract

**Status:** accepted v1 contract for P1.3.
**Builds on:** `01-seat-token-contract.md` and
`14-operation-contract.md`.

## Why this contract exists

The Nomen management application must tell a user what happened and what
safe action fixes it. A status code alone cannot do that. Every expected
management refusal uses
the provider-neutral envelope below; inherited provider errors are translated
at the Nomen boundary and never passed through to the browser.

Authentication and authorization remain deliberately separate:

- `401 authentication_required` means no usable member authentication was
  presented. The remedy is to sign in.
- `403 entitlement_required` means the member is signed in but lacks a product
  entitlement. The response names the missing entitlement.
- `403 permission_required` means the member is entitled but lacks delegated
  authority for this action. The response names the required permission.
- `403 step_up_required` means the member is allowed to act after satisfying a
  stronger authentication requirement. It is not reported as a failed login.

A missing entitlement is never collapsed into `401`.

## Envelope

```json
{
  "error": {
    "type": "entitlement_required",
    "reason": "missing_entitlement",
  "message": "This tenant does not include Nomen administration.",
    "remedy": {
      "kind": "request_entitlement",
      "label": "Request access"
    },
    "retry": "operator_action",
    "missing_entitlement": "nomen:manage",
    "diagnostic_ref": "diag_example"
  }
}
```

`type` and `reason` are stable machine values. `message` and `remedy.label` are
display text and may be localized. A client branches on `type` or structured
detail, never on display text.

`diagnostic_ref` is an opaque support correlation value. It must not contain a
  token, secret, provider assertion, user email, raw upstream response or stack
trace. Expected refusals do not expose inherited provider error identifiers.

## Catalog and transport mapping

| type | HTTP | gRPC | retry | required detail | default remedy |
|---|---:|---|---|---|---|
| `authentication_required` | 401 | `UNAUTHENTICATED` | `operator_action` | — | `sign_in` |
| `entitlement_required` | 403 | `PERMISSION_DENIED` | `operator_action` | `missing_entitlement` | `request_entitlement` |
| `permission_required` | 403 | `PERMISSION_DENIED` | `operator_action` | `required_permission` | `request_permission` |
| `step_up_required` | 403 | `PERMISSION_DENIED` | `operator_action` | `required_assurance` | `step_up` |
| `conflict` | 409 | `ABORTED` | `replan` | — | `refresh_and_review` |
| `invalid_request` | 422 | `INVALID_ARGUMENT` | `operator_action` | `field` when one field caused it | `correct_request` |
| `rate_limited` | 429 | `RESOURCE_EXHAUSTED` | `same_request` | positive `retry_after_seconds` | `retry_after` |
| `service_unavailable` | 503 | `UNAVAILABLE` | `same_request` or `operator_action` | safe `diagnostic_ref` | `retry_later` |

The HTTP response also carries the normal protocol headers:

- `401` includes a suitable `WWW-Authenticate` challenge;
- `429` includes `Retry-After`, equal to `retry_after_seconds`;
- cache headers prevent storage of member-specific errors.

The gRPC status uses the listed canonical code and attaches the management
error message as typed details. REST and gRPC therefore carry the same `type`,
`reason`, remedy and structured fields.

## Detail-field rules

Only fields relevant to the selected type are populated:

- `missing_entitlement` is a Nomen scope identifier, not a pricing or plan
  decision;
- `required_permission` is the delegated permission needed by the action;
- `required_assurance` is a stable assurance or step-up policy identifier;
- `field` is a public request field path and never a database column;
- `current_revision` may accompany a conflict so the client can replan, but the
  server never applies against it automatically;
- `retry_after_seconds` is positive and is used only for throttling;
- `diagnostic_ref` is safe to display and share with a private-community
  operator.

Unknown fields are ignored by clients. Unknown `type` values render the generic
service-unavailable experience and preserve the diagnostic reference; they do
not guess that retrying is safe.

## Mapping operation refusals

The P1.2 operation reasons map as follows:

- malformed plan or idempotency input → `422 invalid_request`;
- expired plans, digest mismatch, stale revisions, idempotency reuse and writes
  after terminal completion → `409 conflict`;
- corrupt server-owned progress sequence or phase →
  `503 service_unavailable` with operator action required.

Authorization and prerequisite checks may return their own catalog type before
an operation is created. Provider outages become `service_unavailable`; their
raw response body is retained only in protected operator diagnostics.

## Nomen rendering contract

`testdata/nomen/management-error-remedies.json` is the standalone browser
fixture. Each catalog type has a distinct consequence-first title, explanation
and primary action. The application may restyle or localize these values, but
it must preserve the action kind and required structured detail. Optional host
adapters must satisfy the same fixture contract.

Browser code receives only the envelope. Operator credentials, protected
secret references, upstream session material and raw provider errors remain in
Nomen's server-side management boundary.

## Done when

- every catalog type has an exact HTTP and gRPC mapping;
- validation rejects a missing type-specific detail or an unsafe retry shape;
- every P1.2 refusal reason maps deterministically;
- the domain and protobuf vocabulary agree;
- Nomen browser fixtures render a distinct remedy for every catalog type;
- tests prove a missing entitlement is typed `403`, never `401`.
