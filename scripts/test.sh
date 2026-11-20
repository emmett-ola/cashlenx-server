#!/usr/bin/env bash
# Single test entry point for the Server.
#
# Runs the Go package suite, which is what CI and the release gate call. The
# build-tagged migration integration test and the API/database smoke scripts
# are separate entry points: they need disposable MongoDB or MySQL instances
# and a running server.
#
# Extra arguments are forwarded to `go test`, for example one package:
#   scripts/test.sh ./service/statistic_service/...
set -euo pipefail

project_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)"
cd "$project_dir"

exec go test ./... "$@"
