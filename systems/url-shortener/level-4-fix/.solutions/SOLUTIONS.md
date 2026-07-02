# Solutions — URL Shortener Level 4

Revealed via `sdl reveal-solution`. If you're reading this before fixing
the system yourself, you're stealing the learning from future-you.

All four root causes live in `config.yaml`.

## Root Cause 1: Cache TTL of 1 second → Symptom 1 (terrible hit rate)

`cache.ttl: 1s`. Every cached entry expires almost immediately, so even
hot URLs are gone before the next read. The write-through and backfill
logic all "work" — the data just evaporates.

**Fix:** `ttl: 24h`.
**How you could tell:** hit rate near zero while "Cache read outcomes"
shows plenty of successful `set` operations, and cache node item counts
stay near zero despite constant writes. Data going in, nothing staying.

## Root Cause 2: DB pool of 2 → Symptom 2 (high p99, normal-ish p50)

`database.pool_size: 2`. With the cache useless (root cause 1), nearly all
reads take the database path and queue on two connections. The queueing
shows up as tail latency, not errors.

**Fix:** `pool_size: 10`.
**How you could tell:** "Saturation — DB connection pool" pinned at max
while DB query p99 (execution time) stays low — time is spent WAITING for
a connection, not running the query. Classic saturation-without-errors.

## Root Cause 3: virtual_nodes 1 → Symptom 3 (uneven distribution)

`hashing.virtual_nodes: 1`. Each cache node lands at a single point on the
ring, so arc sizes are wildly uneven and one node owns most of the keyspace.
You built this ring in Level 3 — you know virtual nodes are what smooth it.

**Fix:** `virtual_nodes: 128`.

## Root Cause 4: 8mb cache memory → Symptom 4 (memory pressure)

`cache.max_memory: 8mb`. The working set doesn't fit, so Redis sits at its
limit and continuously evicts. With root cause 1 also present the item
count stays low anyway — after fixing the TTL, this one becomes the next
bottleneck. Layered misconfigurations masking each other is deliberate.

**Fix:** `max_memory: 256mb`.

## Verify

```bash
sdl restart && sdl load
sdl validate
```

Hit rate > 85%, p99 < 100ms, three even node lines, evictions ~0.
