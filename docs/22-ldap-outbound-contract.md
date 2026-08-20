# 22 — Outbound LDAP federation contract

**Status:** accepted implementation contract; `ldap_outbound` remains preview
until the bundle-bound OpenLDAP and supported Active Directory proofs pass.
**Depends on:** `18-identity-edge-and-vaultix-contract.md` and the
`SecretReference` custody boundary.

## Resource boundary

`LDAPOutboundConnector` is Tessera's provider-neutral desired state. It owns:

- account/workspace scope, connector id, revision and display name;
- an ordered failover list of `ldap://` StartTLS or `ldaps://` endpoints;
- service bind DN plus a `ldap_bind` Tessera secret reference id;
- tenant-scoped user and group base DNs;
- closed user/group attribute mapping;
- generic, OpenLDAP or Active Directory profile;
- direct or nested group resolution with a bounded depth;
- disabled-user attribute/value or bit-mask policy;
- independent `authenticate`, `import`, `reconcile` and `deprovision` effects;
- bounded connect/search timeouts and result limits;
- bounded lifecycle page size and maximum snapshot users when lifecycle effects
  are enabled;
- optional public PEM trust anchors.

Plain LDAP without StartTLS is not representable. Endpoint URLs may not carry
user-info, paths, query strings or fragments. TLS uses the endpoint hostname,
not `host:port`, requires TLS 1.2 or newer and validates against system roots
plus the reviewed public trust anchors. Bind and search bases must parse as
LDAP distinguished names. Configurable attribute and object-class values use
the closed LDAP descriptor grammar; arbitrary filter fragments are never
accepted as configuration.

## Authentication

Authentication resolves the service bind credential only inside the Vaultix
custody callback, binds, searches beneath the configured user base using an
escaped login value, requires exactly one entry, applies disabled-user policy,
then binds as that user. Wrong password, unknown user and ambiguous user return
typed safe refusals and never include a password or raw provider response.

The caller's user password and the Vaultix bind value are copied only for the
operation and cleared on return. Neither is part of connector state, mapping
preview, health, audit or an operation envelope.

## Mapping and groups

Mapping preview returns a normalized non-persisted sample with source attribute
names, safe values and the evaluated disabled state. Required identity and
login mapping names are validated before dialing; an entry whose mapped subject
or username value is empty fails closed with `invalid_mapping`. Group resolution
searches only the configured group base, escapes every DN used in a filter,
requires every returned group to have its mapped name, deduplicates results and
caps the aggregate unique group set—not merely each individual search—by the
configured result limit. An aggregate overflow fails closed with
`result_limit_exceeded`.

Authentication, import, reconcile and deprovision are separate effects. A
successful login never implies provisioning or suspension. Deprovision cannot
be enabled without reconcile, and the emergency local-owner path is outside
this connector.

## Lifecycle reconciliation

A lifecycle run first obtains a complete, paged directory snapshot beneath the
configured user base. Every page is bounded and the run stops without producing
a plan if the configured maximum user count, mapping rules, group limits,
deadline or directory availability is violated. A partial scan is never treated
as evidence that an upstream user was removed.

The planner keys ownership by connector id plus mapped immutable subject and
compares usernames against the complete tenant target namespace, rejecting
duplicate subjects or usernames before apply. `import` permits creation of missing
connector-owned users. `reconcile` permits profile/group updates, suspends a
present but disabled directory user and reactivates an enabled user previously
suspended by the same connector. `deprovision` permits suspension—not hard
deletion—of a connector-owned user absent from a complete snapshot. Users owned
by another connector and local users, including the emergency owner, are never
changed.

Plans are sorted, content-addressed, revision-bound and short-lived. Preview has
no side effect. Apply must verify the plan hash, connector revision, expiry and
expected target revision, then commit every action atomically or none. Replaying
an already committed plan is an idempotent success. A target revision race,
tampered/expired plan or duplicate identity fails closed with a typed refusal.

## Promotion gate

The capability remains preview until one immutable suite proves:

- StartTLS and LDAPS hostname/root validation, including wrong-root negatives;
- service bind, lookup and user bind with injection-resistant filters;
- direct and nested groups plus mapping preview;
- disabled and removed user behavior;
- paged import plus deterministic create/update/reactivate/suspend previews,
  idempotent atomic apply and stale-revision rejection;
- ordered failover, timeout and dependency-loss behavior;
- cross-tenant base-DN isolation;
- a `ldap_bind` Vaultix reference with no seeded value in state, API, logs or
  evidence;
- OpenLDAP and the supported Windows Active Directory profile.
