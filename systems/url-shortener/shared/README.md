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
sdl level 1      # Watch the system under load
sdl level 2      # Tweak configs, see trade-offs
sdl level 3      # Build the consistent hash ring
sdl level 4      # Diagnose and fix misconfigs
sdl level 5      # Build the services from scratch
```

Your work survives: switching levels parks the current workspace and
restores it when you come back. `sdl status` always tells you where you are.

### Not Sure Which Level?
```bash
sdl diagnose
```

## Everyday Commands

```bash
sdl start          # build + start the stack
sdl load           # steady traffic (--scenario read-spike | hot-key)
sdl dashboard      # open Grafana (http://localhost:3000)
sdl validate       # level-appropriate checks
sdl journal        # create your decision journal
sdl clean          # tear everything down
```

The lab config lives at `workspace/config.yaml` — it controls the cache
provider, eviction policy, TTL, hash ring virtual nodes, DB pool size, and
gateway rate limits. `sdl restart` applies changes; `sdl reset` restores
the level's pristine state.
