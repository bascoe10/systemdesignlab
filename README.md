# SystemDesignLab

Hands-on system design learning. Clone it, run real distributed systems
locally, break them, fix them, and build them — with Grafana dashboards
showing you the consequences of every decision.

Not flashcards. Interview prep with actual depth.

**Who it's for:** the center of the target is mid-level engineers preparing
for senior interviews. Juniors are welcome — start at Level 1 and treat the
tooling (Docker, Grafana) as part of the curriculum. Seniors: `sdl diagnose`
will happily send you straight to Level 4.

## How It Works

Each system has five levels — five versions of the same system:

| Level | Mode | What you do |
|-------|------|-------------|
| 1 | Observe | Full system under load. Learn what healthy looks like |
| 2 | Experiment | Turn config knobs, watch the dashboards move |
| 3 | Build | One core component stubbed out — you build it |
| 4 | Fix | Misconfigured system — diagnose from dashboards, fix it |
| 5 | Scratch | Contracts and tests only — build everything |

One command materializes a level into `workspace/` — your disposable lab
bench. Switching levels parks your work and restores it when you return;
no git skills required beyond `clone`. Infrastructure — Prometheus, the
four Grafana dashboards, the chaos tooling — is identical at every level,
so the observability fluency you build at Level 1 is exactly what you
diagnose with at Level 4.

## Quick Start

Prerequisites: Docker (with Compose v2) and Go 1.23+ — or skip local setup
and open the repo in **GitHub Codespaces / any devcontainer**
(`.devcontainer/` ships everything, ports pre-forwarded).

```bash
git clone https://github.com/bascoe10/systemdesignlab
cd systemdesignlab

./sdl diagnose     # 5-question quiz → suggested starting level
./sdl start        # materialize Level 1 + start the stack
./sdl load         # steady traffic: 80% reads / 20% writes
./sdl dashboard    # Grafana → Golden Signals Overview
```

Then follow `workspace/CONTEXT.md` — every level tells you where you are
and what to do next. `./sdl status` reorients you anytime.

(Tip: `alias sdl=./sdl`. Windows: `cd cli && go build -o ../sdl.exe .`)

Two things worth knowing up front:

- **Level 1's `sdl validate` calibrates a baseline for your machine**
  (`.sdl/baseline.json`). Later levels judge latency against *your*
  healthy numbers, not someone else's laptop. Ratios (hit rate, error
  rate, key distribution) stay absolute — those don't depend on hardware.
- **Hop levels with `sdl level 3`** — your uncommitted work is parked and
  restored automatically when you come back. `sdl reset` re-materializes
  the current level fresh if you want a clean slate.

## Systems

**Phase 1 (available now)**
- **URL Shortener** — consistent hashing, write-through vs cache-aside
  caching, read-heavy optimization, cache eviction, hot keys

**Phase 2 (planned)** — Rate Limiter, Distributed Cache, plus BYOA AI
integration and Postgres/Cassandra + Kafka/RabbitMQ pluggability.
*Gate: Phase 2 starts when 25 people have reported completing Level 3 of
the URL Shortener (post in Discussions — it keeps the roadmap honest).*

**Phase 3 (planned)** — Feed System, Notification Service, and the capstone
Twitter Clone that integrates everything.

## What Makes It Stick

**Four Golden Signals observability.** Every system ships four Grafana
dashboards (Golden Signals, Component Deep Dive, Chaos Impact, SLI/SLO),
identical across all five levels — you build fluency with the tooling,
not just the code.

**Pluggable components.** `sdl cache memcached && sdl restart`, and watch
what changes under load. No code edits.

**Chaos built in.** `sdl chaos kill-cache`, `sdl chaos lag-db`,
`sdl chaos overload` — each with a resilience challenge doc and success
criteria in `docs/resilience-challenges/`. Breaking things on purpose is
the curriculum, not a stunt.

**Decision journals.** Levels 3–5 require `my-journal.md` (gitignored,
personal): constraints, trade-offs, load test numbers, and an AI
assistance log. Articulating decisions is what interviews actually test.

**AI-honest.** Use your own AI assistant if you like — but core components
are marked `AI_FREE_ZONE`, and `docs/ai-failure-cases/` is a dated,
reproducible protocol showing where plausible AI answers fail under load
(with numbers from this repo's own test suite).

## Command Reference

```bash
sdl start [--level N]        # materialize + start (continues where you left off)
sdl load [--scenario s] [--rate r] [--duration d]
sdl dashboard                # open Grafana
sdl validate                 # level-appropriate checks
sdl status                   # where am I, what's next
sdl level <1-5>              # hop levels (work parked & restored)
sdl reset                    # fresh copy of the current level
sdl journal                  # create your decision journal
sdl cache <redis|memcached>  # swap cache provider (then: sdl restart)
sdl restart | stop | clean   # apply config / pause / tear down
sdl chaos kill-cache | lag-db | overload | restore
sdl reveal-solution          # Level 4, when you're truly done
```

## Repository Layout

```
cli/              the sdl CLI — the product's spine
systems/          source of truth per system: shared/ code + level-N/ overlays
infrastructure/   Prometheus + Grafana (identical for every system and level)
docs/             design spec, ADRs, authoring guide, AI failure cases,
                  resilience challenges
workspace/        (gitignored) your materialized lab bench
.sdl/             (gitignored) state, parked levels, your baseline
```

Contributions welcome — see [CONTRIBUTING.md](CONTRIBUTING.md) and
[docs/SYSTEM_AUTHORING.md](docs/SYSTEM_AUTHORING.md).
