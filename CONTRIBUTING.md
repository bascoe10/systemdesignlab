# Contributing to SystemDesignLab

## The Model

All content lives on `main`. There is no branch generation and no build
step for content: the `sdl` CLI **materializes** a level into `workspace/`
at runtime from `systems/<name>/shared/` plus the level overlay
(see [ADR 0002](docs/adr/0002-cli-materialized-workspaces.md)).

```
systems/<name>/
  shared/            code + compose + load tests shared by all levels
    services/        one Go module; per-service main packages + internal/
    docker-compose.yml   (paths relative to the workspace root)
    system.yaml.tmpl
  level-1-observe/   per-level overlay: docs + config.yaml
  level-2-experiment/
  level-3-build/     + stubs/ (replaces the real component in the workspace)
  level-4-fix/       + .solutions/ (hidden until `sdl reveal-solution`)
  level-5-scratch/   + stubs/*.go.tmpl (skeleton mains) + contracts/ (OpenAPI)
```

Level 5 stub mains ship as `.go.tmpl` so the repo's root Go module doesn't
try to build them; the materializer renders them to `main.go`.

## Development Loop

```bash
go test ./...                                  # CLI
(cd systems/url-shortener/shared/services && go test ./internal/...)

./sdl start --level 3 && ./sdl load            # end-to-end, exactly as users run it
./sdl materialize url-shortener 5 /tmp/lab     # inspect any level's assembly
(cd /tmp/lab/services && go build ./...)       # stub levels must compile
```

CI enforces: services tests (with `-race`), CLI build, compose config
validation for every level × both cache profiles, stub-level compilation,
"L3 ring tests fail on the stub" (the challenge must stay a challenge),
L4 solutions present, and `dashboards/*.json` in sync with
`gen_dashboards.py`.

## Common Contributions

**Quiz questions** — add to `cli/diagnose/questions.yaml` with a
`maps_to_level` tag, and update the bucket counts in `scoring.yaml`.

**Grafana dashboards** — edit
`infrastructure/observability/grafana/gen_dashboards.py` and re-run it;
never edit the JSON by hand. Dashboards must stay identical across levels,
and any new panel needs the backing metric emitted from
`internal/obs/metrics.go` (metric names are a public API — the dashboards
and `cli/validate.go` depend on them).

**Chaos** — chaos lives in `cli/chaos.go` (Go, cross-platform — no bash).
Pair every new injection with a `docs/resilience-challenges/<X>.md`
containing observation prompts and success criteria.

**New systems (Phase 2+)** — read
[docs/SYSTEM_AUTHORING.md](docs/SYSTEM_AUTHORING.md) first; it codifies
the invariants (quantitative L3 tests, config-only dashboard-diagnosable
L4 misconfigs, baseline-relative thresholds, the narrative thread) and the
definition of done for a system PR. Register the new system's level
surgery in `cli/workspace.go` (`materialize`). Note the go/no-go gate in
the README roadmap before investing.

## Style Notes

- Teaching docs state expected outcomes and ask questions; they don't just
  give answers. Hints escalate behind `<details>` folds.
- Level 4 root causes must be fixable in `config.yaml` — no code edits.
- User-facing commands are `sdl` verbs; if a doc tells users to run raw
  `docker` or `git`, that's usually a missing CLI affordance.
