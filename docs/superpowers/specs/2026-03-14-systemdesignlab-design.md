# SystemDesignLab — Design Specification

## Overview

SystemDesignLab is an open-source monorepo for hands-on system design learning. Users clone the repo, run containerized microservices locally, and work through a 5-level progression that takes them from observing a working system to building one from scratch.

**Primary goal:** Interview prep with real depth — not flashcards, but building and breaking actual distributed systems.

**Secondary goal:** Broader learning platform for engineers who want to deeply understand distributed systems.

**Target users:** All experience levels. The 5-level progression accommodates juniors (Levels 1-2), mid-level engineers (Levels 3-4), and seniors (Level 5). A diagnostic quiz recommends an entry point.

## Architecture: Hybrid Monorepo + Scaffolding CLI

### Approach

Single repository with shared infrastructure and per-system directories. Levels are directory overlays — not git branches. A scaffolding CLI assembles the right files into a gitignored `workspace/` directory where the user works.

This avoids the branch-per-level maintenance nightmare (5 levels x 5 systems = 25+ branches) while keeping solutions hidden by default.

### Repository Structure

```
systemdesignlab/
├── infrastructure/
│   ├── docker/
│   │   ├── base/                    # Base Dockerfiles (Go service, Python service)
│   │   ├── compose-fragments/       # Reusable compose pieces
│   │   │   ├── prometheus.yml
│   │   │   ├── grafana.yml
│   │   │   ├── redis.yml
│   │   │   ├── memcached.yml
│   │   │   ├── postgres.yml
│   │   │   └── kafka.yml
│   │   └── networks.yml
│   ├── observability/
│   │   ├── prometheus/              # Prometheus configs, alert rules
│   │   ├── grafana/
│   │   │   ├── provisioning/        # Datasource + dashboard provisioning
│   │   │   └── dashboards/          # Per-system dashboard JSON
│   │   └── sli-slo/                 # SLI/SLO templates
│   ├── chaos/
│   │   ├── scripts/                 # Shell scripts for each chaos scenario
│   │   └── prompts/                 # "What to observe" prompts per scenario
│   └── k8s/                         # Phase 3 — K8s manifests
│
├── systems/
│   └── url-shortener/
│       ├── system.yaml              # Metadata: name, concepts, prerequisites
│       ├── services/
│       │   ├── api-gateway/
│       │   │   ├── Dockerfile
│       │   │   ├── main.go
│       │   │   └── ...
│       │   ├── shortener/
│       │   │   ├── Dockerfile
│       │   │   ├── main.go
│       │   │   ├── interfaces/
│       │   │   │   └── hasher.go    # ConsistentHasher interface
│       │   │   └── ...
│       │   └── redirector/
│       │       └── ...
│       ├── levels/
│       │   ├── level-1-observe/
│       │   │   ├── level.yaml
│       │   │   ├── context.md
│       │   │   ├── walkthrough.md
│       │   │   ├── config.yaml
│       │   │   └── docker-compose.override.yml
│       │   ├── level-2-experiment/
│       │   │   ├── level.yaml
│       │   │   ├── context.md
│       │   │   ├── experiments.md
│       │   │   ├── config.yaml
│       │   │   └── docker-compose.override.yml
│       │   ├── level-3-build/
│       │   │   ├── level.yaml
│       │   │   ├── context.md
│       │   │   ├── briefing.md
│       │   │   ├── stubs/
│       │   │   │   └── hasher.go
│       │   │   ├── tests/
│       │   │   ├── config.yaml
│       │   │   └── docker-compose.override.yml
│       │   ├── level-4-fix/
│       │   │   ├── level.yaml
│       │   │   ├── context.md
│       │   │   ├── symptoms.md
│       │   │   ├── bugs/
│       │   │   ├── health-checks/
│       │   │   └── docker-compose.override.yml
│       │   └── level-5-scratch/
│       │       ├── level.yaml
│       │       ├── context.md
│       │       ├── briefing.md
│       │       ├── contracts/
│       │       ├── stubs/
│       │       ├── tests/
│       │       └── docker-compose.skeleton.yml
│       ├── solutions/
│       │   ├── level-3/
│       │   ├── level-4/
│       │   └── level-5/
│       ├── load-tests/
│       │   ├── steady-state.js      # k6: 80/20 read/write, 1000 RPS
│       │   ├── read-spike.js        # k6: 95% reads, 5000 RPS
│       │   ├── hot-key.js           # k6: 50% reads hit 10 URLs
│       │   └── thresholds.yaml      # Pass/fail criteria per level
│       └── journal-template.md
│
├── cli/
│   ├── cmd/
│   │   ├── diagnose.go
│   │   ├── start.go
│   │   ├── switch.go
│   │   ├── validate.go
│   │   ├── reveal.go
│   │   ├── journal.go
│   │   └── chaos.go
│   ├── diagnose/
│   │   ├── questions.yaml
│   │   └── scoring.yaml
│   └── internal/
│       ├── scaffold/
│       ├── config/
│       └── compose/
│
├── workspace/                       # .gitignored — user's working area
├── config.yaml                      # Default component configuration
├── Makefile
├── .env.example
└── docs/
```

### Compose Assembly

The CLI assembles a `docker-compose.yml` from fragments:

```
Base fragments (infrastructure/docker/compose-fragments/)
  + System services (systems/url-shortener/services/*/Dockerfile)
  + Level override (systems/url-shortener/levels/level-N/docker-compose.override.yml)
  + Config selection (config.yaml: redis → pull redis.yml, not memcached.yml)
  = Final docker-compose.yml in workspace/
```

### Level Overlay Model

Levels don't duplicate service code. They provide overlays — files that replace or augment the base services:

- **Level 1-2:** No code overlay. Different configs and walkthrough docs only.
- **Level 3:** Overlay replaces one component with a stub (empty interface implementation).
- **Level 4:** Overlay injects bugs into the reference implementation (wrong config values, missing indexes, bad hash function).
- **Level 5:** Overlay replaces all services with bare skeletons (interface files + main.go entrypoints only).

All levels scaffold from a known baseline (the reference implementation), not from prior user work. Levels are independent in scaffolding but narratively connected through `context.md`.

### Pluggable Components

Components are swapped via `config.yaml` (source of truth) or CLI wrappers (convenience):

```yaml
cache:
  provider: redis          # or memcached
  config:
    host: cache
    port: 6379
    max_memory: 256mb
    eviction_policy: allkeys-lru

database:
  provider: postgres       # or cassandra (Phase 2)
  config:
    host: db
    port: 5432

queue:
  provider: kafka          # Phase 2
  config:
    brokers: ["kafka:9092"]
```

Services use an interface-based provider pattern. Redis and Memcached both implement `CacheProvider`. The service picks the implementation based on config at startup. Swapping providers requires no code changes — just config + restart.

When the user runs `make switch-cache memcached`, the CLI:
1. Updates `config.yaml`
2. Reassembles `docker-compose.yml` with `memcached.yml` instead of `redis.yml`
3. Restarts containers

### Infrastructure

- **Default:** Docker Compose. Runs on any machine with Docker. Lower barrier to entry.
- **Optional (Phase 3):** K8s manifests (k3s/minikube) for users who want more realistic deployments.

## Level Progression

### Pedagogical Design

Levels are a deliberate progression. Each level builds understanding on the previous one:

```
Level 1: See what healthy looks like        → builds intuition
Level 2: See what changes when you tweak    → builds understanding of trade-offs
Level 3: Build one piece in a working system → builds implementation skill
Level 4: Diagnose using Level 1 skills      → builds debugging skill
Level 5: Build with confidence from L1-L4   → proves mastery
```

### Level Details

**Level 1 — Observe & Understand**

Full working system running under load. The user is not coding — just watching Grafana dashboards, seeing how latency spikes and cache hit rates change. Guided questions prompt thinking: "Why does this break at 1000 requests per second?" Pure observation, building intuition first.

**Level 2 — Tweak & Experiment**

Same working system, but configs are exposed. The user swaps components — Redis for Memcached, changes eviction policies, adjusts shard counts. No code changes, just config-driven experimentation. They see what changes and why it matters under load.

**Level 3 — Build the Missing Piece**

One critical component is stubbed out — for example, the consistent hashing logic. The rest of the system is solid and running. The user implements just that one piece, staying focused on the concept without drowning in boilerplate.

**Level 4 — Fix the Broken System**

Full implementation exists but it's intentionally misconfigured. Cache miss rate at 80%, latency through the roof. The user diagnoses and fixes it using the observability skills they built in Level 1.

**Level 5 — Build from Scratch**

Only service interfaces and tests exist. The user wires everything together. By now they've seen the full system three different ways, so they're building with confidence — not guessing.

### Narrative Thread

Each level contains a `context.md` that explicitly connects to prior levels:

```markdown
## Where You Are — Level 4: Fix the Broken System

By now you've:
- [Level 1] Watched this system handle 1000 RPS. You know what
  healthy Golden Signals look like — p99 under 50ms, cache hit
  rate above 90%, error rate near zero.
- [Level 2] Swapped Redis for Memcached, changed eviction policies,
  adjusted shard counts. You know which knobs affect which signals.
- [Level 3] Built the consistent hashing ring. You understand how
  keys distribute across nodes.

Now: users are complaining that redirects are slow. Open the
dashboard. Something is very wrong. Find it and fix it.
```

### Skip Policy

Users can skip levels, but the system nudges toward the full progression:

- `make start-challenge url-shortener 4` works regardless of prior completion
- The CLI prints: "Level 4 builds on observability skills from Level 1. If you haven't done Level 1 yet, it takes ~15 minutes and gives you the dashboard baseline you'll need to diagnose this system."
- No hard gate — experienced users know what they're doing

## URL Shortener System (Phase 1)

### Services

- **api-gateway** (Go) — HTTP API, routing, rate limiting
- **shortener** (Go) — Core logic: hash generation, URL mapping
- **redirector** (Go) — Handles redirect lookups (read-heavy path)
- **cache** — Pluggable: Redis or Memcached
- **database** — PostgreSQL (Cassandra swap in Phase 2)

### Data Flow

```
Create short URL:
  Client → API Gateway → Shortener → DB (write)
                                   → Cache (write-through)

Redirect:
  Client → API Gateway → Redirector → Cache (hit?) → return
                                    → Cache (miss) → DB → Cache (backfill) → return
```

### Concepts Taught

- Consistent hashing
- Cache-aside vs write-through caching
- Read-heavy optimization
- Sharding strategies
- Cache eviction policies

### Load Test Scenarios

- **Steady state:** 80% reads, 20% writes, 1000 RPS
- **Read spike:** 95% reads, 5000 RPS (tests cache effectiveness)
- **Hot key:** 50% of reads hit 10 URLs (tests hot-spot handling)
- **Cache failure:** Kill cache mid-test (tests fallback behavior)

### Pluggable Components (Phase 1)

| Component | Options | What changes under load |
|-----------|---------|------------------------|
| Cache | Redis / Memcached | Eviction behavior, memory efficiency, feature set |
| Database | PostgreSQL (default) | Cassandra swap comes in Phase 2 |

## Observability

### Four Golden Signals Framework

Observability is a core learning objective, not just infrastructure. Each system teaches monitoring using the Four Golden Signals (Google SRE):

| Signal | What it measures | Example in URL Shortener |
|--------|-----------------|--------------------------|
| **Latency** | Duration of requests (success vs error) | Redirect p99, cache-miss vs cache-hit latency split |
| **Traffic** | Demand on the system | Requests/sec to API gateway, read/write ratio |
| **Errors** | Rate of failed requests | 5xx rate, redirects to expired URLs, cache timeouts |
| **Saturation** | How "full" the system is | Redis memory %, DB connection pool utilization, CPU |

### Grafana Dashboards (4 per system)

1. **Golden Signals Overview** — All four signals on one screen. Annotated panels explaining what each metric means.
2. **Component Deep Dive** — Per-component view (cache, DB, queue). Drill from signal to component.
3. **Chaos Impact** — Before/after split view for chaos experiments.
4. **SLI/SLO Tracker** — Service Level Indicators and Objectives. E.g., "99.5% of redirects complete in < 100ms."

### Observability Learning Per Level

| Level | Observability lesson |
|-------|---------------------|
| 1 | Read the Four Golden Signals dashboard. Understand what healthy looks like. Identify which signal moves during load tests. |
| 2 | Correlate config changes to signal changes. "You switched to Memcached — what happened to the saturation signal? Why?" |
| 3 | Instrument your implementation. Expose metrics from your code. Verify your component shows up on dashboards. |
| 4 | Diagnose the broken system using only dashboards. Which signal is unhealthy? Trace to root cause. |
| 5 | Define SLIs and SLOs for your system. Set up alerts. Justify thresholds in the decision journal. |

## Chaos Toolkit

Simple Make targets wrapping Docker commands. No external framework for Phase 1.

```bash
make chaos-kill-cache         # docker stop <cache-container>
make chaos-lag-network        # tc qdisc add netem delay 200ms on DB container
make chaos-corrupt-shard      # Insert bad data into one shard
make chaos-overload           # Run load test at 10x normal rate
make chaos-partition          # Network partition between services
make chaos-restore            # Undo all chaos — restore healthy state
```

Every chaos command includes an observation prompt: "Before running this, screenshot the Golden Signals dashboard. After, compare. Write in your decision journal: which signals moved, which didn't, and why."

## Decision Journals

Structured markdown template that forces deliberate thinking before, during, and after each challenge.

### Template

```markdown
## Decision Journal — [System Name] — Level [N]

### Before Building
**Problem constraints:**
- Latency requirement: ___ ms (p99)
- Scale requirement: ___ RPS
- Consistency requirement: strong / eventual / doesn't matter

**Component choices:**
- Why [your cache choice] over the alternative?
- What eviction policy and why?
- What hashing strategy and why?

### During Building
**Key decisions made:**
- Decision 1: ___
  - Alternatives considered: ___
  - Why this one: ___

**Where I got stuck:**
- Problem: ___
- How I resolved it: ___

### After Building
**Golden Signals results:**
- Latency (p99): ___
- Traffic handled: ___ RPS
- Error rate: ___%
- Saturation (peak): ___%

**Chaos test results:**
- Cache kill recovery time: ___
- Behavior under 10x load: ___

**What I'd do differently:**
- ___

### AI Assistance (if used)
- What I asked: ___
- What AI suggested: ___
- What I actually implemented: ___
- Why they differed: ___
```

- `make journal` opens the journal for the current challenge, pre-filled with the system's constraints
- Journals live in `workspace/` — the user's artifact, not checked into the repo
- The AI Assistance section is always present but optional in Phase 1

## Validation

### Per-Level Validation

| Level | What `make validate` does |
|-------|--------------------------|
| 1 | Checks user ran load tests (metrics exist in Prometheus). Prints reflection questions. Checks journal has observations. |
| 2 | Checks config changes were made (config change log). Prints comparison prompts. Checks journal has trade-off analysis. |
| 3 | Runs tiered tests (see below). Checks journal has component choices documented. |
| 4 | Verifies Golden Signals returned to healthy baselines. Runs health checks. Checks journal documents diagnostic path. |
| 5 | Full integration tests + load test thresholds + chaos survival. Checks journal has architecture decisions and SLO definitions. |

### Test Tiers (Level 3 & 5)

```
Required (must pass):
  - Interface contract tests
  - Basic functionality
  - "Does it work at all?"

Performance (must meet bar):
  - Load test: p99 < 100ms at 1000 RPS
  - Cache hit rate > 85% under steady state
  - "Does it work well enough?"

Stretch (optional, flagged):
  - Edge cases: all nodes in range fail, hash collision handling
  - Chaos survival: cache kill recovery < 5s
  - "Does it handle the hard stuff?"

Results printed as:
  [pass] Required    12/12 passed
  [pass] Performance 3/3 thresholds met
  [warn] Stretch     2/4 passed (optional)
```

## CLI Commands

```bash
make diagnose                          # Interactive quiz → recommends level
make start-challenge url-shortener 3   # Scaffold Level 3 into workspace/
make start                             # Spin up Docker Compose
make load-test                         # Run k6 against running system
make dashboard                         # Open Grafana in browser
make switch-cache memcached            # Swap cache provider, restart
make chaos-kill-cache                  # Kill cache container
make chaos-lag-network                 # Add 200ms latency
make chaos-restore                     # Restore healthy state
make validate                          # Run tests for current level
make reveal-solution                   # Show reference solution
make journal                           # Open decision journal
make clean                             # Tear down containers, clean workspace
```

### Diagnostic Quiz

10-15 multiple choice questions grouped by concept area:

- **Fundamentals** — caching, load balancing, database basics (gates Level 1-2)
- **Design trade-offs** — CAP theorem, consistency vs availability (gates Level 3)
- **Debugging intuition** — "p99 latency spiked after deploying X, what do you check?" (gates Level 4)
- **Architecture** — designing from requirements, capacity estimation (gates Level 5)

Output recommends a system + level. Biases toward starting at Level 1 even for experienced users, with a note that Level 1 takes ~15 minutes and provides the dashboard baseline needed for Level 4.

Questions and scoring live in `cli/diagnose/` as YAML. Deterministic scoring, no LLM needed.

## Scaffolding Flow

When the user runs `make start-challenge url-shortener 3`:

1. Read `systems/url-shortener/system.yaml` — validate system exists
2. Read `systems/url-shortener/levels/level-3/level.yaml` — get level config
3. Copy `services/` → `workspace/url-shortener/services/`
4. Apply level overlay — replace `services/shortener/hasher.go` with `levels/level-3/stubs/hasher.go`
5. Copy tests into workspace
6. Assemble `docker-compose.yml` from fragments + config.yaml + level override
7. Copy journal template → `workspace/journal.md`
8. Print `context.md` then `briefing.md` to terminal

## Phasing Strategy

### Phase 1 — Foundation + URL Shortener

Prove the full learning model end-to-end with one system.

- Shared infrastructure: Docker Compose, Prometheus, Grafana dashboards, Makefile
- CLI: `diagnose`, `start-challenge`, `switch-cache`, `validate`, `reveal-solution`, `journal`, `chaos-*`
- URL Shortener with all 5 levels (services, overlays, tests, walkthroughs, solutions)
- Load tests with k6
- Basic chaos commands: `kill-cache`, `lag-network`, `overload`, `restore`
- Decision journal templates
- Pluggable components: Redis/Memcached for cache

### Phase 2 — Expand Systems + AI Integration

- Rate Limiter (sliding window, token bucket)
- Distributed Cache (eviction policies, invalidation)
- BYOA AI integration (LLM API hookup, AI-free zone enforcement, auto-populated AI journal sections)
- Additional pluggable components: Postgres/Cassandra, Kafka/RabbitMQ
- More chaos scenarios

### Phase 3 — Full Platform

- Feed System
- Notification Service
- Capstone Twitter Clone (integrates all systems)
- Optional K8s manifests
- Advanced chaos toolkit
- Community contribution framework

## Competitive Landscape

No existing project combines all of SystemDesignLab's differentiators:

- **Theory repos** (System Design Primer, ByteByteGo) — popular but not hands-on
- **Reference architectures** (Google Online Boutique, eShopOnContainers) — runnable but not pedagogical
- **Algorithm challenges** (Fly.io Gossip Glomers) — hands-on but focused on distributed algorithms, not system architecture
- **Build-one-thing tools** (CodeCrafters) — progressive but single-component, not multi-service

SystemDesignLab's unique combination of 5-level progression + pluggable components + chaos engineering + decision journals + Four Golden Signals observability + containerized multi-service environments occupies an uncontested niche.
