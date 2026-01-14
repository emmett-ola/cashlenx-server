# Multi-stage build for fast container startup

# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app

# Install dependencies for building
RUN apk add --no-cache git

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and build
COPY . .
RUN go build -o cashlenx-server main.go

# Final stage
FROM alpine:3.18
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
