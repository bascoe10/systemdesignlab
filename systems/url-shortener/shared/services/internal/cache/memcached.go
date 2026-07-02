package cache

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

// MemcachedProvider keeps one client per cache node.
type MemcachedProvider struct {
	clients map[string]*memcache.Client
	addrs   map[string]string
}

func NewMemcachedProvider(nodes []string) *MemcachedProvider {
	clients := make(map[string]*memcache.Client, len(nodes))
	addrs := make(map[string]string, len(nodes))
	for _, n := range nodes {
		addr := fmt.Sprintf("%s:11211", n)
		c := memcache.New(addr)
		c.Timeout = 500 * time.Millisecond
		clients[n] = c
		addrs[n] = addr
	}
	return &MemcachedProvider{clients: clients, addrs: addrs}
}

func (p *MemcachedProvider) client(node string) (*memcache.Client, error) {
	c, ok := p.clients[node]
	if !ok {
		return nil, fmt.Errorf("memcached: unknown node %q", node)
	}
	return c, nil
}

func (p *MemcachedProvider) Get(_ context.Context, node, key string) (string, error) {
	c, err := p.client(node)
	if err != nil {
		return "", err
	}
	item, err := c.Get(key)
	if errors.Is(err, memcache.ErrCacheMiss) {
		return "", ErrMiss
	}
	if err != nil {
		return "", err
	}
	return string(item.Value), nil
}

func (p *MemcachedProvider) Set(_ context.Context, node, key, value string, ttl time.Duration) error {
	c, err := p.client(node)
	if err != nil {
		return err
	}
	return c.Set(&memcache.Item{Key: key, Value: []byte(value), Expiration: int32(ttl.Seconds())})
}

// Stats issues the raw "stats" command; gomemcache does not expose it.
func (p *MemcachedProvider) Stats(ctx context.Context, node string) (NodeStats, error) {
	addr, ok := p.addrs[node]
	if !ok {
		return NodeStats{}, fmt.Errorf("memcached: unknown node %q", node)
	}
	d := net.Dialer{Timeout: 500 * time.Millisecond}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return NodeStats{}, err
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(time.Second))
	if _, err := fmt.Fprintf(conn, "stats\r\n"); err != nil {
		return NodeStats{}, err
	}
	var s NodeStats
	sc := bufio.NewScanner(conn)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "END" {
			break
		}
		fields := strings.Fields(line) // "STAT <name> <value>"
		if len(fields) != 3 || fields[0] != "STAT" {
			continue
		}
		v, _ := strconv.ParseInt(fields[2], 10, 64)
		switch fields[1] {
		case "bytes":
			s.UsedBytes = v
		case "limit_maxbytes":
			s.MaxBytes = v
		case "curr_items":
			s.Items = v
		case "evictions":
			s.Evictions = v
		}
	}
	return s, sc.Err()
}

func (p *MemcachedProvider) Close() error { return nil }
