// Package config loads the lab-wide config.yaml that drives every service.
// The file lives at the workspace root and is mounted read-only into each
// container at /app/config.yaml. Levels 2 and 4 are built around editing it.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Cache    Cache    `yaml:"cache"`
	Hashing  Hashing  `yaml:"hashing"`
	Database Database `yaml:"database"`
	Gateway  Gateway  `yaml:"gateway"`
	Services Services `yaml:"services"`
}

type Cache struct {
	Provider       string   `yaml:"provider"` // redis | memcached
	Nodes          []string `yaml:"nodes"`
	TTL            string   `yaml:"ttl"`
	MaxMemory      string   `yaml:"max_memory"`
	EvictionPolicy string   `yaml:"eviction_policy"`
}

type Hashing struct {
	VirtualNodes int `yaml:"virtual_nodes"`
}

type Database struct {
	Provider string `yaml:"provider"` // postgres
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	PoolSize int    `yaml:"pool_size"`
}

type Gateway struct {
	RateLimitRPS int `yaml:"rate_limit_rps"`
	Burst        int `yaml:"burst"`
}

type Services struct {
	ShortenerURL  string `yaml:"shortener_url"`
	RedirectorURL string `yaml:"redirector_url"`
}

// Load reads CONFIG_PATH (default /app/config.yaml).
func Load() (*Config, error) {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "/app/config.yaml"
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) validate() error {
	switch c.Cache.Provider {
	case "redis", "memcached":
	default:
		return fmt.Errorf("cache.provider must be redis or memcached, got %q", c.Cache.Provider)
	}
	if len(c.Cache.Nodes) == 0 {
		return fmt.Errorf("cache.nodes must list at least one node")
	}
	if c.Hashing.VirtualNodes < 1 {
		return fmt.Errorf("hashing.virtual_nodes must be >= 1")
	}
	if c.Database.PoolSize < 1 {
		return fmt.Errorf("database.pool_size must be >= 1")
	}
	return nil
}

// CacheTTL parses the cache.ttl duration (e.g. "24h", "1s").
func (c *Config) CacheTTL() (time.Duration, error) {
	if c.Cache.TTL == "" {
		return 24 * time.Hour, nil
	}
	d, err := time.ParseDuration(c.Cache.TTL)
	if err != nil {
		return 0, fmt.Errorf("cache.ttl: %w", err)
	}
	return d, nil
}

// CacheMaxMemoryBytes parses cache.max_memory ("256mb", "8mb", "1gb").
func (c *Config) CacheMaxMemoryBytes() (int64, error) {
	return ParseSize(c.Cache.MaxMemory)
}

// ParseSize converts human sizes like "256mb" to bytes.
func ParseSize(s string) (int64, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return 0, nil
	}
	mult := int64(1)
	switch {
	case strings.HasSuffix(s, "gb"):
		mult, s = 1<<30, strings.TrimSuffix(s, "gb")
	case strings.HasSuffix(s, "mb"):
		mult, s = 1<<20, strings.TrimSuffix(s, "mb")
	case strings.HasSuffix(s, "kb"):
		mult, s = 1<<10, strings.TrimSuffix(s, "kb")
	case strings.HasSuffix(s, "b"):
		s = strings.TrimSuffix(s, "b")
	}
	n, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size %q", s)
	}
	return n * mult, nil
}

// DSN builds the postgres connection string.
func (c *Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		c.Database.User, c.Database.Password, c.Database.Host, c.Database.Port, c.Database.Name)
}
