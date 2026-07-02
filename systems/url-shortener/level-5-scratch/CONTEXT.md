# Where You Are — Level 5: Build from Scratch

By now you've:
- **[Level 1]** Watched this system healthy — you know exactly what the
  dashboard should look like when you're done.
- **[Level 2]** Learned which config knobs move which signals.
- **[Level 3]** Built the hardest component, the consistent hash ring.
- **[Level 4]** Diagnosed a broken deployment from dashboard symptoms alone.

Now build the whole thing. The three services are empty skeletons that
serve `/healthz`, `/metrics`, and `501 Not Implemented` for everything else.
The infrastructure — Postgres, the cache nodes, Prometheus, the same Grafana
dashboards — is already wired and waiting. The dashboards light up as your
services come online.

```bash
sdl start        # everything boots; business endpoints return 501
sdl dashboard    # your empty canvas
```

Read `BRIEFING.md`. Journal your architecture decisions BEFORE coding —
including the SLOs you'd commit to (there's a dashboard for them).

This is the level that proves mastery. Take your time.
