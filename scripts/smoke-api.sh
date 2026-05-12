#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080/api/v0}"
ADMIN_USERNAME="${ADMIN_USERNAME:-admin}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-admin}"
RUN_ID="$(date +%s)-${RANDOM}"
USERNAME="${SMOKE_USERNAME:-beta_smoke_${RUN_ID}}"
PASSWORD="${SMOKE_PASSWORD:-SmokePass123!}"
NEW_PASSWORD="${SMOKE_NEW_PASSWORD:-SmokePass456!}"
CATEGORY_NAME="beta_smoke_expense_${RUN_ID}"
TMP_DIR="$(mktemp -d)"
RESP_FILE="${TMP_DIR}/response.json"

cleanup() {
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

echo "Running CashLenX API smoke flow against ${BASE_URL}"

api GET "/open/health" 200
api GET "/open/version" 200
api POST "/open/auth/logout" 200

api POST "/open/auth/register" 201 "{\"username\":\"${USERNAME}\",\"password\":\"${PASSWORD}\"}"
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
