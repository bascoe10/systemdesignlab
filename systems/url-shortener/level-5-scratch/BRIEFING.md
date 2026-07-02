# Briefing — Build from Scratch

## What You Have

- **API contracts** in `system/contracts/` (OpenAPI) — the exact surface
  each service must expose. The tests and load tests are written against
  these, not against any particular implementation.
- **Skeleton mains** in `system/services/{api-gateway,shortener,redirector}/`
  — they compile, boot, and expose `/metrics` + `/healthz` so the stack
  starts before you've written a line.
- **Building blocks** in `system/services/internal/` — config loader, cache
  providers, Postgres store, metrics helpers. Using them is allowed and
  sensible (real engineers don't rewrite the Redis client); replacing them
  is also allowed. The RING implementation is stubbed — bring your Level 3
  solution or rebuild it.
- **The full test suite**: `go test ./internal/ring/` for the ring, and
  `go test -tags integration ./integration/` against the running stack.
- **The same dashboards and EXPECTED_METRICS.md** as Level 3.

## The Contract (prose version)

1. `POST /api/shorten` `{"url": "https://..."}` → `201 {"code", "short_url"}`;
   reject non-absolute/non-http(s) URLs with 400.
2. `GET /r/{code}` → `302 Location: <target>`; unknown code → 404.
3. Writes go through the shortener; reads through the redirector; both sit
   behind the gateway, which rate-limits per client IP (429 when exceeded).
4. Reads must be cache-accelerated: write-through on create, cache-aside
   with backfill on read miss. Keys are sharded across the three cache
   nodes by consistent hashing.
5. Cache failure must degrade, not break: if cache nodes are unreachable,
   serve from the database. `make chaos-kill-cache` is the acceptance test.
6. Instrument everything. If `make dashboard` stays dark, it didn't happen.
   Emit the metric names in `internal/obs` — the dashboards depend on them.

## Suggested Path

Start with `docker-compose` already up. Get ONE endpoint working end to end
(`shorten` → row in Postgres), then the read path DB-only, then layer the
cache in, then the ring, then the gateway rate limit. Run the integration
tests after each step — they're written to pass incrementally.

## Definition of Done

```bash
make validate
```

- Required: ring unit tests + integration tests pass.
- Performance: `EXPECTED_METRICS.md` targets met under `make load-test`.
- Stretch: survive `make chaos-kill-cache` with error rate < 1%.
- Journal: architecture decisions AND your SLO definitions with reasoning.
