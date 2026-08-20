# 10 — OCI-compatible container runtime contract

**Status:** accepted deployment contract; the image and Podman examples follow
this document.

## Purpose

Tessera ships as one OCI-runtime-compatible image that runs unchanged in
standalone, managed-customer and optional host-integrated deployments. Podman,
Kubernetes or another OCI runtime may supervise it; the image does not contain
cloud-specific orchestration.

The container owns the Tessera process. It does not own PostgreSQL, ingress,
DNS, certificates, billing or mesh inventory. Those remain deployment or
optional adapter concerns.

Managed Tessera supports two launch profiles. The dedicated profile assigns a
customer its own runtime and logical PostgreSQL database. The invite-only
community profile shares a supervised runtime while enforcing tenant boundaries
in every identity, policy, audit and analytics projection. A deployment must
declare its profile; it may not infer isolation from a hostname.

## Runtime shape

| concern | contract |
|---|---|
| process | one foreground `tessera` process; `SIGTERM` is the graceful stop |
| identity | fixed non-root uid/gid `10001`; no Linux capabilities required |
| filesystem | read-only root filesystem; temporary files use runtime tmpfs |
| network | listens on container port `8080`; bind to loopback or a private OCI network |
| edge TLS | deployment ingress terminates TLS; Tessera is told the public HTTPS domain and port |
| database | external PostgreSQL; the image contains no database and declares no data volume |
| desired state | optional read-only blueprint mount at `/etc/tessera/blueprints` |
| health | the image and Podman unit run `tessera ready`, which checks `/debug/ready` locally |
| logs | structured logs go to stderr for the runtime journal or collector |

The runtime image contains the statically linked binary and CA trust roots. It
contains no shell, package manager, source tree, compiler or credential.

## Bootstrap and steady state

The safe image default is steady-state `start`. It fails closed against an
uninitialized database instead of silently creating the inherited development
administrator.

Provisioning may run the same image with `start-from-init` only when all three
bootstrap secrets are attached:

1. the 32-character Tessera master key as `TESSERA_MASTERKEY`;
2. a PostgreSQL DSN as `TESSERA_DATABASE_POSTGRES_DSN`;
3. a deployment-generated first-administrator password as
   `TESSERA_FIRSTINSTANCE_ORG_HUMAN_PASSWORD`.

The database and role named by DSN already exist and have the privileges needed
for schema initialization. A managed operator provisions them; self-hosted
Tessera may use its own PostgreSQL container or server. Tessera never starts
PostgreSQL as a child process.

`start-from-init` is the Podman single-service convenience: it converges schema
and setup steps before starting the foreground server. The recorded setup steps
make subsequent starts idempotent. A managed deployment may instead run `init`
and `setup` as explicit jobs, then use the image's default `start` command.

No bootstrap secret may be passed as a command-line argument, baked into an
image layer, stored in a Quadlet file, written to a blueprint or printed by an
automation log. Podman secret-to-environment injection is acceptable because
the value is resolved at container creation and the repository holds only the
secret name.

## Configuration boundary

Non-secret runtime configuration is supplied through an external environment
file. At minimum it declares:

```text
TESSERA_EXTERNALDOMAIN
TESSERA_EXTERNALPORT=443
TESSERA_EXTERNALSECURE=true
TESSERA_TLS_ENABLED=false
TESSERA_PORT=8080
```

The public domain must be stable before the first instance is created because
it becomes issuer and WebAuthn origin state. Changing it is a migration, not a
container restart option.

Deployment-owned configuration names the image digest, network, database
endpoint, isolation profile, resource limits and secret provider. Tessera-owned
configuration names identity policy, flows, providers and blueprints. Neither
side copies live credential values across this boundary.

## Podman and private-cloud boundary

The example Quadlet is a systemd-generated Podman unit. It is intentionally
rootless-compatible and applies the same constraints required in managed
Tessera: non-root image user, read-only root, dropped capabilities, no-new-
privileges, private networking, loopback-only publication and a readiness
health check.

The `tessera-operator` may render the example with a verified registry image
digest and create the three named Podman secrets. Its service account receives
only the container, database-migration, backup and protected-reference rights
needed by a reviewed operation. It must not rewrite the image or invent a
second Tessera configuration model. Zuul may connect the deployment to a mesh,
but Tessera remains an identity service rather than a peer inventory.

## Verification gates

The container slice is done when:

- a clean OCI build bootstraps the custom protocol option, generates the full
  protocol stubs and produces a static runtime image;
- the image reports uid/gid `10001`, has no shell and runs `tessera --help`;
- the image configuration exposes `8080`, defines the built-in readiness check
  and defaults to the steady-state `start` command;
- the Quadlet generates a valid systemd service on the supported Podman
  version;
- neither the image history nor tracked deployment examples contain a live
  master key, database credential, bootstrap password, token or private key;
- the existing seat-token contract and Automaton verifier remain unchanged.
