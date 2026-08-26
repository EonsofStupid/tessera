# 20 — Capability delivery standard

**Status:** accepted engineering contract.

## Purpose

Nomen is delivered as complete IAM capabilities, not as screens, handlers,
database tables or demonstrations. A capability is the smallest independently
valuable identity outcome an operator can configure, a user or workload can
exercise, and Nomen can explain and recover.

Starting a container proves packaging only. Rendering a control proves
presentation only. Compiling inherited code proves availability only. None of
those facts promotes an IAM capability.

## Capability maturity

Every capability has one recorded state:

1. `inventory` — implementation material exists but has not passed Nomen's
   product, security and conformance gates;
2. `preview` — reachable only in an explicitly non-production profile and
   failed closed everywhere else;
3. `operational` — every required proof below passes for the declared release
   and deployment profile; or
4. `degraded` — an operational capability has lost a named runtime dependency
   or evidence condition and exposes an exact remedy.

There is no state for "mostly working." A capability does not inherit an
operational claim from another product, an upstream test suite or a successful
manual demonstration.

## Definition of operational

A capability is operational only when one versioned evidence manifest binds
all of the following to the release image digest and source revision:

| Layer | Required proof |
|---|---|
| Product contract | actors, outcome, terminology, defaults, limits and failure behavior are documented before implementation |
| Threat model | assets, trust boundaries, abuse cases, downgrade paths and recovery assumptions are reviewed |
| Persistence | migrations, constraints, tenant ownership, retention and rollback behavior pass against supported PostgreSQL versions |
| Service implementation | commands and queries are idempotent where required, validate all input, and return typed errors |
| Authorization | human and machine actors receive least privilege; missing entitlement returns typed `403`; cross-tenant requests fail closed |
| Protocol behavior | applicable standards and real-client/server conformance suites pass, including negative and interoperability cases |
| Operator API | the complete workflow is possible through stable versioned resources without browser-only authority |
| Guided UI | every operator step is usable at Nomen's origin, explains terms in context, offers safe seed suggestions, and never invents state |
| Clyffy boundary | discoverable plan/apply/verify actions use the same authorization and commands as a human and expose consequence and evidence |
| Audit | intent, actor, tenant, decision, target, result and correlation are recorded without credentials or protected payloads |
| Browser evidence | Playwright drives the production build against a real Nomen container and PostgreSQL; mocked success cannot promote a capability |
| Failure evidence | dependency loss, retry, concurrency, restart, revocation, rollback and abuse cases preserve the stated security outcome |
| Operations | health, metrics, logs, backup, restore and upgrade preserve or explicitly degrade the capability |
| Accessibility | keyboard, focus, labeling, error association, reduced motion and supported viewport checks pass |
| Documentation | operator, end-user, API and recovery instructions match the tested release behavior |

Promotion is an automated release decision. The UI reads the server-published
manifest and cannot promote itself because a route or control exists.

The normative machine contract is
`backend/v1/domain/evidence-manifest.schema.json`. Its required evidence layer
IDs are `product_contract`, `threat_model`, `source_legal`, `persistence`,
`service_implementation`, `authorization`, `protocol_behavior`,
`operator_api`, `guided_ui`, `action_equivalence`, `audit`, `browser`,
`failure`, `operations`, `accessibility`, `documentation`, `supply_chain` and
`release_policy`. A layer that has no protocol or other direct behavior still
requires a passing applicability decision; omission does not count as proof.
The domain validator is the promotion authority and derives the discovery
proof digest from the validated manifest.

## Required test pyramid

Each capability owns tests at the narrowest useful boundary and at the real
product boundary:

- pure domain and policy tests;
- PostgreSQL migration, constraint, RLS and transaction tests;
- API authorization, schema and idempotency tests;
- protocol conformance and hostile-input tests;
- Playwright human journey tests against the production image;
- Playwright accessibility and failure-remedy tests;
- Clyffy action equivalence tests against the same service command;
- restart, backup/restore and supported-version upgrade drills.

Snapshots and source-text searches may prevent regressions but never count as
behavioral evidence.

## Source intake and product absorption

Nomen may absorb suitable implementation material from approved upstream
sources only through a recorded intake:

1. identify the exact capability gap and security invariants;
2. record source revision, license and any applicable private permission;
3. inventory code, migrations, protocol fixtures and tests separately;
4. port into Nomen-owned packages, configuration, schema and terminology;
5. preserve required legal notices outside runtime product identity;
6. remove assumptions that conflict with Nomen's tenant, custody, operator
   and deployment contracts;
7. add Nomen UI, API, Clyffy, audit and operational evidence; and
8. run the complete Nomen promotion gate.

Copy volume is not progress. A smaller absorbed implementation with complete
evidence is preferable to two cores running beside each other with divergent
identity state.

## Pull-request rule

Every capability pull request links its contract and evidence manifest, lists
each layer above as passing or incomplete, and leaves the server capability in
`inventory` or `preview` until all required layers pass. Reviewers can reproduce
the evidence from a clean checkout without access to a developer workstation.
