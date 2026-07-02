# Decision Journal

Copy happens automatically: `make journal` creates `my-journal.md` from this
template (gitignored — it stays personal). Required for Levels 3–5;
`make validate` checks that the required sections are filled in.

This journal IS the interview prep. Interviewers evaluate how you articulate
constraints and trade-offs, not whether you memorized an architecture.

---

## System: url-shortener | Level: ___ | Date: ___

### Constraints I Identified
- Latency requirement:
- Scale requirement:
- Consistency vs availability trade-off:
- Read/write ratio:

### Component Decisions
| Decision | Option A | Option B | I Chose | Because |
|----------|----------|----------|---------|---------|
|          |          |          |         |         |

### What Surprised Me


### What I Would Do Differently at 10x Scale


### SLOs I Would Set (Level 5)
| SLI | SLO | Why this threshold |
|-----|-----|--------------------|
|     |     |                    |

### AI Assistance Log
#### Attempt 1
- Prompt: [what I asked]
- AI Response Summary: [what it suggested]
- Why I accepted/rejected it: [my reasoning]
- What I changed: [my modification]

### Load Test Results
| Metric | Target | Actual | Pass/Fail |
|--------|--------|--------|-----------|
| Cache hit rate (steady state) | > 85% | | |
| P99 latency (1000 RPS) | < 100ms | | |
| P50 latency (1000 RPS) | < 15ms | | |
| Throughput (sustained) | > 1000 RPS | | |
