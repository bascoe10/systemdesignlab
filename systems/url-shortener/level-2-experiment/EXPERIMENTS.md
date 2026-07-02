# Experiments — URL Shortener

Run each experiment, predict the outcome BEFORE redeploying, then compare.
`make validate` checks that you actually changed the config at least once.

## Experiment 1: Switch Cache Provider

Change `cache.provider` from `redis` to `memcached` (or run
`make switch-cache PROVIDER=memcached`), then `make redeploy && make load-test`.

Expected outcome:
- Hit rate stays similar — the hash ring doesn't care which backend it shards.
- The Saturation panel changes shape: Memcached pre-allocates slab memory,
  so "used" memory behaves differently from Redis's incremental growth.
- Note: `max_memory` and `eviction_policy` in config.yaml only apply to
  Redis (set at runtime via CONFIG SET). Memcached's limit is fixed at
  startup — a real operational difference between the two, not a lab quirk.

## Experiment 2: Shrink the Cache

Set `cache.max_memory: 16mb`. Run `make redeploy && make load-test SCENARIO=hot-key`.

Expected outcome:
- Eviction rate spikes (Component Deep Dive → "Evictions per cache node").
- Hit rate dips but stays surprisingly decent — LRU keeps the 10 hot keys
  resident while sacrificing the long tail. That asymmetry IS the argument
  for LRU on read-heavy workloads.

## Experiment 3: Break the Ring Balance

Set `hashing.virtual_nodes: 1`. Redeploy, run a steady load test, and watch
"Consistent hashing — ops per cache node".

Expected outcome:
- With one virtual node per cache node, the ring has 3 points on it and the
  arc sizes are wildly uneven — one node takes most of the keys.
- Raise it back through 8 → 32 → 128 and watch the lines converge. This is
  the intuition you'll need when you BUILD the ring at Level 3.

## Experiment 4: Starve the DB Pool

Set `database.pool_size: 2`. Redeploy, then run
`make load-test SCENARIO=read-spike`.

Expected outcome:
- p99 climbs while p50 barely moves: only cache-miss requests queue on the
  pool, and they're the tail. The "Saturation — DB connection pool" panel
  pins at 2/2.
- This is the classic hidden bottleneck: no errors, healthy-looking
  averages, terrible tail.

## Experiment 5: Rate-limit Yourself

Set `gateway.rate_limit_rps: 500`, redeploy, run a steady load test at the
default 1000 RPS.

Expected outcome:
- 429s appear in the Errors panel (tracked separately from 5xx).
- Throughput flattens at ~500 RPS. Where would you rather shed load — at
  the gateway, or by letting the database fall over? That's the trade-off
  rate limiters exist to make explicit.

---

When done, restore the healthy config (`git checkout -- config.yaml`),
run `make validate`, and move on: `git checkout level-3-build/url-shortener`
