# Where You Are — Level 3: Build the Missing Piece

By now you've:
- **[Level 1]** Watched this system healthy. You know the baseline: hit rate
  above 85%, p99 under 100ms, three evenly-loaded cache nodes.
- **[Level 2]** Turned the knobs. You saw with your own eyes that
  `virtual_nodes: 1` wrecks the node distribution — and now you know why
  that knob exists.

Now the consistent hash ring is gone. `system/services/internal/ring/ring.go`
is a stub that returns errors, so every cache lookup degrades to a bypass:
the system still works, but 100% of reads hit the database.

Start the stack and look at the dashboard BEFORE writing code:

```bash
make start && make load-test
```

Component Deep Dive → "Cache read outcomes" shows everything as `bypass`.
That's your "before" picture. Your implementation is done when the same
panel shows >85% hits and the per-node panel shows three even lines.

Read `BRIEFING.md`, then `make journal` — the decision journal is required
at this level.

Next: `git checkout level-4-fix/url-shortener`
