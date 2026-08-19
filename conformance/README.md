# Tessera conformance

This directory is the executable release map for the capabilities in
`docs/18-identity-edge-and-vaultix-contract.md`.

`suites.json` defines the minimum black-box cases. It contains no credentials,
directory exports, cookies or private keys. Implementations write raw evidence
under `.artifacts/workspace/evidence/` and produce a redacted proof envelope;
only a passing proof bound to the installed bundle may enter capability
discovery.

Suite status meanings:

- `planned` — contract exists; capability remains preview/unsupported;
- `implemented` — test driver exists but has not passed release verification;
- `verified` — the exact bundle passed and produced immutable evidence.

Changing a customer-visible behavior requires changing contract 18 and the
relevant suite before changing runtime code.
