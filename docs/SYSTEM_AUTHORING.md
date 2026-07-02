# Authoring a New System

This guide codifies what made the URL Shortener work, so system #2 costs a
fraction of system #1. Follow the invariants; deviate only with a reason
you can defend in review.

## Go / No-Go Gate

Do not start a new system until the previous one has demonstrated
completions (see the public gate in the README roadmap). Content is the
expensive asset here — spend it where the funnel proves people finish.

## Directory Contract

Copy the shape of `systems/url-shortener/`:

```
systems/<name>/
  shared/
    services/            one Go module; per-service mains + internal/
    load-tests/          k6 scenarios (RATE/DURATION overridable via env)
    db/                  schema init
    docker-compose.yml   paths relative to system/ on generated branches
    system.yaml.tmpl     __LEVEL__ / __LEVEL_NAME__ / __ESTIMATED_TIME__ /
                         __GENERATED_FROM__ placeholders
    README.md            "Where do I start?" section
  JOURNAL_TEMPLATE.md
  level-1-observe/       CONTEXT.md QUESTIONS.md config.yaml
  level-2-experiment/    CONTEXT.md EXPERIMENTS.md config.yaml
  level-3-build/         CONTEXT.md BRIEFING.md EXPECTED_METRICS.md
                         config.yaml stubs/<component>.go
  level-4-fix/           CONTEXT.md KNOWN_ISSUES.md config.yaml
                         .solutions/SOLUTIONS.md
  level-5-scratch/       CONTEXT.md BRIEFING.md EXPECTED_METRICS.md
                         config.yaml contracts/ stubs/<svc>/main.go
```

Then register the system in `generator/generate.sh` (the level-specific
surgery in `assemble()`) and add its scrape targets/panels if it introduces
new components.

## The Invariants

1. **Metric names are a public API.** Dashboards, `cli/validate.go`, and
   the docs all bind to them. New metrics go in the system's `internal/obs`
   equivalent with the same naming discipline (`route` labels, not raw
   paths; ratios computable from counters).

2. **The Level 3 stub must degrade, not crash.** The system boots and
   serves with the stub in place; the dashboard shows *why the missing
   component matters* (for the ring: 100% cache bypass). If the stub can
   only panic, you picked the wrong component to stub.

3. **Correctness is quantitative.** The Level 3 test suite has three tiers:
   contract (does it behave), performance (measurable quality bars — the
   ring's <15% distribution stddev is the exemplar), and the live metrics
   in EXPECTED_METRICS.md. "Tests pass" alone is not done.

4. **Level 4 misconfigs are config-only and dashboard-diagnosable.** Every
   root cause: (a) lives in `config.yaml`, (b) produces a symptom visible
   on the shipped panels, (c) is distinguishable from the other causes by
   evidence, not guessing. 3–5 causes; let at least two mask each other —
   layered failure is the real-world lesson. Write SOLUTIONS.md with the
   evidence chain, not just the fix.

5. **Machine-dependent numbers are baseline-relative.** Absolute targets
   only for ratios (hit rates, error rates, distribution skew). Latency
   and throughput validate against the user's Level 1 baseline
   (`.baseline.json`); reference numbers in docs are labeled "typical
   hardware".

6. **Dashboards are identical across all five levels** and generated from
   `gen_dashboards.py` — never hand-edited JSON. New panels need a
   teaching-oriented `description`: what healthy looks like, what a
   deviation means.

7. **The narrative thread is mandatory.** Each CONTEXT.md names what the
   user did at every prior level and connects it to the current challenge.
   Docs ask questions before giving answers; hints escalate behind
   `<details>` folds; expected outcomes are stated *before* the user runs
   the experiment so prediction-vs-reality does the teaching.

8. **AI-free zones are for concepts where building teaches** (the L3
   component, by definition). Pair each with a dated, reproducible entry in
   `docs/ai-failure-cases/` — prompt, test command, recorded numbers — so
   users can re-run it against current models instead of trusting a claim.

## Definition of Done for a System PR

- [ ] All five levels assemble: `bash generator/generate.sh --out /tmp/x <name> N` for N=1..5
- [ ] Stub levels compile; L3 tests fail on the stub and pass on the reference
- [ ] `make start LEVEL=N && make load-test && make validate` green on 1, and on 3–5 with the reference implementation
- [ ] Every L4 symptom diagnosed by a fresh reviewer using only dashboards
- [ ] Docs carry the narrative thread and expected outcomes
- [ ] CI green (compose config both cache profiles, shellcheck, dashboards in sync)
