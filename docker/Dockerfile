ARG GO_BUILD_IMAGE=golang:1.23-alpine@sha256:383395b794dffa5b53012a212365d40c8e37109a626ca30d6151c8348d380b5f
ARG RUNTIME_IMAGE=alpine:3.18@sha256:de0eb0b3f2a47ba1eb89389859a9bd88b28e82f5826b6969ad604979713c2d4f

FROM ${GO_BUILD_IMAGE} AS builder

ARG GO_VERSION
ENV GOFLAGS=-mod=readonly
RUN test "$(go env GOVERSION)" = "go${GO_VERSION}"
WORKDIR /app

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code and build
COPY . .
ARG PRODUCT_VERSION=0.0.0
ARG GIT_COMMIT=unknown
ARG BUILD_TIME=unknown
RUN CGO_ENABLED=0 go build -trimpath \
    -ldflags="-s -w -X github.com/macar-x/cashlenx-server/cmd/open_cmd.Version=${PRODUCT_VERSION} -X github.com/macar-x/cashlenx-server/cmd/open_cmd.BuildTime=${BUILD_TIME} -X github.com/macar-x/cashlenx-server/cmd/open_cmd.GitCommit=${GIT_COMMIT}" \
    -o cashlenx-server main.go

FROM ${RUNTIME_IMAGE}

ARG GIT_COMMIT=unknown
ARG PRODUCT_VERSION=0.0.0
LABEL org.opencontainers.image.revision="${GIT_COMMIT}"
LABEL org.opencontainers.image.version="${PRODUCT_VERSION}"
WORKDIR /app

# Create necessary directories
RUN mkdir -p docs

# Copy the built binary from the build stage
COPY --from=builder /app/cashlenx-server .

# Copy the docs directory containing the OpenAPI spec
COPY --from=builder /app/docs/openapi.yaml /app/docs/

# Default category initialization is a required runtime asset.
COPY --from=builder /app/config/default_categories.json /app/config/default_categories.json

# Use the pre-built binary as entrypoint
ENTRYPOINT ["./cashlenx-server", "open", "start"]
