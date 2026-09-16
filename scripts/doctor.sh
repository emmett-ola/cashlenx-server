#!/usr/bin/env bash
set -euo pipefail
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"
export LIFECYCLE_DIAGNOSTIC_MODE=doctor
exec "$script_dir/status.sh"
