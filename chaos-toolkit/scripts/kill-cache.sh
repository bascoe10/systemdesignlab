#!/usr/bin/env bash
# Kill every cache node for OUTAGE seconds (default 60), then restore.
# Keep a load test running in another terminal and watch the fallback.
# shellcheck source=chaos-toolkit/scripts/_lib.sh
source "$(dirname "${BASH_SOURCE[0]}")/_lib.sh"

OUTAGE="${OUTAGE:-60}"
banner "killing all cache nodes for ${OUTAGE}s"

services="$(cache_services)"
if [[ -z "$services" ]]; then
  echo "no cache containers found — is the stack running? (make start)" >&2
  exit 1
fi

# shellcheck disable=SC2086
dc stop $services

echo "Cache is DOWN. Questions while you wait:"
echo "  - Did the error rate move, or only latency? Why?"
echo "  - Where is the read traffic going now? (Deep Dive → DB queries/sec)"

for ((i=OUTAGE; i>0; i-=10)); do
  echo "  restoring in ${i}s..."
  sleep 10
done

# shellcheck disable=SC2086
dc start $services
echo
echo "Cache restored. Now watch the recovery: how long until the hit rate"
echo "is back above 85%? See chaos-toolkit/resilience-challenges/KILL_CACHE.md"
