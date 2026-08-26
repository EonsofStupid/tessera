# syntax=docker/dockerfile:1

FROM docker.io/library/node@sha256:2bdb65ed1dab192432bc31c95f94155ca5ad7fc1392fb7eb7526ab682fa5bf14 AS ui

WORKDIR /src/web
ENV CI=true

RUN npm install --global pnpm@11.22.0
COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile
COPY web/ ./
RUN pnpm typecheck && pnpm test && pnpm build

FROM docker.io/library/golang@sha256:49be5c3f5f2b766e5ba74e0bb690fea4fa03ebf5df8fe94665d42dfa727acf31 AS build

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
COPY --from=ui /src/internal/api/ui/console/static/ internal/api/ui/console/static/

# protoc-gen-nomen reads its own custom option type. Generate that one type
# with the standard Go plugin first, then build the custom plugins and run the
# complete graph. This keeps clean clones independent of ignored artifacts.
RUN buf generate \
      --template buf.gen.bootstrap.yaml \
      --path proto/nomen/protoc_gen_nomen/v2/options.proto \
    && mkdir -p pkg/grpc/protoc/v2 \
    && cp .artifacts/grpc/github.com/shippinAI/nomen/pkg/grpc/protoc/v2/*.go pkg/grpc/protoc/v2/ \
    && GOBIN=/usr/local/bin go install \
      ./internal/protoc/protoc-gen-authoption \
      ./internal/protoc/protoc-gen-nomen \
    && buf generate \
    && mkdir -p pkg/grpc openapi/v2/nomen \
    && cp -R .artifacts/grpc/github.com/shippinAI/nomen/pkg/grpc/. pkg/grpc/ \
    && cp .artifacts/grpc/nomen/*.swagger.json openapi/v2/nomen/

ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG VERSION=1.0.0-alpha
ARG COMMIT=unknown
ARG BUILD_DATE=1970-01-01T00:00:00Z

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -buildvcs=false -trimpath \
      -ldflags="-s -w \
        -X github.com/shippinAI/nomen/cmd/build.version=${VERSION} \
        -X github.com/shippinAI/nomen/cmd/build.commit=${COMMIT} \
        -X github.com/shippinAI/nomen/cmd/build.date=${BUILD_DATE}" \
      -o /out/nomen . \
    && printf 'nomen:x:10001:10001:Nomen:/:/sbin/nologin\n' > /out/passwd \
    && printf 'nomen:x:10001:\n' > /out/group \
    && mkdir -m 1777 /out/tmp \
    && go clean -cache -modcache

FROM scratch AS runtime

ARG VERSION=1.0.0-alpha
ARG COMMIT=unknown
ARG BUILD_DATE=1970-01-01T00:00:00Z

LABEL org.opencontainers.image.title="Nomen" \
      org.opencontainers.image.description="Standalone identity and access management platform" \
      org.opencontainers.image.source="https://github.com/shippinAI/nomen" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${COMMIT}" \
      org.opencontainers.image.created="${BUILD_DATE}"

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/passwd /etc/passwd
COPY --from=build /out/group /etc/group
COPY --from=build --chown=10001:10001 /out/tmp /tmp
COPY --from=build --chown=10001:10001 /out/nomen /usr/local/bin/nomen

ENV NOMEN_PORT=8080 \
    NOMEN_TLS_ENABLED=false \
    NOMEN_EXTERNALSECURE=true \
    NOMEN_EDITION=public \
    NOMEN_DEMO_CAPS=false \
    NOMEN_TELEMETRY_ENABLED=false \
    NOMEN_INSTRUMENTATION_SERVICENAME=nomen \
    NOMEN_INSTRUMENTATION_LOG_STDERR=json

EXPOSE 8080
USER 10001:10001
STOPSIGNAL SIGTERM

HEALTHCHECK --interval=10s --timeout=3s --start-period=60s --retries=12 \
  CMD ["/usr/local/bin/nomen", "ready"]

ENTRYPOINT ["/usr/local/bin/nomen"]
# Safe steady state: schema/bootstrap is an explicit deployment action.
CMD ["start", "--masterkeyFromEnv", "--tlsMode", "external"]
