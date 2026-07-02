// Package cache provides a pluggable, ring-sharded cache client.
//
// A Provider talks to one kind of backend (Redis or Memcached) across all
// nodes. The Cluster picks the node for each key via consistent hashing.
// If the ring cannot resolve a node (e.g. the Level 3 stub), the Cluster
// degrades to a cache bypass: reads report a miss, writes are dropped, and
// the system keeps serving from the database.
package cache

import (
	"context"
	"errors"
	"time"

	"github.com/systemdesignlab/url-shortener/internal/ring"
)

// ErrMiss is returned by providers when the key is absent.
var ErrMiss = errors.New("cache: miss")

// Provider is the pluggable cache backend. Implementations exist for Redis
// and Memcached; both are selected purely by config.
type Provider interface {
	Get(ctx context.Context, node, key string) (string, error)
	Set(ctx context.Context, node, key, value string, ttl time.Duration) error
	// Stats reports per-node saturation numbers for observability.
	Stats(ctx context.Context, node string) (NodeStats, error)
	Close() error
}

// NodeStats is a snapshot of one cache node, exported as Prometheus gauges.
type NodeStats struct {
	UsedBytes int64
	MaxBytes  int64
	Items     int64
	Evictions int64
}

// Result classifies the outcome of a Cluster read for metrics.
type Result string

const (
	Hit     Result = "hit"
	Miss    Result = "miss"
	OK      Result = "ok"     // successful write
	Bypass  Result = "bypass" // ring unavailable — cache skipped entirely
	Errored Result = "error"
)

// Cluster shards keys across nodes with a ConsistentHasher.
type Cluster struct {
	Ring     ring.ConsistentHasher
	Provider Provider
	TTL      time.Duration

	// OnResult and OnNodeOp are metric hooks; nil-safe.
	OnResult func(op string, r Result)
	OnNodeOp func(node string)
}

func (c *Cluster) report(op string, r Result) {
	if c.OnResult != nil {
		c.OnResult(op, r)
	}
}

// Get returns (value, true) on a hit. Any ring or provider failure degrades
// to a miss so callers always have the database as fallback.
func (c *Cluster) Get(ctx context.Context, key string) (string, bool) {
	node, err := c.Ring.GetNode(key)
	if err != nil {
		c.report("get", Bypass)
		return "", false
	}
	if c.OnNodeOp != nil {
		c.OnNodeOp(node)
	}
	val, err := c.Provider.Get(ctx, node, key)
	switch {
	case err == nil:
		c.report("get", Hit)
		return val, true
	case errors.Is(err, ErrMiss):
		c.report("get", Miss)
	default:
		c.report("get", Errored)
	}
	return "", false
}

// Set writes through to the owning node. Failures are reported, not fatal.
func (c *Cluster) Set(ctx context.Context, key, value string) {
	node, err := c.Ring.GetNode(key)
	if err != nil {
		c.report("set", Bypass)
		return
	}
	if c.OnNodeOp != nil {
		c.OnNodeOp(node)
	}
	if err := c.Provider.Set(ctx, node, key, value, c.TTL); err != nil {
		c.report("set", Errored)
		return
	}
	c.report("set", OK)
}
