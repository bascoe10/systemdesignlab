# Resilience Challenge: Cache Failure

## What Gets Injected

`make chaos-kill-cache` — all cache nodes stopped for 60 seconds
(`OUTAGE=120 make chaos-kill-cache` to change), then restarted.
Run it with a load test going: `make load-test` in another terminal.

## What to Look For in Grafana

- **Golden Signals Overview:** latency spikes; traffic unchanged; the
  interesting question is the error rate — a resilient read path shows
  ~zero errors because reads fall back to the database.
- **Component Deep Dive:** cache read outcomes flip from `hit` to `error`,
  DB queries/sec spikes to match total read traffic.
- **Recovery:** after restore, the cache is EMPTY. Watch the hit rate climb
  from 0% — that's every miss re-warming the cache via backfill.

## The Fix

<details>
<summary>Click to reveal (try to reason it out first)</summary>

The shipped system already implements the core pattern — this challenge is
about *verifying* it and knowing its name:

- Cache-aside with graceful fallback: on cache error, treat as a miss and
  serve from the database; backfill on recovery.
- Client timeouts (500ms here) so a dead cache costs bounded time, not a
  hung request.
- Stretch (Level 5 territory): add a circuit breaker — after N consecutive
  cache failures, stop attempting the cache for a cooldown period and skip
  the timeout cost entirely. Compare p99 during the outage with and
  without it.
</details>

## Success Criteria

| Metric | During Outage | After Recovery |
|--------|--------------|----------------|
| Error rate | < 1% | < 0.1% |
| P99 latency | < 500ms (DB-only path) | < 100ms |
| Cache hit rate | 0% (expected) | > 85% within 60s |
