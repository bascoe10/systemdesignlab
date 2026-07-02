#!/usr/bin/env bash
# Inject DELAY (default 300ms) of network latency on the database container
# using tc netem, run from a sidecar container in the DB's network namespace.
# Undo with: make chaos-restore
source "$(dirname "${BASH_SOURCE[0]}")/_lib.sh"

DELAY="${DELAY:-300ms}"
banner "adding ${DELAY} latency to the database"

db_id="$(dc ps -q db)"
if [[ -z "$db_id" ]]; then
  echo "database container not running — make start first" >&2
  exit 1
fi

if docker run --rm --network "container:${db_id}" --cap-add NET_ADMIN \
  nicolaka/netshoot tc qdisc add dev eth0 root netem delay "$DELAY"; then
  echo
  echo "Done. Every DB round-trip now costs an extra ${DELAY}."
  echo "  - Which requests feel it: cache hits or misses? Check p50 vs p99."
  echo "  - How well does the cache absorb it? Compare hit rate to latency."
  echo "Restore with: make chaos-restore"
else
  cat >&2 <<'EOF'

Could not inject latency. tc netem needs the Linux kernel netem module;
on Docker Desktop (Mac/Windows) this may be unavailable.

Manual fallback: add a Toxiproxy container between the services and the
database, or simulate the effect by setting database.pool_size: 1 in
config.yaml and re-running the load test (queueing produces a similar
tail-latency signature).
EOF
  exit 1
fi
