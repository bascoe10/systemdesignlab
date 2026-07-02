# Expected Metrics — URL Shortener Level 3

When your ring implementation is correct, a steady-state load test
(`sdl load`, 1000 RPS, 2 minutes) should show:

| Metric | Target | Where to look |
|--------|--------|---------------|
| Cache hit rate (steady state) | > 85% | Golden Signals → Cache hit rate |
| P99 latency (1000 RPS) | < 100ms | Golden Signals → Latency |
| P50 latency (1000 RPS) | < 15ms | Golden Signals → Latency |
| Throughput (sustained) | > 1000 RPS | Golden Signals → Traffic |
| Key distribution rel. stddev | < 15% across nodes | Deep Dive → ops per cache node |
| Cache read outcome `bypass` | ~0 | Deep Dive → Cache read outcomes |

Notes:
- Hit rate needs ~60s of load to warm up from 0% — judge the steady state,
  not the first minute.
- Latency numbers above are references for typical hardware. `sdl validate`
  judges p99 against **your own Level 1 baseline** (`.sdl/baseline.json`,
  captured when you validated Level 1): pass = within 1.5× your healthy
  p99. If you skipped Level 1, generic bounds apply — go calibrate, it
  takes 15 minutes.
- Ratios (hit rate, distribution stddev) are machine-independent and stay
  absolute.
