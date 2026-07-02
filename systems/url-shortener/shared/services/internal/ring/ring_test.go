package ring

import (
	"fmt"
	"math"
	"testing"
)

// --- Required tier: interface contract -----------------------------------

func TestEmptyRingReturnsError(t *testing.T) {
	r := NewConsistentHasher(128)
	if _, err := r.GetNode("anything"); err == nil {
		t.Fatal("GetNode on empty ring must return an error")
	}
}

func TestSingleNodeGetsAllKeys(t *testing.T) {
	r := NewConsistentHasher(128)
	mustAdd(t, r, "cache-1")
	for i := 0; i < 100; i++ {
		node, err := r.GetNode(fmt.Sprintf("key-%d", i))
		if err != nil {
			t.Fatalf("GetNode: %v", err)
		}
		if node != "cache-1" {
			t.Fatalf("expected cache-1, got %s", node)
		}
	}
}

func TestGetNodeIsDeterministic(t *testing.T) {
	r := NewConsistentHasher(128)
	mustAdd(t, r, "cache-1", "cache-2", "cache-3")
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key-%d", i)
		first, _ := r.GetNode(key)
		for j := 0; j < 5; j++ {
			again, _ := r.GetNode(key)
			if again != first {
				t.Fatalf("key %s mapped to %s then %s", key, first, again)
			}
		}
	}
}

func TestDuplicateAddFails(t *testing.T) {
	r := NewConsistentHasher(8)
	mustAdd(t, r, "cache-1")
	if err := r.AddNode("cache-1"); err == nil {
		t.Fatal("adding a duplicate node must fail")
	}
}

func TestRemoveMissingNodeFails(t *testing.T) {
	r := NewConsistentHasher(8)
	if err := r.RemoveNode("nope"); err == nil {
		t.Fatal("removing an absent node must fail")
	}
}

func TestRemovedNodeReceivesNoKeys(t *testing.T) {
	r := NewConsistentHasher(128)
	mustAdd(t, r, "cache-1", "cache-2", "cache-3")
	if err := r.RemoveNode("cache-2"); err != nil {
		t.Fatalf("RemoveNode: %v", err)
	}
	for i := 0; i < 1000; i++ {
		node, err := r.GetNode(fmt.Sprintf("key-%d", i))
		if err != nil {
			t.Fatalf("GetNode: %v", err)
		}
		if node == "cache-2" {
			t.Fatal("key mapped to a removed node")
		}
	}
}

// --- Performance tier: distribution quality ------------------------------

// With enough virtual nodes, keys should spread evenly. Target: relative
// standard deviation across nodes below 15% (matches EXPECTED_METRICS.md).
func TestKeyDistributionIsBalanced(t *testing.T) {
	r := NewConsistentHasher(128)
	mustAdd(t, r, "cache-1", "cache-2", "cache-3")

	counts := map[string]int{}
	const keys = 10000
	for i := 0; i < keys; i++ {
		node, err := r.GetNode(fmt.Sprintf("key-%d", i))
		if err != nil {
			t.Fatalf("GetNode: %v", err)
		}
		counts[node]++
	}

	mean := float64(keys) / 3.0
	var variance float64
	for _, c := range counts {
		d := float64(c) - mean
		variance += d * d
	}
	variance /= 3.0
	relStdDev := math.Sqrt(variance) / mean
	if relStdDev > 0.15 {
		t.Fatalf("key distribution too skewed: relative stddev %.1f%% (want < 15%%), counts=%v",
			relStdDev*100, counts)
	}
}

// The whole point of consistent hashing: adding a node moves only ~1/N of
// keys. A naive mod-N hash would move ~(N-1)/N of them.
func TestMinimalRehashingOnNodeAdd(t *testing.T) {
	r := NewConsistentHasher(128)
	mustAdd(t, r, "cache-1", "cache-2", "cache-3")

	const keys = 10000
	before := make([]string, keys)
	for i := 0; i < keys; i++ {
		before[i], _ = r.GetNode(fmt.Sprintf("key-%d", i))
	}

	mustAdd(t, r, "cache-4")

	moved := 0
	for i := 0; i < keys; i++ {
		after, _ := r.GetNode(fmt.Sprintf("key-%d", i))
		if after != before[i] {
			moved++
		}
	}
	frac := float64(moved) / keys
	// Ideal is 1/4 = 25%; allow slack for imperfect balance.
	if frac > 0.35 {
		t.Fatalf("adding one node rehashed %.1f%% of keys (want < 35%%)", frac*100)
	}
	if moved == 0 {
		t.Fatal("adding a node moved zero keys — new node receives no traffic")
	}
}

func mustAdd(t *testing.T, r ConsistentHasher, nodes ...string) {
	t.Helper()
	for _, n := range nodes {
		if err := r.AddNode(n); err != nil {
			t.Fatalf("AddNode(%s): %v", n, err)
		}
	}
}
