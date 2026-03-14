# SystemDesignLab — Design Specification

## Overview

SystemDesignLab is an open-source monorepo for hands-on system design learning. Users clone the repo, check out git branches per level, run containerized microservices locally, and work through a 5-level progression that takes them from observing a working system to building one from scratch.

**Primary goal:** Interview prep with real depth — not flashcards, but building and breaking actual distributed systems.

**Secondary goal:** Broader learning platform for engineers who want to deeply understand distributed systems.

**Target users:** All experience levels. The 5-level progression accommodates juniors (Levels 1-2), mid-level engineers (Levels 3-4), and seniors (Level 5). A diagnostic quiz objectively assesses knowledge and recommends an entry point.

## Architecture: Git Branch-Based Learning

### Branch Structure

Each system has one git branch per level. This is the core navigation mechanism — users move between levels using standard git commands, which is native to how engineers already work.

**Branch naming convention:**

```
level-1-observe/<system-name>
level-2-experiment/<system-name>
level-3-build/<system-name>
level-4-fix/<system-name>
level-5-scratch/<system-name>
```

**Examples:**

```
level-1-observe/url-shortener
level-2-experiment/url-shortener
level-3-build/url-shortener
level-4-fix/url-shortener
level-5-scratch/url-shortener
level-3-build/rate-limiter
level-3-build/capstone-twitter
level-5-scratch/capstone-twitter
```

**Rules:**

- Each branch is a fully self-contained, runnable state of that system. `git checkout` + `make start` is all a user needs.
- Users create their own working branch off the level branch to preserve progress:

```bash
git checkout level-3-build/url-shortener
git checkout -b my-progress/url-shortener-level-3
# ... work on the challenge ...
# Need a refresher on what healthy looks like?
git stash
git checkout level-1-observe/url-shortener
make start && make load-test && make dashboard
# Go back to your work
git checkout my-progress/url-shortener-level-3
git stash pop
```

- Users can pause, check out any level branch for a refresher, then return to their working branch without losing progress.

### Branch Content

Each branch contains the complete runnable state for that system at that level:

```
systemdesignlab/                       # On branch: level-3-build/url-shortener
├── infrastructure/
│   ├── docker/
│   │   ├── base/                      # Base Dockerfiles (Go service)
│   │   └── compose-fragments/         # Reusable compose pieces
│   │       ├── prometheus.yml
│   │       ├── grafana.yml
│   │       ├── redis.yml
│   │       ├── memcached.yml
│   │       ├── postgres.yml
│   │       └── networks.yml
│   ├── observability/
│   │   ├── prometheus/                # Prometheus configs, alert rules
│   │   └── grafana/
│   │       ├── provisioning/          # Datasource + dashboard provisioning
│   │       └── dashboards/            # Dashboard JSON (IDENTICAL across all levels)
│   └── k8s/                           # Phase 3 — K8s manifests
│
├── systems/
│   └── url-shortener/
│       ├── README.md                  # "Where do I start?" + git checkout instructions
│       ├── system.yaml                # System metadata
│       ├── services/
│       │   ├── api-gateway/
│       │   │   ├── Dockerfile
│       │   │   └── main.go
│       │   ├── shortener/
│       │   │   ├── Dockerfile
│       │   │   ├── main.go
│       │   │   └── hasher.go          # STUB on this branch (L3)
│       │   └── redirector/
│       │       ├── Dockerfile
│       │       └── main.go
│       ├── CONTEXT.md                 # "Where you are" — connects to prior levels
│       ├── BRIEFING.md                # Level 3: what to build, interfaces, hints
│       ├── EXPECTED_METRICS.md        # Reference numbers for load test success
│       ├── JOURNAL_TEMPLATE.md        # Decision journal template (committed)
│       ├── my-journal.md              # User's journal (.gitignored)
│       ├── tests/                     # Tests for the stubbed component
│       ├── load-tests/
│       │   ├── steady-state.js
│       │   ├── read-spike.js
│       │   └── hot-key.js
│       └── docker-compose.yml         # Complete compose file for this level
│
├── chaos-toolkit/
│   ├── scripts/                       # Chaos scripts
│   ├── RESILIENCE_CHALLENGE.md        # Per-chaos-command challenge docs
│   └── Makefile                       # Chaos make targets
│
├── cli/
│   ├── cmd/
│   │   ├── diagnose.go
│   │   └── validate.go
│   └── diagnose/
│       ├── questions.yaml             # Community-extensible question bank
│       └── scoring.yaml               # Transparent scoring logic
│
├── docs/
│   └── ai-failure-cases/             # AI failure case library
│
├── config.yaml                        # Component configuration
├── Makefile                           # Top-level make targets
└── .env.example
```

**What changes between branches for the same system:**

| Branch element | L1 Observe | L2 Experiment | L3 Build | L4 Fix | L5 Scratch |
|----------------|-----------|---------------|----------|--------|------------|
| Service code | Full implementation | Full implementation | One component stubbed | Full but misconfigured | Interface stubs only |
| CONTEXT.md | "Welcome" | References L1 | References L1-2 | References L1-3 | References L1-4 |
| Level-specific doc | QUESTIONS.md | EXPERIMENTS.md | BRIEFING.md | KNOWN_ISSUES.md | BRIEFING.md |
| config.yaml | Working defaults | Editable, documented | Working defaults | Broken values | Skeleton |
| tests/ | None | None | Component tests | Health checks | Full test suite |
| docker-compose.yml | Complete | Complete | Complete | Complete (broken config) | Skeleton |
| Grafana dashboards | **IDENTICAL** | **IDENTICAL** | **IDENTICAL** | **IDENTICAL** | **IDENTICAL** |
| load-tests/ | Present | Present | Present | Present | Present |
| EXPECTED_METRICS.md | Reference values | Reference values | Reference values | Reference values | Reference values |

**Critical rule:** Grafana dashboard configuration is IDENTICAL across all five levels for every system. Users must see the same panels every time so they build familiarity with the tooling, not just the code.

### Branch Maintenance

Infrastructure changes (shared Dockerfiles, Prometheus config, Grafana dashboards) must propagate to all branches. This is managed through:

- A `main` branch containing shared infrastructure and tooling
- CI automation that rebases level branches onto `main` when infrastructure changes
- Per-system branch sets that share the same infrastructure base

### Per-System README

Each system's README on every branch includes a "Where do I start?" section:

```markdown
## Where Do I Start?

### Recommended Progression
Start at Level 1 and work up. Each level builds on the previous one.

### Jump to a Level
git checkout level-1-observe/url-shortener     # Watch the system under load
git checkout level-2-experiment/url-shortener   # Tweak configs, see trade-offs
git checkout level-3-build/url-shortener        # Build the hashing ring
git checkout level-4-fix/url-shortener          # Diagnose and fix misconfigs
git checkout level-5-scratch/url-shortener      # Build from scratch

### Preserve Your Progress
git checkout level-3-build/url-shortener
git checkout -b my-progress/url-shortener-level-3
# ... work ... then come back anytime with:
git checkout my-progress/url-shortener-level-3

### Not Sure Which Level?
make diagnose
```

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

When the user runs `make switch-cache memcached`:
1. Updates `config.yaml`
2. Restarts containers with the new provider
3. User runs `make load-test` to see the impact

### Infrastructure

- **Default:** Docker Compose. Runs on any machine with Docker. Lower barrier to entry.
- **Optional (Phase 3):** K8s manifests (k3s/minikube) for users who want more realistic deployments.

### CLI Implementation

The CLI is written in Go and distributed as a pre-built binary for Linux, macOS, and Windows. Users without Go can download the binary from GitHub releases. Users with Go can also build from source with `go build ./cli/...`. Make targets wrap the CLI binary — users interact with `make` commands, not the CLI directly.

## Diagnostic Quiz

An objective assessment tool that replaces self-reported skill level. Invoked via `make diagnose`.

### Implementation

- Presents 5 short system design scenario questions
- Each question is multiple choice with 4 options
- Questions cover: caching trade-offs, database sharding, CAP theorem, load balancing strategies, consistency vs availability

### Question Bank

Questions are stored in `cli/diagnose/questions.yaml` so the community can add more over time:

```yaml
questions:
  - id: cache-eviction
    category: caching
    difficulty: intermediate
    question: |
      Your URL shortener serves 10M redirects/day. 90% of traffic goes to
      URLs created in the last 24 hours. Which cache eviction policy
      maximizes hit rate?
    options:
      a: "FIFO — oldest entries evicted first"
      b: "LRU — least recently used evicted first"
      c: "LFU — least frequently used evicted first"
      d: "Random — random entries evicted"
    correct: b
    explanation: |
      LRU naturally keeps hot URLs (recently created, frequently accessed)
      in cache while evicting stale long-tail URLs. LFU would keep old
      viral URLs too long. FIFO ignores access patterns entirely.
    maps_to_level: 2  # Tests Level 2 knowledge (config trade-offs)

  - id: cap-theorem
    category: consistency
    difficulty: advanced
    question: |
      Your distributed cache spans 3 data centers. A network partition
      isolates DC-3. Users in DC-3 are reading stale data. What do you do?
    options:
      a: "Return errors to DC-3 users until partition heals (CP)"
      b: "Serve stale data with a staleness indicator (AP)"
      c: "Synchronously replicate writes to all DCs before acknowledging"
      d: "Route all DC-3 traffic to DC-1 over the public internet"
    correct: b
    explanation: |
      For a URL shortener, availability matters more than consistency.
      A stale URL redirect is better than no redirect. Option C is
      impossible during a partition. Option D adds latency and a SPOF.
    maps_to_level: 3
```

### Scoring Logic

Scoring is transparent and documented in `cli/diagnose/scoring.yaml`:

```yaml
scoring:
  # Each correct answer adds points to the corresponding level bucket
  # Recommended level = highest level where user scored >= threshold
  level_thresholds:
    level-1: 0          # Default starting point
    level-2: 1          # Got at least 1 config/trade-off question right
    level-3: 2          # Got at least 2 design questions right
    level-4: 3          # Got at least 3 including debugging scenarios
    level-5: 4          # Got 4+ right, including architecture questions
```

### Output

```
Based on your answers, we recommend starting at Level 3.

  You understand caching trade-offs and CAP theorem well, but
  haven't worked through debugging distributed system failures.
  Level 3 will challenge you to build a consistent hashing ring.

  Run: git checkout level-3-build/url-shortener

  Tip: Even experienced engineers benefit from Level 1. It takes
  ~15 minutes and gives you the Grafana dashboard baseline you'll
  need in Level 4.
```

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

### Level 1 — Observe & Understand

Full working system deployed and running under simulated load. The user does ZERO coding at this level.

**What the user gets:**
- Complete running system with all services operational
- Grafana dashboards pre-configured and open
- Simulated load running via k6
- A `QUESTIONS.md` with guided observation prompts

**QUESTIONS.md examples:**

```markdown
## Observation Questions — URL Shortener

Run `make load-test` and open `make dashboard`. Answer these:

1. What is the steady-state P99 latency? What about P50?
   Why is the gap between them significant?

2. Watch the cache hit rate panel. Why does it start at 0%
   and climb to ~90% over the first 60 seconds?

3. Run `make load-test SCENARIO=read-spike`. Why does P99
   latency spike above 200ms at 800 RPS?

4. What happens to cache hit rate when you scale to 3
   replicas? (Hint: look at the consistent hashing panel)

5. Which of the Four Golden Signals degrades first under
   sustained high load? Why that one?
```

**Goal:** Build intuition through observation before touching code. By the end of Level 1, the user knows what "healthy" looks like on the dashboard.

### Level 2 — Tweak & Experiment

Same running system, but config/YAML is now exposed and editable. The user changes cache provider, eviction policy, replica count, and shard count via config only — no Go code changes.

**Workflow:**
1. Edit `config.yaml`
2. `make redeploy && make load-test`
3. Observe impact on Grafana dashboards

**EXPERIMENTS.md — suggested experiments with expected outcomes:**

```markdown
## Experiments — URL Shortener

### Experiment 1: Switch Cache Provider
Change `cache.provider` from `redis` to `memcached` in config.yaml.
Run: make redeploy && make load-test

Expected outcome:
- P99 latency increases slightly (Memcached lacks Redis's data structures)
- Memory usage pattern changes (Memcached uses slab allocation)
- Cache hit rate should remain similar

What to look for:
- Compare the Saturation signal before and after
- Note the difference in eviction behavior under memory pressure

### Experiment 2: Change Eviction Policy
Set `cache.config.eviction_policy` to `allkeys-random`.
Run: make redeploy && make load-test

Expected outcome:
- Cache hit rate drops 10-15% (random eviction ignores access patterns)
- P99 latency increases proportionally

### Experiment 3: Reduce Cache Memory
Set `cache.config.max_memory` from `256mb` to `32mb`.
Run: make redeploy && make load-test SCENARIO=hot-key

Expected outcome:
- Eviction rate spikes (visible in Component Deep Dive dashboard)
- Hot keys still cached, but long-tail URLs evicted aggressively
```

### Level 3 — Build the Missing Piece

One critical component has its core logic replaced with a stub. The rest of the system is fully operational.

**The stub includes:**
- Function signature and interface contract
- TODO comments with hints pointing to relevant concepts
- Inline documentation explaining what the component should do

**Example stub (`hasher.go`):**

```go
// ConsistentHasher distributes keys across nodes using consistent hashing.
// TODO: Implement the hash ring. Consider:
//   - How do you handle node addition/removal without rehashing all keys?
//   - What is the role of virtual nodes in balancing load?
//   - Hint: research "consistent hashing" and "virtual nodes"
//
// AI_FREE_ZONE: Complete this section without AI assistance.
// Consistent hashing is foundational. Build it yourself first.

type ConsistentHasher interface {
    AddNode(nodeID string) error
    RemoveNode(nodeID string) error
    GetNode(key string) (string, error)
}

type hasher struct {
    // TODO: implement
}

func NewConsistentHasher(replicas int) ConsistentHasher {
    // TODO: implement
    panic("not implemented")
}
```

**Tests exist for the stubbed component** — the user knows they're done when:
1. All tests pass (`make validate`)
2. Load test metrics match the reference numbers in `EXPECTED_METRICS.md`

**EXPECTED_METRICS.md:**

```markdown
## Expected Metrics — URL Shortener Level 3

When your implementation is correct, you should see:

| Metric | Target | Tolerance |
|--------|--------|-----------|
| Cache hit rate (steady state) | > 85% | ±5% |
| P99 latency (1000 RPS) | < 100ms | — |
| P50 latency (1000 RPS) | < 15ms | — |
| Throughput (sustained) | > 1000 RPS | — |
| Key distribution std dev | < 15% | across nodes |
```

### Level 4 — Fix the Broken System

Full implementation exists but with 3-5 deliberate misconfigurations. The user must diagnose root causes using observability tools, then fix them.

**KNOWN_ISSUES.md — symptoms only, no causes:**

```markdown
## Known Issues — URL Shortener Level 4

The system is deployed and running but something is very wrong.
Use the Grafana dashboards to diagnose and fix these issues.

### Symptom 1: Terrible Cache Performance
Cache hit rate is 12% under normal load — expected is 85%+.
The system is hitting the database for almost every request.

### Symptom 2: High Latency
P99 latency is 800ms under normal load — expected is under 50ms.
Users are complaining about slow redirects.

### Symptom 3: Uneven Load Distribution
One shortener replica is handling 70% of all requests.
The other two replicas are nearly idle.

### Symptom 4: Memory Pressure
Redis is using 98% of allocated memory despite low traffic.
Eviction rate is extremely high.

DO NOT look at SOLUTIONS.md until you've diagnosed and fixed
all issues. The challenge is in the diagnosis.
```

**SOLUTIONS.md** — exists but is hidden behind a gitignored file that users must explicitly unlock:

```markdown
<!-- SOLUTIONS.md is present in the repo but listed in .gitignore -->
<!-- To reveal: make reveal-solution -->
<!-- This copies SOLUTIONS.md from a hidden location into the working directory -->

## Solutions — URL Shortener Level 4

### Root Cause 1: Wrong Eviction Policy
config.yaml has `eviction_policy: noeviction`. Redis rejects new
writes when full instead of evicting old keys. Fix: set to `allkeys-lru`.

### Root Cause 2: Missing Database Index
...
```

### Level 5 — Build from Scratch

Only interface definitions, test files, and docker-compose skeleton exist. The user implements all business logic.

**What the user gets:**
- OpenAPI contracts defining the API surface of each service
- Interface definitions (Go interfaces for each service)
- The same tests from Level 3 (component tests) plus full integration tests
- The same load test targets from Level 3 define success criteria (`EXPECTED_METRICS.md`)
- A docker-compose skeleton with service entries but no built images
- The same Grafana dashboards — they light up as services come online

**BRIEFING.md for Level 5:**

```markdown
## Level 5 — Build from Scratch

You've seen this system four ways:
- [Level 1] Running healthy — you know what the dashboard should look like
- [Level 2] Under different configs — you know which knobs matter
- [Level 3] With one piece missing — you built the hardest component
- [Level 4] Broken — you diagnosed failures from dashboard symptoms

Now build it all. You have:
- API contracts in contracts/ (OpenAPI specs)
- Interface definitions in each service directory
- The test suite from Level 3
- The same Grafana dashboards — they'll light up as you wire things in
- EXPECTED_METRICS.md with your success criteria

Start with `docker-compose.yml`. Get one service running. Then the next.
```

### Narrative Thread

Each level contains a `CONTEXT.md` that explicitly connects to prior levels:

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

- `git checkout level-4-fix/url-shortener` works regardless of prior completion
- The system README recommends starting at Level 1
- `make diagnose` biases toward lower levels with a note that Level 1 takes ~15 minutes
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

## Capstone System — Twitter Clone

### Why a Capstone

Real system design interviews do not test systems in isolation. Interviewers ask: "Design Twitter." That requires reasoning about how components interact, where failures cascade, and what trade-offs emerge from integration. The capstone simulates this.

### Location

`systems/capstone-twitter/` — present on branches `level-3-build/capstone-twitter` and `level-5-scratch/capstone-twitter` only.

### What Users Must Integrate

| Component | Role in Twitter Clone | Source System |
|-----------|----------------------|---------------|
| URL Shortener | Shortened links in tweets (t.co-style) | systems/url-shortener |
| Rate Limiter | Per-user posting limits, API throttling | systems/rate-limiter |
| Feed System | Fan-out on write vs read, timeline generation | systems/feed-system |
| Notification Service | Async delivery of mentions, likes, follows via message queue | systems/notification-service |

### Capstone Levels

The capstone only has Level 3 and Level 5 — by the time users reach the capstone, they have already observed and experimented with individual systems.

**Level 3 — Build the Missing Piece:** Individual systems are running. The user must implement the integration layer: API orchestration, cross-service communication, failure handling between services.

**Level 5 — Build from Scratch:** Interface definitions and API contracts only. The user builds the entire integrated system.

### DESIGN_DECISIONS.md

Integration questions the user must answer before coding:

```markdown
## Design Decisions — Capstone Twitter Clone

Answer these BEFORE writing code. Put your answers in your decision journal.

### Service Communication
1. Should the URL shortener be called synchronously during tweet creation
   or queued asynchronously? What are the trade-offs?
   - Sync: tweet creation blocked if shortener is slow/down
   - Async: tweet appears without shortened URL initially, eventual consistency

2. If the rate limiter service is down, should tweet creation fail or
   degrade gracefully? How do you implement that?
   - Fail: consistent enforcement, but user-facing errors
   - Degrade: allow tweets through, risk abuse, implement circuit breaker

### Data Flow
3. Fan-out on write or fan-out on read for the timeline?
   - Write: pre-compute timelines, fast reads, expensive for popular users
   - Read: compute on demand, slow for users following many accounts

4. How do you handle a user with 10M followers posting a tweet?
   - Full fan-out: 10M timeline writes per tweet
   - Hybrid: fan-out for normal users, pull-based for celebrities

### Failure Modes
5. The notification service queue is backed up by 100K messages.
   What degrades? What stays healthy? How do you recover?

6. One URL shortener shard is down. What happens to tweets containing
   URLs that hash to that shard?
```

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

**Critical:** Dashboard JSON is identical across all 5 level branches for each system. Dashboards are stored in `infrastructure/observability/grafana/dashboards/` and shared via the branch infrastructure base.

### Observability Learning Per Level

| Level | Observability lesson |
|-------|---------------------|
| 1 | Read the Four Golden Signals dashboard. Understand what healthy looks like. Identify which signal moves during load tests. |
| 2 | Correlate config changes to signal changes. "You switched to Memcached — what happened to the saturation signal? Why?" |
| 3 | Instrument your implementation. Expose metrics from your code. Verify your component shows up on dashboards. |
| 4 | Diagnose the broken system using only dashboards. Which signal is unhealthy? Trace to root cause. |
| 5 | Define SLIs and SLOs for your system. Set up alerts. Justify thresholds in the decision journal. |

## Chaos Toolkit

### Location and Structure

```
chaos-toolkit/
├── scripts/
│   ├── kill-cache.sh
│   ├── lag-database.sh
│   ├── kill-shard.sh
│   ├── overload.sh
│   └── corrupt-queue.sh
├── resilience-challenges/
│   ├── KILL_CACHE.md
│   ├── LAG_DATABASE.md
│   ├── KILL_SHARD.md
│   ├── OVERLOAD.md
│   └── CORRUPT_QUEUE.md
├── restore.sh
└── Makefile
```

### Chaos Commands

Each command is both a Make target and a standalone bash script:

**`make chaos-kill-cache`**
- Kills Redis/Memcached container for 60 seconds, then restores it
- User must observe fallback behavior in Grafana
- Tests: Does the system degrade gracefully? How long until recovery?

**`make chaos-lag-database`**
- Injects 300ms network latency on database container using `tc netem`
- User must observe impact on P99 and consider circuit breaker patterns
- Tests: Which Golden Signal degrades first? How does the cache absorb the impact?

**`make chaos-kill-shard`**
- Takes one database shard offline
- User must observe data unavailability and implement fallback reads
- Tests: What percentage of requests fail? How does the hash ring handle it?

**`make chaos-overload`**
- Runs 10x normal traffic via k6 for 2 minutes
- User must observe which component degrades first
- Tests: What is the system's breaking point? Which resource saturates?

**`make chaos-corrupt-queue`**
- Drops 20% of Kafka/RabbitMQ messages randomly
- User must implement dead letter queue and retry logic
- Tests: Are notifications eventually delivered? What is the message loss rate?

**`make chaos-restore`**
- Undoes all chaos — restores healthy state

### Cross-Platform Note

`chaos-lag-database` uses `tc qdisc add netem` which requires the Linux kernel `netem` module. On Docker Desktop for Windows/Mac, this may not work depending on networking mode. The implementation should detect the platform and either use `tc` (Linux), Toxiproxy sidecar (Windows/Mac fallback), or print a warning with manual instructions.

### Resilience Challenge Documents

Each chaos script has a corresponding `RESILIENCE_CHALLENGE.md`:

```markdown
## Resilience Challenge: Cache Failure

### What Was Injected
Redis container killed for 60 seconds, then restored.

### What to Look For in Grafana
- Golden Signals Overview: Latency spike, Error rate spike, Traffic unchanged
- Component Deep Dive: Cache hit rate drops to 0%, DB query rate spikes
- Watch recovery: How quickly does cache hit rate return to baseline?

### The Fix
<details>
<summary>Click to reveal (try to figure it out first)</summary>

Implement a cache-aside pattern with graceful fallback:
- On cache miss: read from DB, backfill cache
- On cache unavailable: bypass cache entirely, serve from DB
- Add circuit breaker: stop attempting cache after N failures,
  periodically retry to detect recovery

Success criteria:
- Error rate stays below 1% during cache outage
- P99 latency degrades but stays under 500ms (DB-only path)
- Cache hit rate recovers to >85% within 30 seconds of restore
</details>

### Success Criteria
| Metric | During Outage | After Recovery |
|--------|--------------|----------------|
| Error rate | < 1% | < 0.1% |
| P99 latency | < 500ms | < 100ms |
| Cache hit rate | 0% (expected) | > 85% within 30s |
```

## AI Integration

### AI-Free Zones

Certain challenge files are marked with a header comment:

```go
// AI_FREE_ZONE: Complete this section without AI assistance.
// This concept is foundational. Build it yourself first.
```

AI-free zones apply to concepts where building from scratch is essential for understanding:
- Consistent hashing implementation (Level 3, URL Shortener)
- CAP theorem decision points (capstone design decisions)
- Sharding key selection logic
- Circuit breaker implementation

The zones are advisory, not enforced. They teach users to recognize when AI is a crutch vs. when it's a tool.

### AI Failure Case Library

Location: `docs/ai-failure-cases/`

One markdown file per system showing where AI gives plausible but flawed answers:

```markdown
## AI Failure Case: Consistent Hashing — URL Shortener

### The Naive Prompt
"Implement consistent hashing in Go for a URL shortener"

### The AI's Response (typical)
[AI generates a basic hash ring using MD5, fixed number of virtual nodes,
no handling for node failures during rebalancing]

### Why It Fails Under Load
Load test results with AI implementation:
- Key distribution std dev: 45% (target: <15%)
- 30% of keys rehashed when adding one node (target: <10%)
- No graceful handling of node removal during active requests

### The Correct Implementation
[Reference implementation with proper virtual node count tuning,
bounded-load consistent hashing, and safe node removal]

### The Lesson
AI implementations often work for small-scale demos but fail at
production scale. The AI's version has no concept of bounded loads,
ignores the rebalancing problem during node changes, and uses a
hash function with poor distribution properties for this use case.
```

### Prompt Engineering as a Skill

The decision journal includes an AI assistance log that teaches users to critically evaluate AI output:

See Decision Journal section below for the full template.

## Decision Journals

### Required Artifact

The decision journal is **required for Level 3, 4, and 5**. It is the primary artifact users take away from the platform — it trains them to articulate decisions, which is what interviewers are actually evaluating.

### Location

- `JOURNAL_TEMPLATE.md` — committed to the repo on every branch
- `my-journal.md` — user's filled-in journal, gitignored so it stays personal

### Template

```markdown
## System: [name] | Level: [number] | Date: [date]

### Constraints I Identified
- Latency requirement:
- Scale requirement:
- Consistency vs availability trade-off:
- Read/write ratio:

### Component Decisions
| Decision | Option A | Option B | I Chose | Because |
|----------|----------|----------|---------|---------|

### What Surprised Me

### What I Would Do Differently at 10x Scale

### AI Assistance Log
#### Attempt 1
- Prompt: [what I asked]
- AI Response Summary: [what it suggested]
- Why I accepted/rejected it: [my reasoning]
- What I changed: [my modification]

#### Load Test Result After AI-Assisted Implementation
| Metric | Target | Actual | Pass/Fail |
|--------|--------|--------|-----------|

### Load Test Results
| Metric | Target | Actual | Pass/Fail |
|--------|--------|--------|-----------|
| Cache hit rate (steady state) | > 85% | | |
| P99 latency (1000 RPS) | < 100ms | | |
| P50 latency (1000 RPS) | < 15ms | | |
| Throughput (sustained) | > 1000 RPS | | |
```

### Validation

`make validate` checks that `my-journal.md` exists and has non-empty required sections for Levels 3-5. This is a structural check (non-empty sections), not a semantic evaluation. Semantic quality assessment is deferred to Phase 2's AI integration.

## Validation

### Per-Level Validation

| Level | What `make validate` does |
|-------|--------------------------|
| 1 | Checks user ran load tests (metrics exist in Prometheus). Prints QUESTIONS.md answers prompt. |
| 2 | Checks config was changed (diff against original config.yaml). Prints experiment comparison prompts. |
| 3 | Runs tiered tests (see below). Checks `my-journal.md` has non-empty required sections. |
| 4 | Verifies Golden Signals returned to healthy baselines. Runs health checks against KNOWN_ISSUES.md symptoms. Checks journal. |
| 5 | Full integration tests + load test thresholds + chaos survival. Checks journal has architecture decisions and SLO definitions. |

### Test Tiers (Level 3 & 5)

```
Required (must pass):
  - Interface contract tests
  - Basic functionality
  - "Does it work at all?"

Performance (must meet bar):
  - Load test metrics match EXPECTED_METRICS.md targets
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
make diagnose                 # Diagnostic quiz → recommends level + branch
make start                    # Spin up Docker Compose for current branch
make load-test                # Run k6 against running system
make load-test SCENARIO=x     # Run specific scenario (read-spike, hot-key)
make redeploy                 # Rebuild and restart after config changes
make dashboard                # Open Grafana in browser
make switch-cache memcached   # Swap cache provider in config
make chaos-kill-cache         # Kill cache for 60s
make chaos-lag-database       # Add 300ms DB latency
make chaos-kill-shard         # Take one shard offline
make chaos-overload           # 10x traffic for 2 minutes
make chaos-corrupt-queue      # Drop 20% of queue messages
make chaos-restore            # Undo all chaos
make validate                 # Run level-appropriate tests
make reveal-solution          # Unhide SOLUTIONS.md (Level 4)
make journal                  # Open decision journal
make clean                    # Tear down containers
```

## Phasing Strategy

### Phase 1 — Foundation + URL Shortener

Prove the full learning model end-to-end with one system.

- Shared infrastructure: Docker Compose, Prometheus, Grafana dashboards, Makefile
- Git branch structure for URL Shortener (5 level branches)
- CLI: `diagnose`, `validate`, `reveal-solution`, `journal`
- URL Shortener with all 5 levels:
  - Full service implementations + stubs + broken configs + skeletons
  - QUESTIONS.md, EXPERIMENTS.md, BRIEFING.md, KNOWN_ISSUES.md, EXPECTED_METRICS.md
  - CONTEXT.md narrative thread across levels
  - JOURNAL_TEMPLATE.md
- Load tests with k6 (4 scenarios)
- Chaos toolkit: `kill-cache`, `lag-database`, `overload`, `restore`
- Resilience challenge documents
- Pluggable components: Redis/Memcached for cache
- Decision journal templates
- Diagnostic quiz with YAML question bank

### Phase 2 — Expand Systems + AI Integration

- Rate Limiter system (sliding window, token bucket) — 5 level branches
- Distributed Cache system (eviction policies, invalidation) — 5 level branches
- BYOA AI integration:
  - AI-free zone enforcement tooling
  - AI failure case library (`docs/ai-failure-cases/`)
  - Auto-populated AI journal sections from BYOA conversations
- Additional pluggable components: Postgres/Cassandra, Kafka/RabbitMQ
- Additional chaos: `kill-shard`, `corrupt-queue`
- Expanded diagnostic quiz question bank

### Phase 3 — Full Platform

- Feed System — 5 level branches
- Notification Service — 5 level branches
- Capstone Twitter Clone — Level 3 and Level 5 branches only
  - DESIGN_DECISIONS.md with integration questions
  - Cross-service failure cascade testing
- Optional K8s manifests
- Advanced chaos toolkit
- Community contribution framework

## Competitive Landscape

No existing project combines all of SystemDesignLab's differentiators:

- **Theory repos** (System Design Primer, ByteByteGo) — popular but not hands-on
- **Reference architectures** (Google Online Boutique, eShopOnContainers) — runnable but not pedagogical
- **Algorithm challenges** (Fly.io Gossip Glomers) — hands-on but focused on distributed algorithms, not system architecture
- **Build-one-thing tools** (CodeCrafters) — progressive but single-component, not multi-service

SystemDesignLab's unique combination of git-branch-based levels + pluggable components + chaos engineering + required decision journals + AI failure cases + Four Golden Signals observability + capstone integration + containerized multi-service environments occupies an uncontested niche.
