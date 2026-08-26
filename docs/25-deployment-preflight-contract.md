# 25 — Deployment preflight and owner-enrollment contract

**Status:** accepted v1 implementation contract for G2.
**Builds on:** `02-standalone-product-contract.md`,
`13-deployment-lifecycle-contract.md`, and
`21-first-iam-vertical-slice.md`.

## Outcome

An operator can inspect whether a Nomen deployment is safe to initialize or
continue without changing state. After initialization, a one-time protected
bootstrap authority can enroll the first owner with a passkey and confirm an
independent recovery artifact. Neither resource is a public setup endpoint and
neither returns a credential, private key, database value or provider secret.

## Deployment preflight resource

`GET /nomen/v1/deployment/preflight` returns a versioned read-only document.
It requires `nomen.deployment.preflight.read`. Missing authentication is the
ordinary typed `401`; an authenticated identity missing the permission receives
a typed `403` naming that permission.

The response contains:

- `schema_version`, currently `1`;
- a deterministic `resource_revision` over the safe check results;
- `observed_at`;
- overall `status`: `ready`, `action_required`, or `blocked`;
- the absolute public `issuer` observed for this request; and
- exactly five checks in canonical order: `database`, `issuer`, `tls`,
  `asymmetric_signing`, and `notification_delivery`.

Each check contains `id`, `status`, `required`, `summary`, and optional
`reason`, `remediation`, and safe `diagnostic_ref`. Check status is `passed`,
`warning`, or `failed`.

The database, issuer, TLS and asymmetric-signing checks are required. Email
delivery is a warning during first-owner enrollment but becomes required before
member invitation can become operational. `ready` means all required checks
pass and no warning remains; `action_required` means required checks pass but a
warning remains; `blocked` means at least one required check failed.

An HTTPS issuer passes issuer and TLS policy only when its hostname is also a
valid WebAuthn relying-party ID. IP-literal issuers fail the required issuer
check even on loopback because browsers refuse them as WebAuthn RP IDs. Plain
HTTP is accepted only for `localhost` or a `.localhost` hostname and produces a
TLS warning; it never counts as production evidence. Plain HTTP on any other
host fails closed. Issuer evaluation rejects userinfo, query strings and
fragments.

Probe failures are represented as failed or warning checks with allowlisted
diagnostic references. Raw database, key, SMTP or network errors never cross
the API.

## First-owner enrollment resources

Owner enrollment is a resumable server-owned ceremony:

- `GET /nomen/v1/deployment/owner-enrollment` reads state;
- `POST /nomen/v1/deployment/owner-enrollment:begin` starts or resumes a
  short-lived WebAuthn registration ceremony;
- `POST /nomen/v1/deployment/owner-enrollment:complete` verifies the browser
  response and records the passkey; and
- `POST /nomen/v1/deployment/owner-enrollment/recovery:confirm` records that
  the operator exported and verified the independently protected recovery
  artifact.

Before an owner exists, these resources require a single-use bootstrap
authority supplied through a protected deployment channel. It is never placed
in a URL, image, repository, browser storage, log, event payload or response.
After enrollment completes, it is irreversibly consumed and ordinary owner
authorization governs reads and recovery changes.

Before completion, the GET resource also accepts the bootstrap authorization
scheme so a replaced browser process can discover `passkey_pending` or
`recovery_pending` without an ordinary owner session. A bootstrap-authorized
retry of `:begin` with a new idempotency key resumes the persisted owner and
deterministically reconstructs the same challenge options; it does not replace
owner identity or advance revision. Reusing the original idempotency key with
a different payload is still refused. After `complete`, bootstrap-authorized
reads and writes return the typed authentication failure.

The container receives that authority as the `NOMEN_BOOTSTRAP_AUTHORITY`
runtime secret. Nomen refuses bootstrap mutations when it is absent or
shorter than 32 bytes. A client presents it only as
`Authorization: Bootstrap <authority>` over an allowed HTTPS origin. Comparison
uses SHA-256 digests and constant-time equality. The authority value and digest
are process-only: neither is persisted. Completion state is the durable
consumption record, so restoring or retaining the deployment secret cannot
reopen bootstrap enrollment.

`POST .../owner-enrollment:begin` also requires an `Idempotency-Key` header of
16 through 200 visible ASCII characters and this strict JSON body:

```json
{
  "owner_id": "019...",
  "username": "owner@example.test",
  "display_name": "First Owner"
}
```

The response is `201` for a new ceremony or `200` for an idempotent resume and
contains `enrollment` plus the standards-shaped WebAuthn `publicKey` creation
options. The relying-party ID is the issuer hostname, the allowed origin is the
issuer origin, attestation is `none`, resident-key preference is `preferred`,
and user verification is `required`. Owner, username and display name are
bounded UTF-8 strings; unknown JSON fields and trailing JSON values are refused.

`POST .../owner-enrollment:complete` accepts `ceremony_id` and the browser's
JSON-encoded `PublicKeyCredential`. Nomen first binds the returned challenge
to the persisted SHA-256 digest, then performs the complete WebAuthn origin,
RP-ID, user-presence, user-verification, attestation and public-key validation.
On success it stores only credential ID, COSE public key, sign counter, AAGUID,
attestation type, transports and flags. It returns `enrollment` and a newly
generated `recovery_artifact` exactly once. Replays never return the artifact.

The recovery artifact is 256 random bits encoded as
`nomen-recovery-v1.<base64url>`. Only its SHA-256 digest is persisted.
`POST .../owner-enrollment/recovery:confirm` accepts the same `ceremony_id` and
the artifact value, compares its digest in constant time, and returns the
completed enrollment view. Bodies are capped at 128 KiB and all three mutation
responses use `Cache-Control: private, no-store`.
If the browser is replaced during `recovery_pending`, Nomen does not recreate
or redisplay the artifact. The operator resumes by presenting the independently
exported copy; losing it requires an explicitly reviewed recovery/reset flow.

The persisted states are `pending`, `passkey_pending`, `recovery_pending`, and
`complete`. State includes only ceremony identifiers, challenge digests,
expiry, WebAuthn credential public material, recovery-artifact digest, attempt
metadata and audit correlation. Raw challenges may live only in short-lived
protected server session storage; private keys, recovery contents and bootstrap
authority values are never stored in this table.

The protected read returns `schema_version`, deterministic
`resource_revision`, `observed_at`, `state`, optional `ceremony_id`, optional
`owner_id`, `passkey_enrolled`, `recovery_confirmed`, optional `expires_at`, and
`revision`. It never returns the challenge digest, idempotency digest, request
digest, credential reference or bootstrap-authority material. Before a ceremony
exists the state is `pending` with revision zero.

Completing passkey registration is not enough to reach `complete`. Recovery
confirmation must bind the exported artifact digest to the same ceremony. A
retry with the same idempotency key returns the existing outcome; a different
payload under the same key is refused. Expired, replayed, cross-instance or
out-of-order ceremonies fail closed with typed errors.

## UI and evidence

The standalone application renders the five preflight checks, their exact
remediation and current observation time. It does not infer readiness from
`/debug/ready`. Owner enrollment uses the browser WebAuthn API, explains the
relying-party identity before prompting, and never substitutes a mocked
credential in production.

The `bootstrap-owner` Playwright project uses Chromium's virtual authenticator
against the production image and disposable PostgreSQL. It proves interruption
and resume, challenge expiry, replay refusal, bootstrap-authority consumption,
recovery confirmation, restart persistence and absence of credential material
from logs and operator events.

Both `deployment_preflight` and `owner_enrollment` remain `preview` until their
release-bound evidence manifests pass. A successful API response or browser
journey alone cannot promote either capability.
