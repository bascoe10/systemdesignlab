# ADR 0002: The CLI is the spine; workspaces replace generated branches

**Status:** accepted · 2026-07 · supersedes [ADR 0001](0001-generated-branches-are-swappable-distribution.md)

## Context

v1 distributed levels as CI-generated, force-pushed git branches. Building
it proved the cost: the branch model required a generator script, CI push
machinery, a guardrail to keep users off generated branches, and a
git-stash wrapper to make level switching survivable — an entire subsystem
whose job was defending users from the distribution mechanism. ADR 0001
anticipated this and reserved the right to swap the mechanism. We are
exercising that right before shipping, not after collecting friction data,
because the greenfield rewrite makes it free.

The product's core loop is *change something → dashboard moves → understand
why*. Every architectural choice is judged by time-to-that-moment.

## Decision

1. **One Go CLI (`sdl`) is the product spine.** Start, load tests,
   dashboards, validation, level switching, chaos, journal — all
   subcommands of one binary. No Makefile, no bash chaos scripts (Windows
   users exist; bash was the least testable code in v1).

2. **Levels are materialized workspaces, not branches.** `sdl start
   --level 3` assembles `workspace/` from `systems/<name>/shared/` + the
   level overlay — the same assembly logic v1's generator used, now in Go,
   pointed at a local directory instead of a branch. `workspace/` is
   gitignored, self-contained, and disposable (`sdl reset`).

3. **Level switching parks the whole workspace.** `sdl level N` moves the
   current workspace to `.sdl/park/level-<cur>/` and restores (or freshly
   materializes) the target. No git required, no stash conflicts, user
   code edits survive round-trips by construction.

4. **Cross-level state lives outside the workspace.** `.sdl/state.json`
   (current system/level), `.sdl/baseline.json` (the machine's calibrated
   healthy numbers), and `my-journal.md` (repo root) survive switching,
   parking, and resets.

5. **Reference implementations stay visible in `systems/`.** Solutions
   being one directory away is fine — the discipline is the learner's,
   same as the AI-free zones. Level 4's SOLUTIONS.md keeps its reveal
   ritual because the friction is pedagogically useful, not because it's
   secure.

## Consequences

- Deleted: `generator/`, `Makefile`, `chaos-toolkit/scripts/*.sh`, the
  generate-branches workflow, branch guardrails, the stash-based `switch`.
- Users never need git beyond `clone`. Contributors never think about
  branch generation.
- Infrastructure updates reach users on `sdl reset` / level switch instead
  of on push — acceptable for a learning lab, and `sdl status` can surface
  staleness later if it matters.
- The "diff two levels" trick (`git diff level-1... level-3...`) is gone;
  `sdl materialize` to two directories and `diff -r` replaces it for the
  few who want it.
