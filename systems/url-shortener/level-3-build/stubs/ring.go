// Package ring implements consistent hashing with virtual nodes.
//
// LEVEL 3 CHALLENGE: this is a stub. Every method returns an error, so the
// cache layer bypasses caching entirely and all reads fall through to the
// database. Your job is to replace this with a real hash ring.
//
// AI_FREE_ZONE: Complete this section without AI assistance.
// Consistent hashing is foundational. Build it yourself first.
//
// Read BRIEFING.md for the contract, the thinking prompts, and escalating
// hints. The tests in ring_test.go define correctness:
//
//	go test ./internal/ring/
//
// TODO: Implement the hash ring. Consider:
//   - How do you handle node addition/removal without rehashing all keys?
//   - What is the role of virtual nodes in balancing load?
//   - What are the concurrency requirements? GetNode is called on every
//     request from many goroutines.
//   - Hint: research "consistent hashing" and "virtual nodes".
package ring

import "errors"

// ErrEmptyRing is returned by GetNode when no nodes have been added.
var ErrEmptyRing = errors.New("ring: no nodes available")

var errNotImplemented = errors.New("ring: not implemented — this is the Level 3 challenge, see BRIEFING.md")

// ConsistentHasher distributes keys across nodes using consistent hashing.
type ConsistentHasher interface {
	AddNode(nodeID string) error
	RemoveNode(nodeID string) error
	GetNode(key string) (string, error)
	Nodes() []string
}

type stubRing struct{}

// NewConsistentHasher returns a hash ring that places virtualNodes replicas
// of each node on the ring. More virtual nodes = smoother key distribution.
func NewConsistentHasher(virtualNodes int) ConsistentHasher {
	// TODO: implement
	return &stubRing{}
}

func (s *stubRing) AddNode(nodeID string) error {
	// TODO: implement. Duplicate adds must return an error.
	return errNotImplemented
}

func (s *stubRing) RemoveNode(nodeID string) error {
	// TODO: implement. Removing an absent node must return an error.
	return errNotImplemented
}

func (s *stubRing) GetNode(key string) (string, error) {
	// TODO: implement. Must be deterministic: the same key always maps to
	// the same node while membership is unchanged.
	return "", errNotImplemented
}

func (s *stubRing) Nodes() []string {
	// TODO: implement. Return current members, sorted.
	return nil
}
