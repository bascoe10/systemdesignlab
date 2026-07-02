# URL Shortener

A t.co-style URL shortener: an API gateway in front of a write service
(shortener) and a read-optimized service (redirector), backed by PostgreSQL
and a 3-node cache cluster sharded by consistent hashing.

```
Create:   Client → api-gateway → shortener  → Postgres (write)
                                            → cache node via hash ring (write-through)

Redirect: Client → api-gateway → redirector → cache node via hash ring (hit → 302)
                                            → miss → Postgres → async backfill → 302
```

Concepts taught: consistent hashing, write-through vs cache-aside, read-heavy
optimization, cache eviction policies, hot-key behaviour.

## Where Do I Start?

### Recommended Progression
Start at Level 1 and work up. Each level builds on the previous one.

### Jump to a Level
```bash
git checkout level-1-observe/url-shortener      # Watch the system under load
git checkout level-2-experiment/url-shortener   # Tweak configs, see trade-offs
git checkout level-3-build/url-shortener        # Build the consistent hash ring
git checkout level-4-fix/url-shortener          # Diagnose and fix misconfigs
git checkout level-5-scratch/url-shortener      # Build the services from scratch
```

### Preserve Your Progress
```bash
git checkout level-3-build/url-shortener
git checkout -b my-progress/url-shortener-level-3
# ... work ... then come back anytime with:
git checkout my-progress/url-shortener-level-3
```

### Not Sure Which Level?
```bash
make diagnose
```

## Everyday Commands

```bash
make start          # build + start the stack
make load-test      # steady-state traffic (SCENARIO=read-spike|hot-key)
make dashboard      # open Grafana (http://localhost:3000)
make validate       # level-appropriate checks
make journal        # create/open your decision journal
make clean          # tear everything down
```

The lab config lives in `config.yaml` at the repo root — it controls the
cache provider, eviction policy, TTL, hash ring virtual nodes, DB pool size,
and gateway rate limits. `make redeploy` applies changes.
