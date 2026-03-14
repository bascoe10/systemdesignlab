# SystemDesignLab

An open source monorepo for hands-on system design learning.
Clone it, build real distributed systems, and prep for system
design interviews with actual depth.

---

## What is This?

SystemDesignLab bridges the gap between *knowing* system design
concepts and *building* them. You get containerized microservices,
real traffic simulation, observability dashboards, and chaos
engineering — all locally on your machine.

---

## Learning Progression

Each system has five levels:

| Level | Mode | Description |
|-------|------|-------------|
| 1 | Observe | Full system running. Watch dashboards under load |
| 2 | Experiment | Swap components, change configs, see trade-offs |
| 3 | Build the Missing Piece | One critical component stubbed out |
| 4 | Fix the Broken System | Misconfigured system, diagnose and fix it |
| 5 | Build from Scratch | Service stubs and interfaces only |

### Skill-Based Entry
Not sure where to start? Take the diagnostic quiz:
```bash
make diagnose
```
It sizes you up and recommends your entry level.

### Branch-Based Learning
Each level is a git branch. Jump between them without
losing your progress:
```bash
git checkout level-1-observe/url-shortener
git checkout level-3-build/url-shortener
```

---

## Systems

### Individual Systems
- **URL Shortener** — consistent hashing, caching, sharding
- **Rate Limiter** — sliding window, token bucket
- **Feed System** — fan-out on write vs read, pub/sub
- **Distributed Cache** — eviction policies, cache invalidation
- **Notification Service** — message queues, at-least-once delivery

### Capstone — Twitter Clone
Integrates all individual systems. Forces real architectural
decisions across services.

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Services | Go, Python |
| Containers | Docker + Kubernetes (k3s) |
| Cache | Redis / Memcached (pluggable) |
| Database | PostgreSQL / Cassandra (pluggable) |
| Messaging | Kafka / RabbitMQ (pluggable) |
| Observability | Prometheus + Grafana |
| Traffic Sim | k6 / Locust |
| Chaos | Custom chaos toolkit |

---

## Pluggable Components

Swap implementations via config — no code changes needed:
```yaml
# config.yaml
cache:
  provider: redis   # or memcached, custom
database:
  provider: postgres  # or cassandra
queue:
  provider: kafka   # or rabbitmq
```

Redeploy and watch how behavior changes under load.

---

## AI Integration (BYOA — Bring Your Own Agent)

Use your own LLM to assist with challenges. Add your API
key to `.env`:
```bash
cp .env.example .env
# Add your key: OPENAI_API_KEY, ANTHROPIC_API_KEY, etc.
```

### AI Guidelines
- AI assistance is encouraged but not a shortcut
- Certain core checkpoints are AI-free zones
- Each challenge includes failure cases showing where AI falls short
- Decision journals log what you asked, what AI said, what you built

---

## Decision Journal

Before coding each challenge, fill in:
```markdown
## Decision Journal — [System Name] — [Level]

**Problem constraints:**
- Latency requirement:
- Scale requirement:
- Consistency requirement:

**Component choices:**
- Why [Redis] over [Memcached]?
- Why [PostgreSQL] over [Cassandra]?

**AI assistance:**
- What I asked:
- What AI suggested:
- What I actually implemented:
- Why they differed:

**Load test results:**
- Cache hit rate:
- P99 latency:
- Throughput:
```

---

## Chaos Toolkit

Intentionally break your system and learn resilience:
```bash
make chaos-kill-cache        # Kill Redis, observe fallback
make chaos-lag-network       # Add 200ms latency to DB
make chaos-corrupt-shard     # Corrupt one shard's data
make chaos-overload          # 10x normal traffic spike
```

---

## Getting Started

### Prerequisites
- Docker
- k3s or minikube
- Go 1.21+
- Python 3.11+

### Quick Start
```bash
git clone https://github.com/yourusername/systemdesignlab
cd systemdesignlab
make diagnose          # Find your level
make start             # Spin up infrastructure
make load-test         # Simulate traffic
make dashboard         # Open Grafana
```

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) to add new systems
or improve existing ones.

---

## Roadmap

- [ ] URL Shortener (all 5 levels)
- [ ] Rate Limiter
- [ ] Feed System
- [ ] Distributed Cache
- [ ] Notification Service
- [ ] Capstone Twitter Clone
- [ ] Diagnostic quiz
- [ ] Chaos toolkit
- [ ] BYOA AI integration