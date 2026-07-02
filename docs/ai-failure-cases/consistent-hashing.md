# AI Failure Case: Consistent Hashing — URL Shortener

> **Last verified: 2026-07.** Models improve; claims about them rot. This
> document is a *reproducible protocol*, not an article of faith — run it
> against your own model (below) and record what you find in your
> journal's AI Assistance Log. If current models pass cleanly, open a PR
> updating this file: that's a finding too.

## Run It Yourself (5 minutes)

1. Give your model the naive prompt below — no extra context.
2. Drop its output into `system/services/internal/ring/ring.go`
   (adapt to the `ConsistentHasher` interface if needed).
3. `cd system/services && go test ./internal/ring/ -v`
4. Record: which tiers passed? What was the distribution stddev? Run it
   under `-race` with concurrent access if the tests pass.

## The Naive Prompt

> "Implement consistent hashing in Go for a URL shortener"

## The AI's Response (typical)

A basic hash ring: nodes hashed once each (or with a handful of virtual
nodes), positions from `crc32`/`md5` of `"node-name" + strconv.Itoa(i)`,
a sorted slice, binary search for lookup. It compiles, the happy-path
works, and a quick manual test looks fine.

## Why It Fails Under Load

Run the Level 3 test suite against a typical AI implementation and watch
the *performance tier* fail while the *required tier* passes:

- **Key distribution stddev: 20–45%** (target: <15%). Two causes stack:
  too few virtual nodes, and weak hash mixing — FNV/CRC of short similar
  strings (`cache-1#0`, `cache-1#1`, …) produce clustered ring positions.
  This repo's own first implementation hit 20.2% stddev with 128 virtual
  nodes on raw FNV-64a; a murmur3 `fmix64` finalizer brought it under 9%.
  The lab's tests caught it. A code review probably wouldn't have.
- **No concurrency contract.** `GetNode` races `AddNode`/`RemoveNode`.
  Under `-race` with real traffic it's a crash; in production it's a
  sporadic wrong-node lookup that looks like random cache misses.
- **Silent empty-ring behaviour.** Returning `""` instead of an error means
  the caller caches to a node named "" — the failure surfaces two layers
  away, in the cache client.

## The Dashboard Tells the Story

An AI ring that "works" still shows up on Component Deep Dive as one cache
node doing 60%+ of the ops. Correctness by test-pass is not correctness by
distribution. This is why the lab makes you LOOK at the panels.

## The Lesson

AI implementations optimize for plausibility, and consistent hashing is a
concept where plausible and correct diverge quantitatively, not visibly.
You can only evaluate the AI's output if you know what the distribution
target is and where to see it. That's the skill — the code was never the
hard part.

If you use AI at Level 5 (allowed), log the prompt, the response, and what
you changed in your decision journal's AI Assistance Log — then run the
distribution tests and record the numbers.
