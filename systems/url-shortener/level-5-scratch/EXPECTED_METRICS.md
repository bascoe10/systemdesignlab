# Expected Metrics — URL Shortener Level 5

Same bar as Level 3, plus resilience. Latency targets are references for
typical hardware — `make validate` measures p99 against your own Level 1
baseline (1.5× your healthy number); ratios are absolute. Under
`make load-test` (1000 RPS steady state):

| Metric | Target | Where to look |
|--------|--------|---------------|
| Cache hit rate (steady state) | > 85% | Golden Signals → Cache hit rate |
| P99 latency (1000 RPS) | < 100ms | Golden Signals → Latency |
| P50 latency (1000 RPS) | < 15ms | Golden Signals → Latency |
| Throughput (sustained) | > 1000 RPS | Golden Signals → Traffic |
| Key distribution rel. stddev | < 15% across nodes | Deep Dive → ops per node |
| 5xx ratio | < 0.1% | Golden Signals → Errors |

Resilience (stretch — run `make chaos-kill-cache` during a load test):

| Metric | During outage | After restore |
|--------|---------------|---------------|
| Error rate | < 1% | < 0.1% |
| P99 latency | < 500ms (DB-only path) | < 100ms |
| Cache hit rate | 0% (expected) | > 85% within 60s |
