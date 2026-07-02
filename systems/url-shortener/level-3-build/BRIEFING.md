# Briefing — Build the Consistent Hash Ring

## The Job

Implement `system/services/internal/ring/ring.go`. The interface contract:

```go
type ConsistentHasher interface {
    AddNode(nodeID string) error      // duplicate add is an error
    RemoveNode(nodeID string) error   // removing an absent node is an error
    GetNode(key string) (string, error) // deterministic owner for key
    Nodes() []string                  // current members, sorted
}

func NewConsistentHasher(virtualNodes int) ConsistentHasher
```

`GetNode` on an empty ring returns an error — the cache layer treats that
as "bypass cache, go to the database", which is exactly what the stub does
now for every key.

## AI_FREE_ZONE

Complete this component without AI assistance. Consistent hashing is
foundational — it appears in Dynamo, Cassandra, Memcached clients, CDNs,
and roughly every "design a distributed cache" interview. Build it yourself
first. Afterwards, read `docs/ai-failure-cases/consistent-hashing.md` to see
what an AI-generated version typically gets wrong.

## Think Through (before coding — journal these)

1. Why hash the NODES onto a ring at all? What breaks with `hash(key) % N`
   when N changes? (You saw the answer as a dashboard panel in Level 2.)
2. What do virtual nodes buy you? Why did `virtual_nodes: 1` skew the load?
3. Your ring positions come from hashing strings like `"cache-1#0"`,
   `"cache-1#1"`… Similar short strings can hash to clustered positions
   under weak hash functions. How do you get uniform spread?
4. `GetNode` is called on every request from multiple goroutines while
   nodes could be added/removed. What are the concurrency requirements?

## Definition of Done

```bash
make validate
```

1. **Required** — `go test ./internal/ring/` passes: contract behaviour,
   determinism, error cases.
2. **Performance** — distribution and rehashing tests pass: relative stddev
   across nodes < 15% with 128 virtual nodes; adding a node moves < 35% of
   keys.
3. **Live** — `make redeploy && make load-test`, then check the targets in
   `EXPECTED_METRICS.md` on the dashboard.
4. **Journal** — `my-journal.md` has your constraints, decisions, and load
   test numbers filled in.

## Hints (escalating — try without them first)

<details><summary>Hint 1: data structure</summary>
A sorted slice of ring positions plus a map from position to node ID.
GetNode = hash the key, binary search for the first position >= the hash,
wrap to index 0 past the end.
</details>

<details><summary>Hint 2: standard library</summary>
`hash/fnv` for hashing, `sort.Search` for the binary search,
`sync.RWMutex` for concurrent reads.
</details>

<details><summary>Hint 3: if distribution tests fail but logic looks right</summary>
FNV of short similar strings clusters. Run the hash output through an
avalanche/finalizer step (look up "murmur3 fmix64"), or hash with a
stronger function. This is exactly the failure mode described in the AI
failure case doc.
</details>
