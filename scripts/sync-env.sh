#!/usr/bin/env bash
set -euo pipefail

project_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
sample_file="${1:-$project_dir/.env.sample}"
env_file="${2:-$project_dir/.env}"

[[ -f "$sample_file" ]] || { echo "Missing environment sample: $sample_file" >&2; exit 1; }

if [[ ! -f "$env_file" ]]; then
  cp "$sample_file" "$env_file"
  echo "Created $(basename "$env_file") from $(basename "$sample_file")."
  exit 0
fi

added=0
while IFS= read -r line || [[ -n "$line" ]]; do
  [[ "$line" =~ ^[A-Za-z_][A-Za-z0-9_]*= ]] || continue
  key="${line%%=*}"
  if ! grep -q "^${key}=" "$env_file"; then
    if [[ $added -eq 0 ]]; then
      printf '\n# Added from %s; review before deployment.\n' "$(basename "$sample_file")" >> "$env_file"
    fi
    printf '%s\n' "$line" >> "$env_file"
    echo "Added missing key: $key"
    added=$((added + 1))
  fi
done < "$sample_file"

echo "Environment sync complete; existing values were preserved."
