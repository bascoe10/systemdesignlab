# Known Issues — URL Shortener Level 4

The system is deployed and running but something is very wrong. Use the
Grafana dashboards to diagnose and fix these issues. Symptoms only — the
root causes are yours to find.

## Symptom 1: Terrible Cache Performance
Cache hit rate hovers near zero under normal load — expected is 85%+.
The system is hitting the database for almost every request, even for URLs
that were just read a moment ago.

## Symptom 2: High Tail Latency
P99 latency is hundreds of milliseconds under a steady 1000 RPS — expected
is under 100ms. P50 looks almost normal. Users are complaining about slow
redirects, but averages look "fine".

## Symptom 3: Uneven Load Distribution
One cache node is handling the large majority of all cache operations.
The other two are nearly idle. (Component Deep Dive → ops per cache node.)

## Symptom 4: Memory Pressure
The cache is pinned at its memory limit despite modest traffic, and the
eviction rate is continuously high.

---

There are **four distinct root causes**. Some symptoms share a cause's
side effects — part of the diagnosis is working out which is which.

For each one, journal: the symptom → which dashboard panel implicates which
component → your hypothesis → the fix → the panel after `sdl restart`.

DO NOT run `sdl reveal-solution` until you've diagnosed and fixed all four.
The challenge is in the diagnosis.

Done when: `sdl validate` passes (live metrics back at healthy baselines).
