#!/usr/bin/env bash
set -euo pipefail

go test -v -race -covermode=atomic -coverprofile=coverage.out ./...
