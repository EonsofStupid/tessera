# 23 — Release governance record

**Status:** active; records non-secret product-owner decisions and unresolved
release prerequisites.
**Recorded:** 2026-08-21.
**Program:** `22-certification-and-parity-program.md`.

## Current product-owner decision

| Field | Recorded decision |
|---|---|
| Product owner | Jesse Hall |
| Legal product owner | AngryVibes LLC |
| Product | Nomen |
| Intended service modes | Private standalone deployment or Shippin-managed deployment |
| Intended audience | AngryVibes LLC users and invited managed customers; no public self-service enrollment |
| Public surface | Public commercial product/entry page with no private tenant, identity, policy, audit or deployment facts |
| Artifact distribution | Controlled managed-container distribution; no public general-purpose container distribution is selected |
| Source distribution | Not an open-source product mode. Nomen is licensed only by AngryVibes LLC and shippin.ai. |
| Public release authority | AngryVibes LLC through product owner/operator Jesse Hall, together with shippin.ai as an authorized licensor |
| Product license | AngryVibes LLC and shippin.ai only. Nomen absorbs no third-party product license. |
| Product version | `1.0.0-alpha`, frozen. No 1.1, no v2 product, no second name until Nomen Vault, Nomen Mesh, and the first IAM slice are actually built. |
| Market posture | Private commercial product; no government, finance, healthcare or other regulated-market claim is assumed |

## Frozen 1.0.0-alpha

This is the first Nomen. Internal package paths such as `backend/v1`,
`backend/v3`, or gRPC `v2` are inherited layout, not product versions. Do not
rename or mint a new product version to look finished. `/nomen/v1` is the one
management surface and stays frozen with this alpha.

The version string changes only when those services are operational and the
product owner records a new number here.

## Product license decision (recorded)

Nomen's product license is **AngryVibes LLC and shippin.ai only**.

The product does **not** absorb AGPL, Apache, MIT, BSD, or any other third-party
product license from ZITADEL, Authentik, or similar identity products. Runtime
source is Nomen. `upstream/` snapshots, if present on disk, are non-product
reference and are not shipped. Go/npm dependency licenses stay at the
dependency boundary and are inventoried by the G1 SBOM; they are not Nomen's
license.

Protected permission evidence remains outside Git (`L0.2`). This record names
the product licensors. It does not paste waiver text.

This record does not put private correspondence or permission documents in the
repository. It defines the product and distribution facts that counsel must
evaluate against those protected records.

## Consequences of the selected distribution model

The legal and release review must address both distinct activities:

1. operating modified software for private users as a standalone or
   Shippin-managed network-accessible service; and
2. distributing modified executable and container artifacts through a
   controlled managed-customer channel.

The disposition must separately answer whether Nomen may use, modify, host,
copy and distribute each retained source class, and what source availability,
notice, attribution, offer, trademark, termination and downstream-recipient
conditions attach to each activity.

## Protected permission evidence

The original inbound permission messages, contracts, waivers, attachments
and identities remain in protected business records. Git stores only the
following non-secret review fields after the records exist:

| Field | Required meaning | Current state |
|---|---|---|
| `evidence_id` | stable opaque identifier resolvable by AngryVibes LLC/Jesse Hall and counsel | not supplied |
| `source_owner` | inbound rights holder | known categories; record mapping missing |
| `grantor_authority` | counsel-confirmed authority of the person/entity granting rights | unverified |
| `grantee` | exact person or legal entity receiving rights | must cover AngryVibes LLC directly or through a counsel-approved assignment from Jesse Hall |
| `effective_date` | effective date and any expiry/renewal date | not supplied |
| `source_scope` | repositories, paths, versions/commits and enterprise exclusions | not supplied |
| `use_scope` | internal use and modification rights | not supplied |
| `host_scope` | right to operate the modified work as a hosted managed service | not supplied |
| `distribution_scope` | right to distribute modified binaries and containers | not supplied |
| `source_obligation` | source publication, source offer or network-source obligations | not supplied |
| `sublicense_transfer` | rights needed by managed customers/operators receiving the container | not supplied |
| `notice_attribution` | required copyright, license, provenance and public wording | not supplied |
| `trademark_scope` | permitted name/logo/reference language and prohibited endorsement claims | not supplied |
| `termination` | revocation, breach, survival and version-upgrade rules | not supplied |
| `counsel_disposition_id` | final written legal conclusion for the release modes above | missing |
| `public_wording_digest` | digest of the exact approved public provenance/attribution text | missing |

No capability gate relies on an oral summary where a written permission is
required. Donations, sponsorship or informal encouragement do not substitute
for rights covering the selected hosted and container-distribution modes.

## Counsel review question set

Counsel should give a written, source-class-specific answer to these questions:

1. Is AngryVibes LLC the grantee? If Jesse Hall is named personally, may the
   rights be assigned to or exercised by AngryVibes LLC without new consent?
2. Do the documents cover private standalone and Shippin-managed
   network-accessible services?
3. Do they cover distribution of modified OCI images, binaries, CLI artifacts,
   templates, generated clients and documentation?
4. Which source revisions, paths and generated artifacts are covered? Are any
   enterprise or separately licensed paths excluded?
5. What source publication, corresponding-source, written-offer or downstream
   notice obligations apply to hosted use and controlled managed-container
   distribution?
6. Do third-party dependencies retain independent obligations that the waiver
   cannot alter?
7. May Nomen use upstream names in factual provenance, and what exact public
   wording is approved without implying partnership or endorsement?
8. Are the permissions perpetual, worldwide, irrevocable and royalty-free? If
   not, what term, territory, payment, audit or termination conditions apply?
9. Do updates or future upstream versions require new permission?
10. Are warranty, indemnity, export-control, privacy or data-processing terms
    attached to the permission or commercial relationship?
11. The product-owner selection is AngryVibes LLC and shippin.ai only, with no
    absorbed third-party product license. Confirm that `LICENSE` and `NOTICE`
    match that selection for the intended hosted and container-distribution
    modes.
12. What evidence may Nomen retain and publish to prove compliance without
    disclosing confidential correspondence?

This question set is an engineering handoff, not legal advice.

## Remaining product-owner decisions

These decisions do not block read-only source and branch inventory, but they
block the named later gate:

| Decision | Blocks |
|---|---|
| Protected evidence IDs for inbound source permissions | L0 source/permission gate |
| Public domain and DNS authority | stable issuer, WebAuthn RP ID and public TLS |
| Source host, CI identity and OCI registry | signed release factory |
| Managed hosting region/provider | threat model, data residency, subprocessors and audit scope |
| Security contact and vulnerability disclosure destination | public release |

AngryVibes LLC's jurisdiction, address and other corporate particulars are not
requested or stored by this engineering program. Counsel or service providers
may collect them privately when legally required.
