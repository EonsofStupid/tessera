# syntax=docker/dockerfile:1

ARG GO_VERSION=1.25.11

FROM docker.io/library/golang:${GO_VERSION}-bookworm AS build

WORKDIR /src

# Protocol generators are pinned because their output is part of the binary.
# They live only in this build stage.
RUN GOBIN=/usr/local/bin go install github.com/bufbuild/buf/cmd/buf@v1.67.0 \
    && GOBIN=/usr/local/bin go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11 \
    && GOBIN=/usr/local/bin go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.1 \
    && GOBIN=/usr/local/bin go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.28.0 \
    && GOBIN=/usr/local/bin go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.28.0 \
    && GOBIN=/usr/local/bin go install github.com/envoyproxy/protoc-gen-validate@v1.3.3 \
    && GOBIN=/usr/local/bin go install connectrpc.com/connect/cmd/protoc-gen-connect-go@v1.19.1

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# protoc-gen-zitadel reads its own custom option type. Generate that one type
# with the standard Go plugin first, then build the custom plugins and run the
# complete graph. This keeps clean clones independent of ignored artifacts.
RUN buf generate \
      --template buf.gen.bootstrap.yaml \
      --path proto/zitadel/protoc_gen_zitadel/v2/options.proto \
    && mkdir -p pkg/grpc/protoc/v2 \
    && cp .artifacts/grpc/github.com/EonsofStupid/tessera/pkg/grpc/protoc/v2/*.go pkg/grpc/protoc/v2/ \
    && GOBIN=/usr/local/bin go install \
      ./internal/protoc/protoc-gen-authoption \
      ./internal/protoc/protoc-gen-zitadel \
    && buf generate \
    && mkdir -p pkg/grpc openapi/v2/zitadel \
    && cp -R .artifacts/grpc/github.com/EonsofStupid/tessera/pkg/grpc/. pkg/grpc/ \
    && cp .artifacts/grpc/zitadel/*.swagger.json openapi/v2/zitadel/

ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=1970-01-01T00:00:00Z

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -buildvcs=false -trimpath \
      -ldflags="-s -w \
        -X github.com/EonsofStupid/tessera/cmd/build.version=${VERSION} \
        -X github.com/EonsofStupid/tessera/cmd/build.commit=${COMMIT} \
        -X github.com/EonsofStupid/tessera/cmd/build.date=${BUILD_DATE}" \
      -o /out/tessera . \
    && printf 'tessera:x:10001:10001:Tessera:/:/sbin/nologin\n' > /out/passwd \
    && printf 'tessera:x:10001:\n' > /out/group \
    && mkdir -m 1777 /out/tmp

FROM scratch AS runtime

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=1970-01-01T00:00:00Z

LABEL org.opencontainers.image.title="Tessera" \
      org.opencontainers.image.description="Shippin identity and authorization service" \
      org.opencontainers.image.source="https://github.com/EonsofStupid/tessera" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${COMMIT}" \
      org.opencontainers.image.created="${BUILD_DATE}"

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/passwd /etc/passwd
COPY --from=build /out/group /etc/group
COPY --from=build --chown=10001:10001 /out/tmp /tmp
COPY --from=build --chown=10001:10001 /out/tessera /usr/local/bin/tessera

ENV TESSERA_PORT=8080 \
    TESSERA_TLS_ENABLED=false \
    TESSERA_EXTERNALSECURE=true \
    TESSERA_TELEMETRY_ENABLED=false \
    TESSERA_INSTRUMENTATION_SERVICENAME=tessera \
    TESSERA_INSTRUMENTATION_LOG_STDERR=json

EXPOSE 8080
USER 10001:10001
STOPSIGNAL SIGTERM

HEALTHCHECK --interval=10s --timeout=3s --start-period=60s --retries=12 \
  CMD ["/usr/local/bin/tessera", "ready"]

ENTRYPOINT ["/usr/local/bin/tessera"]
# Safe steady state: schema/bootstrap is an explicit deployment action.
CMD ["start", "--masterkeyFromEnv", "--tlsMode", "external"]
