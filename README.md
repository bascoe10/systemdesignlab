# SystemDesignLab

Hands-on system design learning. Clone it, run real distributed systems
locally, break them, fix them, and build them — with Grafana dashboards
showing you the consequences of every decision.

Not flashcards. Interview prep with actual depth.

**Who it's for:** the center of the target is mid-level engineers preparing
for senior interviews. Juniors are welcome — start at Level 1 and treat the
tooling (Docker, git, Grafana) as part of the curriculum. Seniors: `make
diagnose` will happily send you straight to Level 4.

## How It Works

Each system ships as five git branches — five levels of the same system:

| Level | Branch | What you do |
|-------|--------|-------------|
| 1 | `level-1-observe/<system>` | Full system under load. Learn what healthy looks like |
| 2 | `level-2-experiment/<system>` | Turn config knobs, watch the dashboards move |
| 3 | `level-3-build/<system>` | One core component stubbed out — you build it |
| 4 | `level-4-fix/<system>` | Misconfigured system — diagnose from dashboards, fix it |
| 5 | `level-5-scratch/<system>` | Contracts and tests only — build everything |

Between levels, only the `system/` directory changes. Infrastructure —
Prometheus, the four Grafana dashboards, the chaos toolkit — is identical
everywhere, so the observability skills you build at Level 1 are exactly
the ones you diagnose with at Level 4.

## Quick Start

Prerequisites: Docker (with Compose v2), Go 1.23+, make — or skip local
setup entirely and open the repo in **GitHub Codespaces / any devcontainer**
(`.devcontainer/` ships everything, ports pre-forwarded).

```bash
git clone https://github.com/yourusername/systemdesignlab
cd systemdesignlab

make diagnose      # 5-question quiz → recommends your entry level
git checkout level-1-observe/url-shortener

make start         # build + start the stack
make load-test     # steady-state traffic (80% reads / 20% writes)
make dashboard     # Grafana → Golden Signals Overview
```

Then follow `system/CONTEXT.md` — every branch tells you where you are and
what to do next.

Two things worth knowing up front:

- **Level 1's `make validate` calibrates a baseline for your machine**
  (`.baseline.json`). Later levels judge latency against *your* healthy
  numbers, not someone else's laptop. Ratios (hit rate, error rate, key
  distribution) stay absolute — those don't depend on hardware.
- **Hop levels with `make level-3`** (or `sdl switch 3`) — it parks your
  uncommitted work and restores it when you come back, and `make start`
  keeps you off the CI-managed generated branches automatically.

Working from `main` (contributors) works too: `make start LEVEL=3`
assembles the level into `.lab/` with the same generator CI uses for
branches, and runs it there.

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
dashboards (Golden Signals, Component Deep Dive, Chaos Impact, SLI/SLO).
The dashboards are identical across all five levels — you build fluency
with the tooling, not just the code.

**Pluggable components.** Swap Redis for Memcached in `config.yaml`,
`make redeploy`, and watch what changes under load. No code edits.

**Chaos toolkit.** `make chaos-kill-cache`, `make chaos-lag-database`,
`make chaos-overload` — each with a resilience challenge doc and success
criteria. Breaking things on purpose is the curriculum, not a stunt.

**Decision journals.** Levels 3–5 require `my-journal.md` (gitignored,
personal): constraints, trade-offs, load test numbers, and an AI
assistance log. Articulating decisions is what interviews actually test.

**AI-honest.** Use your own AI assistant if you like — but core components
are marked `AI_FREE_ZONE`, and `docs/ai-failure-cases/` shows precisely
where plausible AI answers fail under load (with numbers from this repo's
own test suite).

## Everyday Commands

```bash
make diagnose                    # quiz → recommended level
make start / stop / clean        # lifecycle
make load-test                   # SCENARIO=read-spike|hot-key, RATE=, DURATION=
make dashboard                   # open Grafana
make redeploy                    # apply config.yaml changes
make switch-cache PROVIDER=memcached
make validate                    # level-appropriate checks
make journal                     # create your decision journal
make reveal-solution             # Level 4, when you're truly done
make chaos-kill-cache            # and friends; make chaos-restore undoes
```

## Repository Layout (main branch)

```
infrastructure/   Prometheus + Grafana (shared by every system and level)
systems/          Source of truth per system: shared/ code + level-N/ overlays
cli/              diagnose, validate, journal, reveal-solution
chaos-toolkit/    chaos scripts + resilience challenges
generator/        assembles level branches from main (CI runs it on push)
docs/             design spec, AI failure cases
```

Level branches are generated — never commit to them. See
[CONTRIBUTING.md](CONTRIBUTING.md).
