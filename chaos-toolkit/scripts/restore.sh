#!/usr/bin/env bash
# Undo all chaos: remove injected network latency, restart stopped cache
# nodes, and make sure every service is up.
# shellcheck source=chaos-toolkit/scripts/_lib.sh
source "$(dirname "${BASH_SOURCE[0]}")/_lib.sh"

banner "restoring healthy state"

db_id="$(dc ps -q db || true)"
if [[ -n "$db_id" ]]; then
  docker run --rm --network "container:${db_id}" --cap-add NET_ADMIN \
    nicolaka/netshoot tc qdisc del dev eth0 root 2>/dev/null \
    && echo "removed DB network latency" \
    || echo "no injected DB latency found (fine)"
fi

services="$(cache_services)"
if [[ -n "$services" ]]; then
  # shellcheck disable=SC2086
  dc start $services
fi
dc up -d

echo
echo "All services running. Verify recovery on the Chaos Impact dashboard:"
echo "how long does each signal take to return to baseline?"
