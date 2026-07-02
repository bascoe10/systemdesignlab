# Where You Are — Level 1: Observe & Understand

Welcome. This is a complete, healthy URL shortener. You write zero code at
this level — your job is to learn what "healthy" looks like on a dashboard,
because every later level asks you to recognize when it isn't.

```bash
make start        # build and start everything (~2 min first time)
make load-test    # steady traffic: 80% reads, 20% writes
make dashboard    # Grafana → Golden Signals Overview
```

Then open `QUESTIONS.md` and work through the observation prompts.

By the end of this level you should be able to answer, from memory:
- What is the steady-state p99? The cache hit rate? The error rate?
- Which Golden Signal moves first when traffic spikes?

That baseline is the diagnostic skill Level 4 is built on — and it's
literal: finish with `make validate`, which records your machine's healthy
numbers into `.baseline.json`. Levels 3–5 judge your implementations
against *your* baseline, so don't skip this.

Next: `make level-2` (or `git checkout level-2-experiment/url-shortener`)
