# 21 — Automaton session and TTY conformance

**Status:** required release gate for the first Automaton workspace session.

This gate proves the real browser boundary rather than calling the Tessera
verifier in isolation. A passing run crosses every one of these boundaries:

1. Tessera mints a short-lived, asymmetrically signed
   `shippin.seat-token.v1` for exactly `automaton:ws-0001`.
2. Automaton discovers Tessera's OIDC metadata and JWKS, verifies issuer,
   audience, signature, lifetime and the seat-token schema, and runs in strict
   mode so its legacy static credential cannot authenticate.
3. `POST /auth/session` exchanges the bearer token for an `HttpOnly`,
   `SameSite=Lax` cookie that cannot outlive the token.
4. An unauthenticated WebSocket upgrade to `/api/terminal` is a `401`.
5. The cookie-authenticated upgrade is accepted only when both
   `hosting:active` and `terminal:advanced` are present.
6. Automaton proxies the binary terminal protocol to the workspace-local
   `mato-pty` Unix socket, creates the named `tessera-first-tty` session and
   returns its `HELLO` frame.
7. A command written through the WebSocket is observed in the live PTY output.

The test does not place a token, cookie or terminal transcript in the
repository. Tokens stay in `.artifacts/`, the session cookie exists only in the
probe process, and the output assertion uses a fixed non-secret marker.

## Run it

Start Tessera and mint the probe token:

```bash
bash dev/up.sh
bash dev/seat-probe.sh
```

Start Automaton in strict mode for this workspace, with its PTY host running,
then execute:

```bash
AUTOMATON_URL=http://127.0.0.1:8111 node dev/automaton-tty-probe.mjs
```

The probe refuses a non-loopback URL. Production browser access belongs behind
the TLS boundary, but this local gate deliberately uses loopback so it can
verify the complete cookie flow without weakening `Secure` cookie behavior for
remote deployments.

## Promotion rule

This gate proves the Tessera seat-token and Automaton terminal integration. It
does not promote LDAP, proxy, visual-flow or Vaultix capabilities. Each of
those remains preview-only until the independent proof required by
`18-identity-edge-and-vaultix-contract.md` passes and is bound to the installed
bundle digest.
