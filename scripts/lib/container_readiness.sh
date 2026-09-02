#!/usr/bin/env bash

# Keep startup readiness independent from Compose-managed health status so the
# lifecycle works with Docker Compose and compatible frontends such as nerdctl.
wait_for_container_command() {
  local container_name="$1"
  shift

  local timeout_seconds=180
  local deadline=$((SECONDS + timeout_seconds))
  local status

  while ((SECONDS < deadline)); do
    status="$(docker inspect --format '{{.State.Status}}' "$container_name" 2>/dev/null || true)"
    case "$status" in
      exited | dead)
        echo "Container stopped before becoming ready: $container_name" >&2
        docker logs --tail 50 "$container_name" >&2 || true
        return 1
        ;;
    esac

    if docker exec "$container_name" "$@" >/dev/null 2>&1; then
      echo "Container is ready: $container_name"
      return 0
    fi
    sleep 2
  done

  echo "Timed out after ${timeout_seconds}s waiting for container: $container_name" >&2
  docker logs --tail 50 "$container_name" >&2 || true
  return 1
}
