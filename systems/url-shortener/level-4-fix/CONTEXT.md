# Where You Are — Level 4: Fix the Broken System

By now you've:
- **[Level 1]** Watched this system healthy. You know what good Golden
  Signals look like: p99 under 100ms, cache hit rate above 85%, three
  evenly-loaded cache nodes.
- **[Level 2]** Turned every knob in `config.yaml` and watched which panel
  each one moves.
- **[Level 3]** Built the consistent hash ring. You know exactly what keeps
  the per-node distribution flat.

Now: users are complaining that redirects are slow. The system is deployed
and "running" — no crashes, no obvious errors — but something is very wrong.

```bash
sdl start && sdl load
sdl dashboard
```

Read `KNOWN_ISSUES.md` for the symptoms. Diagnose from the dashboards, fix
the configuration, `sdl restart && sdl load`, repeat until the
Golden Signals match your Level 1 baseline. `sdl validate` checks the live
metrics against healthy thresholds.

Rules of engagement:
- The bugs are all in configuration — you don't need to touch Go code.
- Journal each diagnosis: symptom → hypothesis → evidence → fix. That
  narrative is the debugging answer interviewers want to hear.
- `SOLUTIONS.md` exists but stays hidden until you run
  `sdl reveal-solution`. Diagnosis is the entire point of this level.

Next: `sdl level 5`
