// Package ring implements consistent hashing with virtual nodes.
//
// This is the component you build yourself on the Level 3 branch — there,
// this file is replaced by a stub. The interface is the contract; the tests
// in ring_test.go define correctness.
package ring

import (
	"errors"
	"fmt"
	"hash/fnv"
	"sort"
	"sync"
)

// ErrEmptyRing is returned by GetNode when no nodes have been added.
var ErrEmptyRing = errors.New("ring: no nodes available")

// ConsistentHasher distributes keys across nodes using consistent hashing.
type ConsistentHasher interface {
	AddNode(nodeID string) error
	RemoveNode(nodeID string) error
	GetNode(key string) (string, error)
	Nodes() []string
}

type ring struct {
	mu           sync.RWMutex
	virtualNodes int
	hashes       []uint64          // sorted virtual-node positions
	owner        map[uint64]string // position -> node ID
	members      map[string]bool
}

// NewConsistentHasher returns a hash ring that places virtualNodes replicas
// of each node on the ring. More virtual nodes = smoother key distribution.
func NewConsistentHasher(virtualNodes int) ConsistentHasher {
	if virtualNodes < 1 {
		virtualNodes = 1
	}
	return &ring{
		virtualNodes: virtualNodes,
		owner:        make(map[uint64]string),
		members:      make(map[string]bool),
	}
}

func hashOf(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return fmix64(h.Sum64())
}

// fmix64 is the MurmurHash3 finalizer. Raw FNV of short, similar strings
// ("cache-1#0", "cache-1#1", …) clusters on the ring; this avalanche step
// spreads virtual nodes uniformly, which is what keeps key distribution
// stddev under the 15% target.
func fmix64(x uint64) uint64 {
	x ^= x >> 33
	x *= 0xff51afd7ed558ccd
	x ^= x >> 33
	x *= 0xc4ceb9fe1a85ec53
	x ^= x >> 33
	return x
}

func (r *ring) AddNode(nodeID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.members[nodeID] {
		return fmt.Errorf("ring: node %q already present", nodeID)
	}
	r.members[nodeID] = true
	for i := 0; i < r.virtualNodes; i++ {
		pos := hashOf(fmt.Sprintf("%s#%d", nodeID, i))
		// FNV collisions across distinct vnode labels are astronomically
		// unlikely; last writer wins if one ever occurs.
		if _, exists := r.owner[pos]; !exists {
			r.hashes = append(r.hashes, pos)
		}
		r.owner[pos] = nodeID
	}
	sort.Slice(r.hashes, func(i, j int) bool { return r.hashes[i] < r.hashes[j] })
	return nil
}

func (r *ring) RemoveNode(nodeID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.members[nodeID] {
		return fmt.Errorf("ring: node %q not present", nodeID)
	}
	delete(r.members, nodeID)
	kept := r.hashes[:0]
	for _, pos := range r.hashes {
		if r.owner[pos] == nodeID {
			delete(r.owner, pos)
		} else {
			kept = append(kept, pos)
		}
	}
	r.hashes = kept
	return nil
}

func (r *ring) GetNode(key string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.hashes) == 0 {
		return "", ErrEmptyRing
	}
	pos := hashOf(key)
	i := sort.Search(len(r.hashes), func(i int) bool { return r.hashes[i] >= pos })
	if i == len(r.hashes) {
		i = 0 // wrap around the ring
	}
	return r.owner[r.hashes[i]], nil
}

func (r *ring) Nodes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.members))
	for n := range r.members {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}
