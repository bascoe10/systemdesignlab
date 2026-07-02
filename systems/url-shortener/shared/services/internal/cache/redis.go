package cache

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisProvider keeps one client per cache node.
type RedisProvider struct {
	clients map[string]*redis.Client
}

func NewRedisProvider(nodes []string) *RedisProvider {
	clients := make(map[string]*redis.Client, len(nodes))
	for _, n := range nodes {
		clients[n] = redis.NewClient(&redis.Options{
			Addr:         fmt.Sprintf("%s:6379", n),
			DialTimeout:  500 * time.Millisecond,
			ReadTimeout:  500 * time.Millisecond,
			WriteTimeout: 500 * time.Millisecond,
		})
	}
	return &RedisProvider{clients: clients}
}

func (p *RedisProvider) client(node string) (*redis.Client, error) {
	c, ok := p.clients[node]
	if !ok {
		return nil, fmt.Errorf("redis: unknown node %q", node)
	}
	return c, nil
}

func (p *RedisProvider) Get(ctx context.Context, node, key string) (string, error) {
	c, err := p.client(node)
	if err != nil {
		return "", err
	}
	val, err := c.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrMiss
	}
	return val, err
}

func (p *RedisProvider) Set(ctx context.Context, node, key, value string, ttl time.Duration) error {
	c, err := p.client(node)
	if err != nil {
		return err
	}
	return c.Set(ctx, key, value, ttl).Err()
}

// ApplyServerConfig pushes maxmemory and eviction policy to every node.
// Runtime CONFIG SET is what lets Level 2/4 users change cache behaviour
// from config.yaml without rebuilding containers.
func (p *RedisProvider) ApplyServerConfig(ctx context.Context, maxMemoryBytes int64, evictionPolicy string) error {
	var firstErr error
	for node, c := range p.clients {
		if maxMemoryBytes > 0 {
			if err := c.ConfigSet(ctx, "maxmemory", strconv.FormatInt(maxMemoryBytes, 10)).Err(); err != nil && firstErr == nil {
				firstErr = fmt.Errorf("node %s: %w", node, err)
			}
		}
		if evictionPolicy != "" {
			if err := c.ConfigSet(ctx, "maxmemory-policy", evictionPolicy).Err(); err != nil && firstErr == nil {
				firstErr = fmt.Errorf("node %s: %w", node, err)
			}
		}
	}
	return firstErr
}

func (p *RedisProvider) Stats(ctx context.Context, node string) (NodeStats, error) {
	c, err := p.client(node)
	if err != nil {
		return NodeStats{}, err
	}
	var s NodeStats
	info, err := c.Info(ctx, "memory", "stats").Result()
	if err != nil {
		return NodeStats{}, err
	}
	for _, line := range strings.Split(info, "\n") {
		k, v, ok := strings.Cut(strings.TrimSpace(line), ":")
		if !ok {
			continue
		}
		switch k {
		case "used_memory":
			s.UsedBytes, _ = strconv.ParseInt(v, 10, 64)
		case "maxmemory":
			s.MaxBytes, _ = strconv.ParseInt(v, 10, 64)
		case "evicted_keys":
			s.Evictions, _ = strconv.ParseInt(v, 10, 64)
		}
	}
	s.Items, _ = c.DBSize(ctx).Result()
	return s, nil
}

func (p *RedisProvider) Close() error {
	var firstErr error
	for _, c := range p.clients {
		if err := c.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
