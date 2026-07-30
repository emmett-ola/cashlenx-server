ARG GO_BUILD_IMAGE=golang:1.23-alpine
ARG RUNTIME_IMAGE=alpine:3.18

FROM ${GO_BUILD_IMAGE} AS builder
WORKDIR /app

# Install dependencies for building
RUN apk add --no-cache git

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and build
COPY . .
RUN go build -o cashlenx-server main.go

FROM ${RUNTIME_IMAGE}

ARG GIT_COMMIT=unknown
LABEL org.opencontainers.image.revision="${GIT_COMMIT}"
WORKDIR /app

# Install curl for health checks
RUN apk add --no-cache curl

# Create necessary directories
RUN mkdir -p docs

# Copy the built binary from the build stage
COPY --from=builder /app/cashlenx-server .

# Copy the docs directory containing the OpenAPI spec
COPY --from=builder /app/docs/openapi.yaml /app/docs/

# Use the pre-built binary as entrypoint
ENTRYPOINT ["./cashlenx-server", "open", "start"]
