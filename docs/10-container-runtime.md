# 10 — OCI-compatible container runtime contract

**Status:** accepted deployment contract; the image and Podman examples follow
this document.

## Purpose

Nomen ships as one OCI-runtime-compatible image that runs unchanged in
standalone, managed-customer and optional host-integrated deployments. Podman,
Kubernetes or another OCI runtime may supervise it; the image does not contain
cloud-specific orchestration.

The container owns the Nomen process. It does not own PostgreSQL, ingress,
DNS, certificates, billing or mesh inventory. Those remain deployment or
optional adapter concerns.

Managed Nomen supports two launch profiles. The dedicated profile assigns a
customer its own runtime and logical PostgreSQL database. The invite-only
community profile shares a supervised runtime while enforcing tenant boundaries
in every identity, policy, audit and analytics projection. A deployment must
declare its profile; it may not infer isolation from a hostname.

## Runtime shape

| concern | contract |
|---|---|
| process | one foreground `nomen_product` process; `SIGTERM` is the graceful stop |
| identity | fixed non-root uid/gid `10001`; no Linux capabilities required |
| filesystem | read-only root filesystem; temporary files use runtime tmpfs |
| network | listens on container port `8080`; bind to loopback or a private OCI network |
| edge TLS | deployment ingress terminates TLS; Nomen is told the public HTTPS domain and port |
| database | external PostgreSQL; the image contains no database and declares no data volume |
| desired state | optional read-only blueprint mount at `/etc/nomen/blueprints` |
| health | the image and Podman unit run `nomen ready`, which checks `/debug/ready` locally |
| logs | structured logs go to stderr for the runtime journal or collector |

The runtime image contains the statically linked binary and CA trust roots. It
contains no shell, package manager, source tree, compiler or credential.

The canonical local build is `bash dev/build-container.sh`. It asks Podman for
the Docker-v2 image manifest because Podman's default OCI manifest silently
drops Dockerfile `HEALTHCHECK` metadata. Docker-v2 remains OCI-runtime
compatible and preserves the declared readiness command. Release automation
must inspect the resulting image configuration and refuse an image whose user,
entrypoint, command or healthcheck differs from this contract.

## Bootstrap and steady state

The safe image default is steady-state `start`. It fails closed against an
uninitialized database instead of silently creating the inherited development
administrator.

The image default `NOMEN_EDITION=public`. Redis stays off. Set
`NOMEN_EDITION=enterprise` only on the paid/licensed path. Hosted free demo
adds `NOMEN_DEMO_CAPS=true` (`26-editions-and-demo-tier.md`).

Provisioning may run the same image with `start-from-init` when these secrets
are attached:

1. the 32-character Nomen master key as `NOMEN_MASTERKEY`;
2. a PostgreSQL DSN as `NOMEN_DATABASE_POSTGRES_DSN`;
3. optionally a deployment-generated first-administrator password as
   `NOMEN_FIRSTINSTANCE_ORG_HUMAN_PASSWORD`.

If that password is omitted, setup does not create a human. After schema
initialization the operator enrolls the first owner with the passkey ceremony
in `25-deployment-preflight-contract.md` using `NOMEN_BOOTSTRAP_AUTHORITY`.
Never put the password in the image, a compose file, a Quadlet, or git.

After schema initialization and before the first Nomen owner exists, the
steady-state container may additionally receive a minimum-32-byte
`NOMEN_BOOTSTRAP_AUTHORITY` through Podman secret-to-environment injection.
It authorizes only the owner-enrollment resources in
`25-deployment-preflight-contract.md`. Once recovery confirmation reaches
`complete`, recreate the container without that secret and remove the Podman
secret object; persisted completion prevents reuse even before physical
removal.

The database and role named by DSN already exist and have the privileges needed
for schema initialization. A managed operator provisions them; self-hosted
Nomen may use its own PostgreSQL container or server. Nomen never starts
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
NOMEN_EXTERNALDOMAIN
NOMEN_EXTERNALPORT=443
NOMEN_EXTERNALSECURE=true
NOMEN_TLS_ENABLED=false
NOMEN_PORT=8080
NOMEN_EDITION=public
```

A public compose example lives at `deploy/public.compose.yaml`. It starts
PostgreSQL and Nomen with secret files the operator provides; the repository
holds only the secret *names*. Local development that used an earlier baked
human must delete `.pgdata` and run `bash dev/up.sh` again so the first
account is created on that database, not reused from a leftover cluster.

The public domain must be stable before the first instance is created because
it becomes issuer and WebAuthn origin state. Changing it is a migration, not a
container restart option.

Deployment-owned configuration names the image digest, network, database
endpoint, isolation profile, resource limits and secret provider. Nomen-owned
configuration names identity policy, flows, providers and blueprints. Neither
side copies live credential values across this boundary.

## Podman and private-cloud boundary

The example Quadlet is a systemd-generated Podman unit. It is intentionally
rootless-compatible and applies the same constraints required in managed
Nomen: non-root image user, read-only root, dropped capabilities, no-new-
privileges, private networking, loopback-only publication and a readiness
health check.

The `nomen-operator` may render the example with a verified registry image
digest and create the three named Podman secrets. Its service account receives
only the container, database-migration, backup and protected-reference rights
needed by a reviewed operation. It must not rewrite the image or invent a
second Nomen configuration model. Nomen Mesh may connect the deployment to a
mesh, but Nomen remains an identity service rather than a peer inventory.

## Verification gates

The container slice is done when:

- a clean OCI build bootstraps the custom protocol option, generates the full
  protocol stubs and produces a static runtime image;
- the image reports uid/gid `10001`, has no shell and runs `nomen --help`;
- the image configuration exposes `8080`, defines the built-in readiness check
  and defaults to the steady-state `start` command;
- the Quadlet generates a valid systemd service on the supported Podman
  version;
- neither the image history nor tracked deployment examples contain a live
  master key, database credential, bootstrap password, token or private key;
- the existing seat-token contract and Automaton verifier remain unchanged.
