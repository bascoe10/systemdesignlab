# Resilience Challenge: Database Latency

## What Gets Injected

`make chaos-lag-database` — 300ms of network latency on every database
round-trip (`DELAY=500ms` to change), via `tc netem` in the DB container's
network namespace. Undo with `make chaos-restore`.

Note: needs the Linux `netem` kernel module. On Docker Desktop the script
prints a manual fallback if injection fails.

## What to Look For in Grafana

- **p50 vs p99 split:** cache hits never touch the database, so p50 barely
  moves while p99 absorbs the full 300ms+. The cache is your shock absorber
  — this experiment measures exactly how much it absorbs.
- **DB pool saturation:** each query now holds a connection ~300ms longer.
  Watch "Saturation — DB connection pool" — at high load the pool becomes
  the next bottleneck (queueing on top of latency).
- **Write path:** every `POST /api/shorten` pays the toll. Compare the
  `shorten` route latency against `redirect`.

## The Fix

<details>
<summary>Click to reveal</summary>

You can't fix the network from the application — you contain it:

- The cache hit rate is the main defense: at 90% hits, only 10% of reads
  pay the latency. Compute effective mean latency to convince yourself.
- Bound the damage: request timeouts and a properly-sized pool prevent the
  slow path from starving the fast path.
- Real-world escalations: read replicas closer to the service, or a
  circuit breaker that sheds the slow dependency when SLOs are burning.
</details>

## Success Criteria

| Metric | During Lag | Notes |
|--------|-----------|-------|
| P50 latency | < 20ms | proves cache hits bypass the lag |
| Error rate | < 1% | slow ≠ failed |
| Pool saturation | watched & explained | journal the mechanism |
