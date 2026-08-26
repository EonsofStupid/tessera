# Nomen with rootless Podman

This directory is a deployable standalone Nomen example and the reference
shape for managed-customer deployments. Both isolation profiles run the same
OCI image: a dedicated runtime and logical database per customer, or an
invite-only community deployment with tenant isolation inside Nomen.
PostgreSQL, ingress and secret creation remain deployment-owned dependencies.

## Build the image

From the repository root:

```bash
podman build --format docker --tag localhost/nomen/nomen:dev .
podman run --rm localhost/nomen/nomen:dev --help
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
install -m 0644 deploy/podman/nomen.container "$HOME/.config/containers/systemd/nomen.container"
install -m 0644 deploy/podman/nomen.network "$HOME/.config/containers/systemd/nomen.network"
install -m 0600 deploy/podman/nomen.env.example "$HOME/.config/containers/systemd/nomen.env"
```

Edit `nomen.env` before first boot. The external domain becomes issuer and
WebAuthn origin state; changing it later is a migration.

Create the required secrets interactively so values do not enter the unit,
environment file or shell history:

```bash
systemd-ask-password "Nomen 32-character master key" | tr -d '\n' | podman secret create nomen-masterkey -
systemd-ask-password "Nomen PostgreSQL DSN" | tr -d '\n' | podman secret create nomen-database-dsn -
systemd-ask-password "Nomen initial administrator password" | tr -d '\n' | podman secret create nomen-bootstrap-password -
```

The DSN must name an existing PostgreSQL database and sufficiently privileged
role. For example, its shape is
`postgresql://USER:PASSWORD@HOST:5432/DATABASE?sslmode=verify-full`; do not copy
a real value into this repository.

Start and inspect the user service:

```bash
systemctl --user daemon-reload
systemctl --user start nomen.service
systemctl --user status nomen.service
podman healthcheck run nomen
```

The published port is loopback-only. The deployment ingress must proxy the
public HTTPS host to `127.0.0.1:8080`; do not publish Nomen directly on every
host interface.

## Managed deployment differences

A managed deployment replaces the local image tag with a verified registry
digest and supplies PostgreSQL, ingress and secret custody. It may split
`init`, `setup` and the default `start` into separate jobs. It must preserve the
same image, health command, non-root user and configuration boundary. The
least-privilege `nomen-operator` owns lifecycle actions; Nomen does not
require Shippin or any other host shell to install or operate.
