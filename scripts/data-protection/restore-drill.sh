#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

project_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"
cd "$project_dir"

backup_path="${1:-}"
env_file=""
temporary_dir=""
container_name=""
started_at_epoch="$(date +%s)"

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

cleanup() {
  if [[ -n "$container_name" ]]; then
    docker rm -f "$container_name" >/dev/null 2>&1 || true
  fi
  [[ -z "$temporary_dir" ]] || rm -rf -- "$temporary_dir"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM HUP

[[ -n "$backup_path" ]] || {
  echo "Usage: $0 <encrypted-backup-path>" >&2
  exit 2
}

for command_name in docker openssl tar sha256sum realpath awk; do
  command -v "$command_name" >/dev/null 2>&1 || {
    echo "Required command is unavailable: $command_name" >&2
    exit 1
  }
done

env_file="$(resolve_env_file)"
backup_path="$(realpath "$backup_path")"
[[ -f "$backup_path" && -s "$backup_path" ]] || {
  echo "Backup artifact must be a non-empty regular file." >&2
  exit 1
}
checksum_path="$backup_path.sha256"
[[ -f "$checksum_path" ]] || { echo "Backup checksum sidecar is missing." >&2; exit 1; }
(
  cd "$(dirname "$backup_path")"
  sha256sum -c "$(basename "$checksum_path")"
)

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

temporary_dir="$(mktemp -d)"
openssl enc -d -aes-256-cbc -pbkdf2 -iter 200000 -md sha256 \
  -pass stdin \
  -in "$backup_path" \
  -out "$temporary_dir/payload.tar.gz" < "$key_file"

if tar -tzf "$temporary_dir/payload.tar.gz" | awk '
  /^\// || /(^|\/)\.\.($|\/)/ { unsafe = 1 }
  END { exit unsafe ? 0 : 1 }
'; then
  echo "Backup archive contains an unsafe path." >&2
  exit 1
fi
mkdir -p "$temporary_dir/content"
tar -xzf "$temporary_dir/payload.tar.gz" -C "$temporary_dir/content"
(
  cd "$temporary_dir/content"
  sha256sum -c manifest.sha256
)

metadata="$temporary_dir/content/metadata.env"
database_type="$(awk -F= '$1 == "DATABASE_TYPE" { print $2 }' "$metadata")"
database_name="$(awk -F= '$1 == "DATABASE_NAME" { print $2 }' "$metadata")"
format_version="$(awk -F= '$1 == "FORMAT_VERSION" { print $2 }' "$metadata")"
[[ "$format_version" == "1" ]] || { echo "Unsupported backup format." >&2; exit 1; }
[[ "$database_name" =~ ^[A-Za-z0-9_]+$ ]] || { echo "Backup database name is invalid." >&2; exit 1; }

suffix="$(date -u +%Y%m%d%H%M%S)-$$-$RANDOM"
drill_password="$(openssl rand -hex 24)"
container_name="cashlenx-restore-drill-${database_type}-${suffix}"

wait_for_command() {
  local attempts="$1"
  shift
  for (( attempt=1; attempt<=attempts; attempt++ )); do
    if "$@" >/dev/null 2>&1; then
      return 0
    fi
    sleep 2
  done
  return 1
}

case "$database_type" in
  mongodb)
    image="$(awk -F= '$1 == "MONGO_IMAGE" { print $2 }' docker/dependencies/images.env)"
    [[ "$image" =~ ^mongo:[A-Za-z0-9_.-]+@sha256:[0-9a-f]{64}$ ]] || {
      echo "MONGO_IMAGE must be digest-pinned in docker/dependencies/images.env." >&2
      exit 1
    }
    docker image inspect "$image" >/dev/null 2>&1 || {
      echo "Restore-drill image is not available locally: $image" >&2
      exit 1
    }
    docker run --detach --rm --network none --name "$container_name" \
      --env MONGO_INITDB_ROOT_USERNAME=drill \
      --env MONGO_INITDB_ROOT_PASSWORD="$drill_password" \
      --env MONGO_INITDB_DATABASE="$database_name" \
      "$image" >/dev/null
    wait_for_command 60 docker exec "$container_name" mongosh --quiet \
      --username drill --password "$drill_password" --authenticationDatabase admin \
      --eval 'quit(db.runCommand({ ping: 1 }).ok ? 0 : 1)' "127.0.0.1:27017/$database_name" || {
        echo "Disposable MongoDB did not become ready." >&2
        exit 1
      }
    docker cp "$temporary_dir/content/database.archive.gz" "$container_name:/tmp/database.archive.gz"
    docker exec "$container_name" sh -ec '
      exec mongorestore \
        --host 127.0.0.1 \
        --port 27017 \
        --username drill \
        --password "$MONGO_INITDB_ROOT_PASSWORD" \
        --authenticationDatabase admin \
        --archive=/tmp/database.archive.gz \
        --gzip \
        --drop
    '
    object_count="$(docker exec "$container_name" mongosh --quiet \
      --username drill --password "$drill_password" --authenticationDatabase admin \
      --eval "const d=db.getSiblingDB('$database_name'); const n=d.getCollectionNames(); if (!n.includes('schema_migrations')) quit(3); print(n.length)" \
      "127.0.0.1:27017/$database_name" | tail -n 1 | tr -d '\r')"
    ;;
  mysql)
    image="$(awk -F= '$1 == "MYSQL_IMAGE" { print $2 }' docker/dependencies/images.env)"
    [[ "$image" =~ ^mysql:[A-Za-z0-9_.-]+@sha256:[0-9a-f]{64}$ ]] || {
      echo "MYSQL_IMAGE must be digest-pinned in docker/dependencies/images.env." >&2
      exit 1
    }
    docker image inspect "$image" >/dev/null 2>&1 || {
      echo "Restore-drill image is not available locally: $image" >&2
      exit 1
    }
    docker run --detach --rm --network none --name "$container_name" \
      --tmpfs /var/lib/mysql:rw,noexec,nosuid,size=1g \
      --env MYSQL_ROOT_PASSWORD="$drill_password" \
      --env MYSQL_DATABASE="$database_name" \
      "$image" >/dev/null
    wait_for_command 120 docker exec --env MYSQL_PWD="$drill_password" "$container_name" \
      mysqladmin ping --host=127.0.0.1 --user=root --silent || {
        echo "Disposable MySQL did not become ready." >&2
        exit 1
      }
    docker exec -i --env MYSQL_PWD="$drill_password" "$container_name" \
      mysql --host=127.0.0.1 --user=root "$database_name" < "$temporary_dir/content/database.sql"
    object_count="$(docker exec --env MYSQL_PWD="$drill_password" "$container_name" \
      mysql --batch --skip-column-names --host=127.0.0.1 --user=root "$database_name" \
      --execute="SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = '$database_name'; SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = '$database_name' AND table_name = 'schema_migrations';" | tr -d '\r')"
    table_count="$(printf '%s\n' "$object_count" | sed -n '1p')"
    migration_table_count="$(printf '%s\n' "$object_count" | sed -n '2p')"
    [[ "$migration_table_count" == "1" ]] || { echo "Restored MySQL data lacks schema_migrations." >&2; exit 1; }
    object_count="$table_count"
    ;;
  *) echo "Unsupported backup database type: $database_type" >&2; exit 1 ;;
esac

[[ "$object_count" =~ ^[1-9][0-9]*$ ]] || {
  echo "Restore drill found no database objects." >&2
  exit 1
}

evidence_dir_setting="$(read_env_value RESTORE_DRILL_EVIDENCE_DIR)"
if [[ -n "$evidence_dir_setting" ]]; then
  evidence_dir="$evidence_dir_setting"
  [[ "$evidence_dir" == /* ]] || evidence_dir="$project_dir/$evidence_dir"
else
  evidence_dir="$(dirname "$(dirname "$backup_path")")/restore-drills"
fi
mkdir -p "$evidence_dir"
evidence_dir="$(realpath "$evidence_dir")"
evidence_path="$evidence_dir/restore-drill-${database_type}-${suffix}.json"
duration_seconds=$(( $(date +%s) - started_at_epoch ))
artifact_sha256="$(sha256sum "$backup_path" | awk '{ print $1 }')"
printf '{"result":"passed","database":"%s","database_name":"%s","artifact_sha256":"%s","object_count":%s,"duration_seconds":%s,"completed_at":"%s"}\n' \
  "$database_type" "$database_name" "$artifact_sha256" "$object_count" "$duration_seconds" "$(date -u +%Y-%m-%dT%H:%M:%SZ)" > "$evidence_path"

echo "Disposable restore drill passed: $evidence_path"
