# Resilience Challenge: Overload

## What Gets Injected

`sdl chaos overload` — ~5x steady-state traffic (default 5000 rps) for two
minutes (`--rate 10000 --duration 1m` to change).

## What to Look For in Grafana

- **Which Golden Signal degrades first?** You predicted this in Level 1
  Question 4 — now check your answer.
- **The saturation cascade:** k6 VUs ramp → gateway → redirector → cache →
  (misses) → DB pool. Find the first panel that pins at its ceiling; that's
  your bottleneck. Everything downstream of it is actually *protected* by
  it.
- **Rate limiter engagement:** if you tuned `gateway.rate_limit_rps` low,
  429s appear and the backend stays healthy — load shedding working as
  designed. If it's high, the backend eats the full spike.

## The Fix

<details>
<summary>Click to reveal</summary>

"Fixing" overload means choosing where to fail:

- **Shed at the edge:** set the gateway rate limit just below the measured
  breaking point. Users get fast 429s instead of slow timeouts.
- **Scale the bottleneck:** whatever pinned first is the scaling candidate.
  Journal what you'd scale and what the *next* bottleneck would be.
- **Protect the database:** the pool size cap is itself backpressure —
  starving it (Level 4 style) trades errors for tail latency. Know which
  trade you're making.
</details>

## Success Criteria

| Metric | Target |
|--------|--------|
| Breaking point | measured and journaled (rps) |
| First saturated resource | identified with panel evidence |
| Error mode | explained: shed (429) vs degraded (latency) vs failed (5xx) |
