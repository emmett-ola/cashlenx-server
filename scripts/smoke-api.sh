#!/usr/bin/env bash
set -euo pipefail

RUN_ID="$(date +%s)-${RANDOM}"
MANAGE_MONGODB="${SMOKE_MANAGE_MONGODB:-false}"
MANAGE_SERVER="${SMOKE_MANAGE_SERVER:-false}"
MONGO_CONTAINER_NAME="${SMOKE_MONGO_CONTAINER_NAME:-cashlenx-smoke-mongodb-${RUN_ID}}"
MONGO_IMAGE="${SMOKE_MONGO_IMAGE:-mongo:7.0}"
MONGO_ROOT_USERNAME="${SMOKE_MONGO_ROOT_USERNAME:-cashlenx}"
MONGO_ROOT_PASSWORD="${SMOKE_MONGO_ROOT_PASSWORD:-cashlenx123}"
SMOKE_DB_NAME="${SMOKE_DB_NAME:-cashlenx_smoke_${RUN_ID//-/_}}"
SERVER_PORT="${SMOKE_SERVER_PORT:-18080}"
API_VERSION="${API_VERSION:-v0}"
BASE_URL="${BASE_URL:-}"
ADMIN_USERNAME="${ADMIN_USERNAME:-admin}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-admin}"
USERNAME="${SMOKE_USERNAME:-beta_smoke_${RUN_ID}}"
PASSWORD="${SMOKE_PASSWORD:-SmokePass123!}"
NEW_PASSWORD="${SMOKE_NEW_PASSWORD:-SmokePass456!}"
CATEGORY_NAME="beta_smoke_expense_${RUN_ID}"
TMP_DIR="$(mktemp -d)"
RESP_FILE="${TMP_DIR}/response.json"
SERVER_PID=""

usage() {
  cat <<'EOF'
Usage: scripts/smoke-api.sh [--with-mongodb] [--managed]

Runs the CashLenX API smoke flow.

Default:
  Smoke an already-running API at BASE_URL.

Options:
  --with-mongodb  Start a disposable MongoDB container and remove it on exit.
                  Use this when you will start the API separately with the
                  printed MONGO_DB_URI.
  --managed       Start disposable MongoDB and a local API server, then remove
                  MongoDB and stop the server on exit.

Useful env:
  BASE_URL, SMOKE_SERVER_PORT, API_VERSION
  SMOKE_MONGO_IMAGE, SMOKE_MONGO_CONTAINER_NAME
  SMOKE_MONGO_ROOT_USERNAME, SMOKE_MONGO_ROOT_PASSWORD, SMOKE_DB_NAME
  ADMIN_USERNAME, ADMIN_PASSWORD
  SMOKE_ALLOW_REMOTE=true to allow a non-local BASE_URL
EOF
}

for arg in "$@"; do
  case "$arg" in
    --with-mongodb)
      MANAGE_MONGODB=true
      ;;
    --managed)
      MANAGE_MONGODB=true
      MANAGE_SERVER=true
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "unknown argument: $arg" >&2
      usage >&2
      exit 1
      ;;
  esac
done

if [[ -z "$BASE_URL" ]]; then
  if [[ "$MANAGE_SERVER" == "true" ]]; then
    BASE_URL="http://localhost:${SERVER_PORT}/api/${API_VERSION}"
  else
    BASE_URL="http://localhost:8080/api/${API_VERSION}"
  fi
fi

if [[ "${SMOKE_ALLOW_REMOTE:-false}" != "true" ]]; then
  case "$BASE_URL" in
    http://localhost:*|http://127.0.0.1:*|http://0.0.0.0:*)
      ;;
    *)
      echo "refusing to run smoke flow against non-local BASE_URL: ${BASE_URL}" >&2
      echo "set SMOKE_ALLOW_REMOTE=true only for an intentionally disposable remote test environment" >&2
      exit 1
      ;;
  esac
fi

cleanup() {
  if [[ -n "$SERVER_PID" ]]; then
    echo "Stopping smoke API server..."
    kill "$SERVER_PID" >/dev/null 2>&1 || true
    wait "$SERVER_PID" >/dev/null 2>&1 || true
  fi
  if [[ "$MANAGE_MONGODB" == "true" ]]; then
    echo "Removing smoke MongoDB container ${MONGO_CONTAINER_NAME}..."
    docker rm -f "$MONGO_CONTAINER_NAME" >/dev/null 2>&1 || true
  fi
  rm -rf "${TMP_DIR}"
}
trap cleanup EXIT

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing required command: $1" >&2
    exit 1
  fi
}

need_cmd curl
need_cmd python3

if [[ "$MANAGE_MONGODB" == "true" ]]; then
  need_cmd docker
fi

if [[ "$MANAGE_SERVER" == "true" ]]; then
  need_cmd go
fi

start_mongodb() {
  echo "Starting disposable MongoDB container ${MONGO_CONTAINER_NAME}..."
  docker run -d \
    --name "$MONGO_CONTAINER_NAME" \
    -p 127.0.0.1::27017 \
    -e MONGO_INITDB_ROOT_USERNAME="$MONGO_ROOT_USERNAME" \
    -e MONGO_INITDB_ROOT_PASSWORD="$MONGO_ROOT_PASSWORD" \
    -e MONGO_INITDB_DATABASE="$SMOKE_DB_NAME" \
    "$MONGO_IMAGE" >/dev/null

  local mapped_port
  mapped_port="$(docker port "$MONGO_CONTAINER_NAME" 27017/tcp | sed 's/.*://')"
  if [[ -z "$mapped_port" ]]; then
    echo "failed to determine smoke MongoDB host port" >&2
    exit 1
  fi

  export DB_TYPE="mongodb"
  export DB_NAME="$SMOKE_DB_NAME"
  export MONGO_DB_URI="mongodb://${MONGO_ROOT_USERNAME}:${MONGO_ROOT_PASSWORD}@localhost:${mapped_port}/${SMOKE_DB_NAME}?authSource=admin&retryWrites=false"

  echo "Waiting for smoke MongoDB on localhost:${mapped_port}..."
  for _ in {1..60}; do
    if docker exec "$MONGO_CONTAINER_NAME" mongosh --quiet --eval 'db.adminCommand({ ping: 1 }).ok' >/dev/null 2>&1; then
      echo "Smoke MongoDB is ready."
      echo "MONGO_DB_URI=${MONGO_DB_URI}"
      return
    fi
    sleep 1
  done

  echo "smoke MongoDB did not become ready in time" >&2
  docker logs "$MONGO_CONTAINER_NAME" >&2 || true
  exit 1
}

start_server() {
  echo "Starting local smoke API server on port ${SERVER_PORT}..."
  export ENV="${ENV:-test}"
  export SERVER_HOST="${SERVER_HOST:-127.0.0.1}"
  export SERVER_PORT
  export API_VERSION
  export SCHEMA_VALIDATION="${SCHEMA_VALIDATION:-true}"
  export JWT_SECRET="${JWT_SECRET:-smoke-test-secret-change-me}"
  export AUTH_REGISTRATION_ENABLED="${AUTH_REGISTRATION_ENABLED:-true}"
  export ADMIN_USERNAME
  export ADMIN_PASSWORD

  go run main.go open start -p "$SERVER_PORT" >"${TMP_DIR}/server.log" 2>&1 &
  SERVER_PID="$!"

  echo "Waiting for smoke API at ${BASE_URL}/open/health..."
  for _ in {1..90}; do
    if curl -fsS "${BASE_URL}/open/health" >/dev/null 2>&1; then
      echo "Smoke API server is ready."
      return
    fi
    if ! kill -0 "$SERVER_PID" >/dev/null 2>&1; then
      echo "smoke API server exited before becoming ready" >&2
      cat "${TMP_DIR}/server.log" >&2 || true
      exit 1
    fi
    sleep 1
  done

  echo "smoke API server did not become ready in time" >&2
  cat "${TMP_DIR}/server.log" >&2 || true
  exit 1
}

json_get() {
  local path="$1"
  python3 - "$RESP_FILE" "$path" <<'PY'
import json
import sys

with open(sys.argv[1], "r", encoding="utf-8") as f:
    data = json.load(f)

current = data
for part in sys.argv[2].split("."):
    if isinstance(current, dict) and part in current:
        current = current[part]
    else:
        print("")
        sys.exit(0)

if current is None:
    print("")
elif isinstance(current, (dict, list)):
    print(json.dumps(current))
else:
    print(current)
PY
}

assert_json_ok() {
  python3 - "$RESP_FILE" <<'PY'
import json
import sys

with open(sys.argv[1], "r", encoding="utf-8") as f:
    data = json.load(f)

if "code" not in data:
    raise SystemExit("response missing code field")
PY
}

api() {
  local method="$1"
  local path="$2"
  local expected="$3"
  local body="${4:-}"
  local token="${5:-}"
  local url="${BASE_URL}${path}"
  local request_id="smoke-${RUN_ID}"
  local status
  local args=(-sS -X "$method" "$url" -o "$RESP_FILE" -w "%{http_code}" -H "Accept: application/json" -H "X-Request-ID: ${request_id}")

  if [[ -n "$token" ]]; then
    args+=(-H "Authorization: Bearer ${token}")
  fi
  if [[ -n "$body" ]]; then
    args+=(-H "Content-Type: application/json" -d "$body")
  fi

  status="$(curl "${args[@]}")"
  if [[ "$status" != "$expected" ]]; then
    echo "FAILED ${method} ${path}: HTTP ${status}, expected ${expected}" >&2
    cat "$RESP_FILE" >&2 || true
    exit 1
  fi

  assert_json_ok
  echo "OK ${method} ${path} -> ${status}"
}

download() {
  local path="$1"
  local expected="$2"
  local token="$3"
  local output="$4"
  local status

  status="$(curl -sS -X GET "${BASE_URL}${path}" -o "$output" -w "%{http_code}" -H "Authorization: Bearer ${token}" -H "X-Request-ID: smoke-${RUN_ID}")"
  if [[ "$status" != "$expected" ]]; then
    echo "FAILED GET ${path}: HTTP ${status}, expected ${expected}" >&2
    exit 1
  fi
  if [[ ! -s "$output" ]]; then
    echo "FAILED GET ${path}: downloaded file is empty" >&2
    exit 1
  fi
  echo "OK GET ${path} -> ${status}"
}

if [[ "$MANAGE_MONGODB" == "true" ]]; then
  start_mongodb
fi

if [[ "$MANAGE_SERVER" == "true" ]]; then
  start_server
fi

echo "Running CashLenX API smoke flow against ${BASE_URL}"

api GET "/open/health" 200
api GET "/open/version" 200
api POST "/open/auth/logout" 200

api POST "/open/auth/login" 200 "{\"username\":\"${ADMIN_USERNAME}\",\"password\":\"${ADMIN_PASSWORD}\",\"device_id\":\"smoke-admin-setup\",\"device_name\":\"Smoke Script\"}"
ADMIN_ACCESS_TOKEN="$(json_get "data.access_token")"
if [[ -z "$ADMIN_ACCESS_TOKEN" ]]; then
  echo "admin setup login response did not include access_token" >&2
  cat "$RESP_FILE" >&2
  exit 1
fi
api POST "/admin/user" 201 "{\"username\":\"${USERNAME}\",\"password\":\"${PASSWORD}\",\"email_address\":\"${USERNAME}@example.test\",\"is_email_verified\":true}" "$ADMIN_ACCESS_TOKEN"

api POST "/open/auth/login" 200 "{\"username\":\"${USERNAME}\",\"password\":\"${PASSWORD}\",\"device_id\":\"smoke\",\"device_name\":\"Smoke Script\"}"
ACCESS_TOKEN="$(json_get "data.access_token")"
REFRESH_TOKEN="$(json_get "data.refresh_token")"
if [[ -z "$ACCESS_TOKEN" || -z "$REFRESH_TOKEN" ]]; then
  echo "login response did not include access_token and refresh_token" >&2
  cat "$RESP_FILE" >&2
  exit 1
fi

api POST "/open/auth/login" 200 "{\"refresh_token\":\"${REFRESH_TOKEN}\",\"device_id\":\"smoke\",\"device_name\":\"Smoke Script\"}"
ACCESS_TOKEN="$(json_get "data.access_token")"
REFRESH_TOKEN="$(json_get "data.refresh_token")"

api GET "/user/profile" 200 "" "$ACCESS_TOKEN"
api PUT "/user/profile" 200 "{\"nickname\":\"Beta Smoke\",\"gender\":\"other\"}" "$ACCESS_TOKEN"

api POST "/category" 201 "{\"name\":\"${CATEGORY_NAME}\",\"type\":\"expense\",\"remark\":\"smoke\"}" "$ACCESS_TOKEN"
api GET "/category?type=expense&limit=10&offset=0" 200 "" "$ACCESS_TOKEN"
api GET "/category/tree?type=expense" 200 "" "$ACCESS_TOKEN"

api POST "/cash/expense" 201 "{\"belongs_date\":\"2026-01-15\",\"category_name\":\"${CATEGORY_NAME}\",\"amount\":12.34,\"description\":\"smoke expense\"}" "$ACCESS_TOKEN"
api GET "/cash?limit=10&offset=0" 200 "" "$ACCESS_TOKEN"
api GET "/cash/range?from=2026-01-01&to=2026-01-31" 200 "" "$ACCESS_TOKEN"
api GET "/cash/summary/daily/2026-01-15" 200 "" "$ACCESS_TOKEN"

api GET "/statistic/summary/daily/2026-01-15" 200 "" "$ACCESS_TOKEN"
api GET "/statistic/breakdown/monthly/202601" 200 "" "$ACCESS_TOKEN"
api GET "/statistic/top/monthly/202601" 200 "" "$ACCESS_TOKEN"
api GET "/statistic/dashboard/monthly/202601" 200 "" "$ACCESS_TOKEN"
download "/statistic/export?from_date=2026-01-01&to_date=2026-01-31&format=csv" 200 "$ACCESS_TOKEN" "${TMP_DIR}/export.csv"
download "/user/database/backup" 200 "$ACCESS_TOKEN" "${TMP_DIR}/user-backup.json"

api PUT "/user/password" 200 "{\"old_password\":\"${PASSWORD}\",\"new_password\":\"${NEW_PASSWORD}\"}" "$ACCESS_TOKEN"
api POST "/open/auth/login" 200 "{\"username\":\"${USERNAME}\",\"password\":\"${NEW_PASSWORD}\",\"device_id\":\"smoke\",\"device_name\":\"Smoke Script\"}"
ACCESS_TOKEN="$(json_get "data.access_token")"
REFRESH_TOKEN="$(json_get "data.refresh_token")"
api POST "/open/auth/logout" 200 "{\"refresh_token\":\"${REFRESH_TOKEN}\"}"
api DELETE "/user/account" 200 "" "$ACCESS_TOKEN"

api POST "/open/auth/login" 200 "{\"username\":\"${ADMIN_USERNAME}\",\"password\":\"${ADMIN_PASSWORD}\",\"device_id\":\"smoke-admin\",\"device_name\":\"Smoke Script\"}"
ADMIN_ACCESS_TOKEN="$(json_get "data.access_token")"
if [[ -z "$ADMIN_ACCESS_TOKEN" ]]; then
  echo "admin login response did not include access_token" >&2
  cat "$RESP_FILE" >&2
  exit 1
fi
api GET "/admin/user?limit=5&offset=0" 200 "" "$ADMIN_ACCESS_TOKEN"
download "/admin/database/backup" 200 "$ADMIN_ACCESS_TOKEN" "${TMP_DIR}/admin-backup.json"

echo "Smoke flow completed successfully."
