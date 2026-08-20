#!/usr/bin/env bash
# Ephemeral OpenLDAP StartTLS/LDAPS conformance. Every credential and private
# key is generated under a temporary directory and removed on exit.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
NAME=tessera-ldap-conformance
if podman container exists "$NAME"; then
  printf 'refusing to replace existing container %s\n' "$NAME" >&2
  exit 1
fi

RUNTIME="$(mktemp -d -t tessera-ldap-conformance.XXXXXX)"
cleanup() {
  podman logs "$NAME" > /tmp/tessera-openldap-last.log 2>&1 || true
  podman rm -f "$NAME" >/dev/null 2>&1 || true
  # The rootless container remaps its certificate mount. Remove only this
  # mktemp-created path from inside the same user namespace.
  podman unshare rm -rf -- "$RUNTIME" >/dev/null 2>&1 || true
}
trap cleanup EXIT INT TERM
trap 'code=$?; printf "LDAP probe failed at line %s (exit %s)\n" "$LINENO" "$code" | tee /tmp/tessera-ldap-last-error >&2; exit "$code"' ERR

ADMIN_PASSWORD="$(openssl rand -hex 24)"
USER_PASSWORD="$(openssl rand -hex 24)"
mkdir -p "$RUNTIME/certs"
openssl req -x509 -newkey rsa:3072 -nodes -days 1 -subj '/CN=Tessera LDAP Conformance CA' \
  -keyout "$RUNTIME/ca.key" -out "$RUNTIME/certs/ca.crt" >/dev/null 2>&1
openssl req -newkey rsa:3072 -nodes -subj '/CN=localhost' \
  -keyout "$RUNTIME/certs/server.key" -out "$RUNTIME/server.csr" >/dev/null 2>&1
printf 'subjectAltName=DNS:localhost,IP:127.0.0.1\nextendedKeyUsage=serverAuth\n' > "$RUNTIME/server.ext"
openssl x509 -req -days 1 -in "$RUNTIME/server.csr" -CA "$RUNTIME/certs/ca.crt" -CAkey "$RUNTIME/ca.key" -CAcreateserial -CAserial "$RUNTIME/ca.srl" \
  -extfile "$RUNTIME/server.ext" -out "$RUNTIME/certs/server.crt" >/dev/null 2>&1
# Use the standardized RFC 7919 finite-field group so image startup has a
# bounded runtime without weakening or checking in TLS parameters.
openssl genpkey -genparam -algorithm DH -pkeyopt group:ffdhe2048 \
  -out "$RUNTIME/certs/dhparam.pem" >/dev/null 2>&1
cp "$RUNTIME/certs/ca.crt" "$RUNTIME/ca-public.crt"
chmod 600 "$RUNTIME/certs/"*.key

ENV_FILE="$RUNTIME/openldap.env"
umask 077
printf '%s\n' \
  'LDAP_ORGANISATION=Tessera Conformance' \
  'LDAP_DOMAIN=example.test' \
  "LDAP_ADMIN_PASSWORD=$ADMIN_PASSWORD" \
  "LDAP_CONFIG_PASSWORD=$(openssl rand -hex 24)" \
  'LDAP_TLS=true' \
  'LDAP_TLS_CRT_FILENAME=server.crt' \
  'LDAP_TLS_KEY_FILENAME=server.key' \
  'LDAP_TLS_CA_CRT_FILENAME=ca.crt' \
  'LDAP_TLS_VERIFY_CLIENT=never' > "$ENV_FILE"

podman run -d --name "$NAME" --env-file "$ENV_FILE" \
  -p 127.0.0.1:1389:389 -p 127.0.0.1:1636:636 \
  -v "$RUNTIME/certs:/container/service/slapd/assets/certs:Z,U" \
  docker.io/osixia/openldap:1.5.0 >/dev/null

READY=false
for _ in $(seq 1 240); do
  if openssl s_client -connect localhost:1636 -servername localhost -CAfile "$RUNTIME/ca-public.crt" -verify_return_error </dev/null >/dev/null 2>&1; then
    READY=true
    break
  fi
  sleep 0.5
done
if [[ "$READY" != true ]]; then
  printf 'OpenLDAP did not pass its TLS readiness check\n' >&2
  podman logs --tail 80 "$NAME" >&2 || true
  exit 1
fi

TESSERA_LDAP_ADMIN_PASSWORD="$ADMIN_PASSWORD" \
TESSERA_LDAP_USER_PASSWORD="$USER_PASSWORD" \
TESSERA_LDAP_CA_FILE="$RUNTIME/ca-public.crt" \
  go test -tags ldapconformance ./backend/v1/federation/ldapoutbound -run TestOpenLDAPConformance -v

printf '\nOpenLDAP StartTLS/LDAPS conformance passed\n'
