# Where You Are — Level 2: Tweak & Experiment

By now you've:
- **[Level 1]** Watched this system handle steady load. You know the healthy
  baseline: p99 well under 100ms, cache hit rate above 85%, errors near zero.

Now the config is yours. `workspace/config.yaml` controls the cache
provider, eviction policy, TTL, memory limits, hash ring virtual nodes, DB
pool size, and gateway rate limits. Every knob maps to a visible change on
the dashboards — the loop is:

```bash
vim workspace/config.yaml
sdl restart && sdl load
sdl dashboard     # what moved? why?
```

Work through `EXPERIMENTS.md`. Each experiment states an expected outcome —
the learning happens when you compare it against what actually happened.

Next: `sdl level 3`
