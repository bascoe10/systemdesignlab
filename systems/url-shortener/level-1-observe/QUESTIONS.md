# Observation Questions — URL Shortener

Run `make load-test` and open `make dashboard`. Answer these (in writing —
articulating observations is the skill being trained):

## Golden Signals Overview

1. What is the steady-state p99 latency? What about p50? Why is the gap
   between them significant? (Hint: which requests take the slow path?)

2. Watch the cache hit rate panel. Why does it start near 0% and climb
   toward 90% over the first minute of the load test?

3. The traffic panel shows three services. Why does the redirector receive
   roughly 4x the shortener's traffic? Where is that ratio defined?

4. Which of the Four Golden Signals would degrade FIRST under sustained
   10x load? Form a hypothesis now — you'll test it with
   `make chaos-overload`.

## Component Deep Dive

5. Look at "Consistent hashing — ops per cache node". Are the three nodes
   evenly loaded? What property of the hash ring makes that true?

6. Run `make load-test SCENARIO=hot-key`. What happens to the per-node
   distribution? Why can't consistent hashing fix a hot key?

7. Compare "DB queries/sec" against the cache hit rate. Explain the
   relationship in one sentence.

## Chaos preview (optional, fun)

8. With a load test running, run `make chaos-kill-cache`, and watch the
   Chaos Impact dashboard. Did the system keep serving redirects? At what
   latency cost? How fast did the hit rate recover after restore?

When you can answer all of these, run `make validate`, then move on:
`git checkout level-2-experiment/url-shortener`
