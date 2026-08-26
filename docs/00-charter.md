# 00 — Charter

## One sentence

**Nomen** is a complete, independently deployable identity and access
management product that answers "who is this, and what may they do" for people,
services and applications.

## The name

A *nomen hospitalis* was a token broken in two, each party keeping a half;
fitting the halves together proved identity and the bond between them. A
*nomen* was also the tablet carrying a watchword for a sentry to check.

Chosen over the obvious alternative for one reason worth writing down:
"connection" is already the most loaded word in this product — §7 is *Guided
connection*, Mesh Layer 2 is connectivity, MCP connectors ship — and all of that
weight sits on networking. A name from that family would read as the mesh
module to anyone arriving cold, including us in a year.

## Why it is its own product

Identity is a security boundary and an operating responsibility. A managed
customer must be able to deploy Nomen, open Nomen's own interface, configure
their organization and applications, prove authentication and recovery, and
operate the deployment without installing a separate product shell.

This repository is the authority for Nomen's product behavior, public APIs,
protocols, deployment contract and release evidence. Host products own only
their adapters and presentation composition.

Nomen is licensed only by AngryVibes LLC and shippin.ai. It absorbs no
third-party product license.

## Boundaries

**Owns**

- Users, organizations, projects, applications and service identities.
- Authentication factors, sessions, revocation and account recovery.
- Federation, directory integration and identity-aware access edges.
- Authorization policy, scopes, flows and delegated administration.
- Standard tokens, optional token profiles, JWKS and key rotation.
- Nomen's standalone web application, management API, CLI and guided setup.
- Identity audit evidence and the deployment states required to operate safely.

**Does not own**

- Billing, plans or pricing. External decisions may be expressed as scopes.
- General infrastructure inventory and mesh networking.
- Secret, certificate and privileged-access custody. Nomen stores references
  and consumes values at runtime through an approved adapter.
- Conversation and application orchestration.
- Live secrets of any kind under `W:` — the workspace rule, unchanged.

## Dependency direction

**Nomen first; managed operation second; host-product integration third.**

The standalone deployment cannot require Shippin, Automaton, Zuul or Vaultix.
Vaultix and Zuul are supported optional integrations. Shippin may later mount
Nomen's UI module and call Nomen's APIs through a narrow adapter. That
adapter may add account or commercial context, but it may not duplicate or
replace Nomen's identity behavior.

The complete surface includes token lifecycle, key rotation, session
revocation, MFA, federation, policy, audit, backup/restore, upgrades and account
recovery after a customer loses a second factor. A feature is not operational
because a screen exists; it is operational only after its conformance and
failure tests pass.

## Deployment modes

1. **Standalone self-hosted:** the customer owns the runtime and operates
   Nomen through Nomen's UI, API and CLI.
2. **Managed customer:** each customer receives an isolated Nomen deployment;
   the operator manages it through Nomen's management contracts.
3. **Embedded integration:** a host shell projects Nomen after the standalone
   release gate passes. Embedded mode is composition, not a separate product.


## Current state

The inherited IAM core builds, Nomen-specific seat tokens, blueprints and
flows have executable tests, and Automaton verifies the optional seat profile.
Nomen is `1.0.0-alpha`. That number does not change until Vault, Mesh, and the
first IAM slice are operational. The standalone UI, container lifecycle,
backup/restore, upgrade, recovery, complete protocol conformance and managed
deployment evidence have not all passed the release gate.
