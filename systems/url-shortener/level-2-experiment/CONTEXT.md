# Where You Are — Level 2: Tweak & Experiment

By now you've:
- **[Level 1]** Watched this system handle steady load. You know the healthy
  baseline: p99 well under 100ms, cache hit rate above 85%, errors near zero.

Now the config is yours. `config.yaml` at the repo root controls the cache
provider, eviction policy, TTL, memory limits, hash ring virtual nodes, DB
pool size, and gateway rate limits. Every knob maps to a visible change on
the dashboards — the loop is:

```bash
vim config.yaml
make redeploy && make load-test
make dashboard     # what moved? why?
```

Work through `EXPERIMENTS.md`. Each experiment states an expected outcome —
the learning happens when you compare it against what actually happened.

Next: `git checkout level-3-build/url-shortener`
