#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

project_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"
cd "$project_dir"

env_file=""
backup_root=""
staging_dir=""
partial_artifact=""
database_type="unknown"
tier="${1:-daily}"
artifact_path=""

resolve_env_file() {
  local requested="${ENV_FILE:-.env}"
  local candidate
  if [[ "$requested" == /* ]]; then
    candidate="$requested"
  else
    candidate="$project_dir/$requested"
  fi

  [[ -e "$candidate" ]] || { echo "Missing environment file: $requested" >&2; return 1; }
  [[ -f "$candidate" ]] || { echo "Environment path is not a file: $requested" >&2; return 1; }

  local resolved
  resolved="$(realpath "$candidate")"
  case "$resolved" in
    "$project_dir"/*) printf '%s\n' "$resolved" ;;
    *) echo "ENV_FILE must stay inside $project_dir: $requested" >&2; return 1 ;;
  esac
}

read_env_value() {
  local key="$1"
  awk -F= -v wanted="$key" '
    $0 ~ "^[[:space:]]*(export[[:space:]]+)?" wanted "[[:space:]]*=" {
      value = substr($0, index($0, "=") + 1)
      gsub(/^[[:space:]]+|[[:space:]]+$/, "", value)
      if ((substr(value, 1, 1) == "\"" && substr(value, length(value), 1) == "\"") ||
          (substr(value, 1, 1) == "\047" && substr(value, length(value), 1) == "\047")) {
        value = substr(value, 2, length(value) - 2)
      }
      result = value
    }
    END { print result }
  ' "$env_file"
}

require_positive_integer() {
  local key="$1"
  local value="$2"
  [[ "$value" =~ ^[1-9][0-9]*$ ]] || {
    echo "$key must be a positive integer." >&2
    return 1
  }
}

resolve_operator_path() {
  local configured="$1"
  local default_value="$2"
  local candidate="${configured:-$default_value}"
  if [[ "$candidate" != /* ]]; then
    candidate="$project_dir/$candidate"
  fi
  mkdir -p "$candidate"
  realpath "$candidate"
}

write_status() {
  local result="$1"
  local exit_code="$2"
  [[ -n "$backup_root" && -d "$backup_root" ]] || return 0
  local status_dir="$backup_root/status"
  mkdir -p "$status_dir"
  local tmp="$status_dir/latest.json.partial"
  local artifact_name=""
  [[ -n "$artifact_path" && -f "$artifact_path" ]] && artifact_name="$(basename "$artifact_path")"
  printf '{"result":"%s","exit_code":%s,"database":"%s","tier":"%s","artifact":"%s","timestamp":"%s"}\n' \
    "$result" "$exit_code" "$database_type" "$tier" "$artifact_name" "$(date -u +%Y-%m-%dT%H:%M:%SZ)" > "$tmp"
  mv -f "$tmp" "$status_dir/latest.json"
}

cleanup() {
  local exit_code=$?
  [[ -z "$partial_artifact" ]] || rm -f -- "$partial_artifact" "$partial_artifact.sha256.partial"
  [[ -z "$staging_dir" ]] || rm -rf -- "$staging_dir"
  if (( exit_code == 0 )); then
    write_status success 0
  else
    write_status failed "$exit_code"
  fi
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM HUP

case "$tier" in
  daily|weekly|monthly) ;;
  *) echo "Usage: $0 [daily|weekly|monthly]" >&2; exit 2 ;;
esac

for command_name in docker openssl tar sha256sum realpath awk df; do
  command -v "$command_name" >/dev/null 2>&1 || {
    echo "Required command is unavailable: $command_name" >&2
    exit 1
  }
done

env_file="$(resolve_env_file)"
database_type="$(read_env_value DB_TYPE)"
database_type="${database_type:-mongodb}"

configured_root="$(read_env_value BACKUP_ROOT)"
backup_root="$(resolve_operator_path "$configured_root" ./backups)"
[[ "$backup_root" != "/" && "$backup_root" != "$project_dir" ]] || {
  echo "BACKUP_ROOT must not resolve to a filesystem or repository root." >&2
  exit 1
}

key_setting="$(read_env_value BACKUP_ENCRYPTION_KEY_FILE)"
[[ -n "$key_setting" && "$key_setting" != *CHANGE_ME* ]] || {
  echo "BACKUP_ENCRYPTION_KEY_FILE must name a configured passphrase file." >&2
  exit 1
}
if [[ "$key_setting" == /* ]]; then
  key_file="$key_setting"
else
  key_file="$project_dir/$key_setting"
fi
[[ -f "$key_file" && -s "$key_file" ]] || {
  echo "BACKUP_ENCRYPTION_KEY_FILE must be a non-empty regular file." >&2
  exit 1
}

minimum_free_mib="$(read_env_value BACKUP_MIN_FREE_MIB)"
minimum_free_mib="${minimum_free_mib:-1024}"
require_positive_integer BACKUP_MIN_FREE_MIB "$minimum_free_mib"
available_kib="$(df -Pk "$backup_root" | awk 'NR == 2 { print $4 }')"
require_positive_integer available_kib "$available_kib"
if (( available_kib < minimum_free_mib * 1024 )); then
  echo "Backup capacity check failed: BACKUP_MIN_FREE_MIB is not available." >&2
  exit 1
fi

case "$tier" in
  daily) retention_key=BACKUP_DAILY_RETENTION; retention_default=7 ;;
  weekly) retention_key=BACKUP_WEEKLY_RETENTION; retention_default=4 ;;
  monthly) retention_key=BACKUP_MONTHLY_RETENTION; retention_default=12 ;;
esac
retention="$(read_env_value "$retention_key")"
retention="${retention:-$retention_default}"
require_positive_integer "$retention_key" "$retention"

tier_dir="$backup_root/$tier"
mkdir -p "$tier_dir"
staging_dir="$(mktemp -d "$backup_root/.staging.XXXXXX")"
timestamp="$(date -u +%Y%m%dT%H%M%SZ)"
artifact_name="cashlenx-${database_type}-${timestamp}.tar.gz.enc"
artifact_path="$tier_dir/$artifact_name"
partial_artifact="$artifact_path.partial"

case "$database_type" in
  mongodb)
    container_name="$(read_env_value MONGO_CONTAINER_NAME)"
    container_name="${container_name:-cashlenx-mongodb}"
    docker inspect "$container_name" >/dev/null 2>&1 || {
      echo "MongoDB container is not running or does not exist: $container_name" >&2
      exit 1
    }
    docker exec "$container_name" sh -ec '
      exec mongodump \
        --host 127.0.0.1 \
        --port "${MONGO_CONTAINER_PORT:-27017}" \
        --username "$MONGO_INITDB_ROOT_USERNAME" \
        --password "$MONGO_INITDB_ROOT_PASSWORD" \
        --authenticationDatabase admin \
        --db "$MONGO_INITDB_DATABASE" \
        --archive \
        --gzip
    ' > "$staging_dir/database.archive.gz"
    dump_file="$staging_dir/database.archive.gz"
    database_name="$(read_env_value DB_NAME)"
    database_name="${database_name:-cashlenx}"
    ;;
  mysql)
    container_name="$(read_env_value MYSQL_CONTAINER_NAME)"
    container_name="${container_name:-cashlenx-mysql}"
    docker inspect "$container_name" >/dev/null 2>&1 || {
      echo "MySQL container is not running or does not exist: $container_name" >&2
      exit 1
    }
    docker exec "$container_name" sh -ec '
      MYSQL_PWD="$MYSQL_ROOT_PASSWORD" exec mysqldump \
        --host=127.0.0.1 \
        --port="${MYSQL_CONTAINER_PORT:-3306}" \
        --user=root \
        --single-transaction \
        --routines \
        --triggers \
        --events \
        --hex-blob \
        --set-gtid-purged=OFF \
        "$MYSQL_DATABASE"
    ' > "$staging_dir/database.sql"
    dump_file="$staging_dir/database.sql"
    database_name="$(read_env_value DB_NAME)"
    database_name="${database_name:-cashlenx}"
    ;;
  *) echo "DB_TYPE must be mongodb or mysql." >&2; exit 1 ;;
esac

[[ -n "$dump_file" && -s "$dump_file" ]] || {
  echo "Database dump is empty." >&2
  exit 1
}

printf 'FORMAT_VERSION=1\nDATABASE_TYPE=%s\nDATABASE_NAME=%s\nCREATED_AT=%s\n' \
  "$database_type" "$database_name" "$(date -u +%Y-%m-%dT%H:%M:%SZ)" > "$staging_dir/metadata.env"
(
  cd "$staging_dir"
  sha256sum metadata.env "$(basename "$dump_file")" > manifest.sha256
  tar -czf payload.tar.gz metadata.env manifest.sha256 "$(basename "$dump_file")"
)

openssl enc -aes-256-cbc -pbkdf2 -iter 200000 -md sha256 -salt \
  -pass stdin \
  -in "$staging_dir/payload.tar.gz" \
  -out "$partial_artifact" < "$key_file"
mv -f "$partial_artifact" "$artifact_path"
partial_artifact=""
(
  cd "$tier_dir"
  sha256sum "$artifact_name" > "$artifact_name.sha256.partial"
  mv -f "$artifact_name.sha256.partial" "$artifact_name.sha256"
)

mapfile -t retained_artifacts < <(
  for candidate in "$tier_dir"/cashlenx-*.tar.gz.enc; do
    [[ -f "$candidate" ]] || continue
    basename "$candidate"
  done | sort -r
)
for (( index=retention; index<${#retained_artifacts[@]}; index++ )); do
  expired="${retained_artifacts[$index]}"
  rm -f -- "$tier_dir/$expired" "$tier_dir/$expired.sha256"
done

echo "Encrypted backup created: $artifact_path"
