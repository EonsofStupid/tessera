#!/usr/bin/env bash
# Build Nomen's OCI-runtime-compatible image with the Docker-v2 image
# manifest. Podman's default OCI manifest drops Dockerfile HEALTHCHECK metadata;
# Docker-v2 preserves it and remains consumable by Podman and Kubernetes.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TAG="${NOMEN_IMAGE_TAG:-localhost/nomen:1.0.0-alpha}"
VERSION="${NOMEN_BUILD_VERSION:-1.0.0-alpha}"
REVISION="${NOMEN_BUILD_REVISION:-uncommitted}"
BUILD_DATE="${NOMEN_BUILD_DATE:-1970-01-01T00:00:00Z}"

exec podman build \
  --format docker \
  --tag "$TAG" \
  --build-arg "VERSION=$VERSION" \
  --build-arg "COMMIT=$REVISION" \
  --build-arg "BUILD_DATE=$BUILD_DATE" \
  "$ROOT"
