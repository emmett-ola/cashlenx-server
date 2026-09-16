#!/usr/bin/env bash
set -euo pipefail

# Prevent Git Bash on Windows from rewriting container-internal absolute paths.
export MSYS_NO_PATHCONV=1

project_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)"
. "$project_dir/scripts/lib/container_lifecycle.sh"
container_runtime_init auto

image_ref="${1:?image reference is required}"
expected_version="${2:?expected version is required}"
expected_revision="${3:?expected revision is required}"

label() {
  container image inspect "$image_ref" --format "{{ index .Config.Labels \"$1\" }}"
}

[[ "$(label org.opencontainers.image.version)" == "$expected_version" ]] || { echo "Image version label mismatch." >&2; exit 1; }
[[ "$(label org.opencontainers.image.revision)" == "$expected_revision" ]] || { echo "Image revision label mismatch." >&2; exit 1; }

container run --rm --entrypoint sh "$image_ref" -ec '
  test -x /app/cashlenx-server
  test -s /app/docs/openapi.yaml
  test -s /app/config/default_categories.json
  if find /app -type f \( -name ".env" -o -name ".env.*" -o -name "*.pem" -o -name "*.key" \) -print -quit 2>/dev/null | grep -q .; then
    echo "Prohibited environment or credential file found in image." >&2
    exit 1
  fi
  if find /app -type d -name .git -print -quit 2>/dev/null | grep -q .; then
    echo "Git metadata found in image." >&2
    exit 1
  fi
'

version_output="$(container run --rm --entrypoint /app/cashlenx-server "$image_ref" open version)"
grep -Fqx "CashLenX v${expected_version}" <<<"$version_output" || { echo "Embedded runtime version mismatch." >&2; exit 1; }
grep -Fqx "Git Commit: ${expected_revision}" <<<"$version_output" || { echo "Embedded runtime revision mismatch." >&2; exit 1; }
