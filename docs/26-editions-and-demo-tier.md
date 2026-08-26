# 26 — Editions: public, demo, and enterprise

**Status:** accepted product split. Implementation follows this file.
**Product boundary:** `02-standalone-product-contract.md`.

Nomen ships **one binary** from the same contracts. Distribution may use two
OCI tags (`nomen:public`, `nomen:enterprise`) that differ only by default
`NOMEN_EDITION`. A hosted **demo** is the public edition with caps, not a
third codebase.

## Why two editions

A single “everything on” image cannot deploy for a stranger and cannot be sold.
Public must boot on empty PostgreSQL with no Redis, no Vault, and no Mesh.
Enterprise withholds those until the operator is on a paid path.

## Editions

| Edition | Who | What is live | What is withheld |
|---|---|---|---|
| **public** | self-host community | identity, first owner, OIDC/SAML preview, recreatable Postgres tables, PostgreSQL cache | Redis, Nomen Vault adapter, Nomen Mesh adapter, ClickHouse, high availability |
| **enterprise** | paid nomen.sh and licensed self-host | public plus Redis cache, Vault trust, Mesh trust, HA and support terms | nothing in the identity core; Vault and Mesh remain sibling instances |

`NOMEN_EDITION` is `public` or `enterprise`. Default is `public`. Redis
`Caches.Connectors.Redis.Enabled` is refused unless the edition is
`enterprise`. A missing withheld capability is a typed `403` with
`type=entitlement_required`, `reason=edition_public_withheld`, and
`missing_entitlement` naming the capability. It is never a bare `401`.

Wire component roles `vaultix` and `zuul` remain historical ids in the
discovery document (`16-capability-discovery-contract.md`). Product language
is Nomen Vault and Nomen Mesh. Public reports those components
`not_present` with reason `edition_public_withheld`.

## Free demo (hosted public)

Set `NOMEN_DEMO_CAPS=true` on nomen.sh free hosting only. Self-host public
has **no** user cap.

The hosted demo is public edition with these caps, enough to sign in, create
the first owner, and see the landing page:

- 1 instance
- 1 organization
- 25 humans
- PostgreSQL only
- no Redis
- Vault and Mesh named as products that trust this issuer, adapters off
- no tenant counts or audit exports on the public page

Exceeding a demo cap is a typed `403` with `type=entitlement_required`,
`reason=demo_cap_exceeded`, and `missing_entitlement` of
`nomen.demo.instance`, `nomen.demo.organization`, or `nomen.demo.user`.
The first instance, organization, and owner created at setup are allowed;
the cap applies to later mutations on the running process.

The demo is not a second IAM. Turning `NOMEN_DEMO_CAPS` off and setting
`NOMEN_EDITION=enterprise` is the paid path.

## First account

The first human is created **at setup on that deployment**. Git never contains
a username, password, recovery artifact, or PAT. Setup steps files must not
contain a password field value; if one is present, setup refuses.

Supply at deploy time, or skip and enroll later:

- `NOMEN_FIRSTINSTANCE_ORG_HUMAN_USERNAME`
- `NOMEN_FIRSTINSTANCE_ORG_HUMAN_PASSWORD` (env or secret injection only)
- display name and email through the same `NOMEN_FIRSTINSTANCE_ORG_HUMAN_*` keys

If the password env is unset, setup does not create a human. The operator
uses the first-owner passkey ceremony (`25-deployment-preflight-contract.md`).
Either way, the account is the first created on that database, not a baked-in
founder.

## Vault and Mesh

Nomen Vault (secrets) and Nomen Mesh (network) **trust this issuer**. They do
not get `nomen_vault_*` or `nomen_mesh_*` tables in this database. Enterprise
enables the trust adapters; public only names the products.

## Public environment document

`GET /ui/console/assets/environment.json` may include `edition` (`public` or
`enterprise`) and `demo_caps` (boolean). Those fields are not tenant counts.
