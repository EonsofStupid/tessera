# Tessera with rootless Podman

This directory is a deployable example for Shippin private cloud and a
reference for the managed Shippin Cloud service definition. Both run the same
OCI image. PostgreSQL, ingress and secret creation remain outside Tessera.

## Build the image

From the repository root:

```bash
podman build --format docker --tag localhost/shippin/tessera:dev .
podman run --rm localhost/shippin/tessera:dev --help
```

The multi-stage build generates the inherited protocol stubs, compiles a
static binary and copies only the binary, CA roots and non-root account records
into the final image. Docker v2 image format is selected because it preserves
the embedded health-check metadata; Podman executes it through the same OCI
runtime. The Quadlet also declares the health check explicitly so supervision
does not depend on image metadata.

## Prepare a rootless Quadlet

Quadlet is Podman's systemd unit generator. Install the example in the user
unit search path and keep its non-secret environment beside it:

```bash
install -d "$HOME/.config/containers/systemd/blueprints"
install -m 0644 deploy/podman/tessera.container "$HOME/.config/containers/systemd/tessera.container"
install -m 0644 deploy/podman/shippin-control-plane.network "$HOME/.config/containers/systemd/shippin-control-plane.network"
install -m 0600 deploy/podman/tessera.env.example "$HOME/.config/containers/systemd/tessera.env"
```

Edit `tessera.env` before first boot. The external domain becomes issuer and
WebAuthn origin state; changing it later is a migration.

Create the required secrets interactively so values do not enter the unit,
environment file or shell history:

```bash
systemd-ask-password "Tessera 32-character master key" | tr -d '\n' | podman secret create tessera-masterkey -
systemd-ask-password "Tessera PostgreSQL DSN" | tr -d '\n' | podman secret create tessera-database-dsn -
systemd-ask-password "Tessera initial administrator password" | tr -d '\n' | podman secret create tessera-bootstrap-password -
```

The DSN must name an existing PostgreSQL database and sufficiently privileged
role. For example, its shape is
`postgresql://USER:PASSWORD@HOST:5432/DATABASE?sslmode=verify-full`; do not copy
a real value into this repository.

Start and inspect the user service:

```bash
systemctl --user daemon-reload
systemctl --user start tessera.service
systemctl --user status tessera.service
podman healthcheck run shippin-tessera
```

The published port is loopback-only. Shippin ingress should proxy the public
HTTPS host to `127.0.0.1:8080`; do not publish Tessera directly on every host
interface.

## Managed cloud differences

Shippin Cloud replaces the local image tag with a registry digest and supplies
its managed PostgreSQL endpoint, ingress and secret provider. It may split
`init`, `setup` and the default `start` into separate jobs. It must preserve the
same image, health command, non-root user and configuration boundary.
