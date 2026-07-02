#!/usr/bin/env bash
# Shared helpers for chaos scripts. Scripts run from a level branch root
# (or an assembled .lab directory) where system/docker-compose.yml exists.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
COMPOSE_FILE="$ROOT/system/docker-compose.yml"

if [[ ! -f "$COMPOSE_FILE" ]]; then
  echo "error: $COMPOSE_FILE not found — run chaos commands from a level branch (make chaos-*)" >&2
  exit 1
fi

CACHE_PROVIDER="$(awk '/^cache:/{f=1} f && /provider:/{print $2; exit}' "$ROOT/config.yaml")"

dc() {
  COMPOSE_PROFILES="$CACHE_PROVIDER" docker compose -f "$COMPOSE_FILE" "$@"
}

cache_services() {
  dc ps --services --all | grep -E "^(redis|memcached)-" || true
}

banner() {
  echo
  echo "=== CHAOS: $1 ==="
  echo "Watch it happen: Grafana → Chaos Impact dashboard (http://localhost:3000)"
  echo
}
