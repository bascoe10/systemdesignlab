package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/systemdesignlab/url-shortener/internal/config"
	"github.com/systemdesignlab/url-shortener/internal/obs"
	"github.com/systemdesignlab/url-shortener/internal/ring"
)

// NewClusterFromConfig wires provider + hash ring from config.yaml.
// Ring errors are deliberately non-fatal: on the Level 3 branch the ring is
// a stub that returns errors, and the system must still boot and serve
// traffic straight from the database.
func NewClusterFromConfig(cfg *config.Config) (*Cluster, error) {
	var provider Provider
	switch cfg.Cache.Provider {
	case "redis":
		provider = NewRedisProvider(cfg.Cache.Nodes)
	case "memcached":
		provider = NewMemcachedProvider(cfg.Cache.Nodes)
	default:
		return nil, fmt.Errorf("unknown cache provider %q", cfg.Cache.Provider)
	}

	r := ring.NewConsistentHasher(cfg.Hashing.VirtualNodes)
	for _, n := range cfg.Cache.Nodes {
		if err := r.AddNode(n); err != nil {
			log.Printf("ring: AddNode(%s): %v (cache will be bypassed)", n, err)
		}
	}

	ttl, err := cfg.CacheTTL()
	if err != nil {
		return nil, err
	}
	return &Cluster{
		Ring:     r,
		Provider: provider,
		TTL:      ttl,
		OnResult: func(op string, res Result) { obs.CacheResult(op, string(res)) },
		OnNodeOp: obs.CacheNodeOp,
	}, nil
}

// StartMaintenance runs a background loop that (a) applies runtime cache
// server config for Redis and (b) exports per-node saturation stats.
// Run it from exactly one service (the shortener) to avoid duplicate writes.
func StartMaintenance(cfg *config.Config, provider Provider) {
	apply := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if rp, ok := provider.(*RedisProvider); ok {
			maxMem, err := cfg.CacheMaxMemoryBytes()
			if err != nil {
				log.Printf("cache: bad max_memory: %v", err)
			} else if err := rp.ApplyServerConfig(ctx, maxMem, cfg.Cache.EvictionPolicy); err != nil {
				log.Printf("cache: apply server config: %v", err)
			}
		}

		for _, node := range cfg.Cache.Nodes {
			s, err := provider.Stats(ctx, node)
			if err != nil {
				continue // node down (chaos experiments do this on purpose)
			}
			obs.CacheNodeMemUsed.WithLabelValues(node).Set(float64(s.UsedBytes))
			obs.CacheNodeMemMax.WithLabelValues(node).Set(float64(s.MaxBytes))
			obs.CacheNodeItems.WithLabelValues(node).Set(float64(s.Items))
			obs.CacheNodeEvictions.WithLabelValues(node).Set(float64(s.Evictions))
		}
	}
	apply()
	go func() {
		for range time.Tick(10 * time.Second) {
			apply()
		}
	}()
}
