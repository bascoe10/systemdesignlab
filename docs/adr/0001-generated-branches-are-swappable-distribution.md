# ADR 0001: Generated branches are a swappable distribution mechanism

**Status:** superseded by [ADR 0002](0002-cli-materialized-workspaces.md) · 2026-07

## Decision

Level branches (`level-N-*/<system>`) are build artifacts generated from
`main` by `generator/generate.sh`. All content authoring happens on `main`;
branches are force-pushed by CI and never edited directly.

## Consequence we are counting on

Because distribution is decoupled from authoring, the branch mechanism is
**replaceable without touching content**. If real-world usage shows git
branches are too hostile for the audience (stranded commits, pull
confusion, branch clutter), the fallbacks are, in order:

1. `sdl get <system> <level>` — materialize a level into a plain directory
   from a release tarball; no git required.
2. A separate `systemdesignlab-labs` distribution repo, keeping this repo
   contributor-only.

Both reuse `assemble()` in the generator as-is. Do not build either until
user friction data demands it; do not let any new feature assume branches
are permanent.

## Guardrails shipped with the branch model (v1, historical)

- `make start`/`make redeploy` auto-moved users off generated code branches
  (levels 3–5) onto `my-progress/*` branches; levels 1–2 got a notice.
- `sdl switch <level>` wrapped the stash/checkout/restore dance.
- A CI force-push moves only the remote ref — local commits survive in the
  local branch; stranding requires pulling a generated branch, which the
  guardrails make unlikely.
