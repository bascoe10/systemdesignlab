package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// cacheProvider reads cache.provider from the workspace config — it selects
// the Compose profile, so redis and memcached node sets swap cleanly while
// both expose the same cache-1..3 network aliases.
func cacheProvider(root string) string {
	raw, err := os.ReadFile(filepath.Join(workspaceDir(root), "config.yaml"))
	if err != nil {
		return "redis"
	}
	var cfg struct {
		Cache struct {
			Provider string `yaml:"provider"`
		} `yaml:"cache"`
	}
	if yaml.Unmarshal(raw, &cfg) != nil || cfg.Cache.Provider == "" {
		return "redis"
	}
	return cfg.Cache.Provider
}

// compose runs docker compose against the workspace with the given profiles.
func compose(root, profiles string, args ...string) error {
	full := append([]string{"compose", "-f", filepath.Join(workspaceDir(root), "docker-compose.yml")}, args...)
	cmd := exec.Command("docker", full...)
	cmd.Env = append(os.Environ(), "COMPOSE_PROFILES="+profiles)
	cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
	return cmd.Run()
}

func composeOut(root, profiles string, args ...string) (string, error) {
	full := append([]string{"compose", "-f", filepath.Join(workspaceDir(root), "docker-compose.yml")}, args...)
	cmd := exec.Command("docker", full...)
	cmd.Env = append(os.Environ(), "COMPOSE_PROFILES="+profiles)
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

// bothProfiles is used for teardown: after a provider switch, the OLD
// profile's containers must go too (they hold the cache-1..3 aliases).
const bothProfiles = "redis,memcached"

func cmdStart(args []string) error {
	fs := flag.NewFlagSet("start", flag.ExitOnError)
	system := fs.String("system", "", "system to run (default: current, or url-shortener)")
	level := fs.Int("level", 0, "level to run (default: current, or 1)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, err := findRoot()
	if err != nil {
		return err
	}

	// Defaults: continue where you left off; first run = url-shortener L1.
	st := loadState(root)
	sys, lvl := "url-shortener", 1
	if st != nil {
		sys, lvl = st.System, st.Level
	}
	if *system != "" {
		sys = *system
	}
	if *level != 0 {
		lvl = *level
	}
	if lvl < 1 || lvl > 5 {
		return fmt.Errorf("--level must be 1-5")
	}

	if _, err := ensureWorkspace(root, sys, lvl); err != nil {
		return err
	}

	if err := compose(root, cacheProvider(root), "up", "-d", "--build"); err != nil {
		return err
	}
	fmt.Println()
	fmt.Println("  Gateway     http://localhost:8080")
	fmt.Println("  Grafana     http://localhost:3000   (sdl dashboard)")
	fmt.Println("  Prometheus  http://localhost:9090")
	fmt.Println()
	fmt.Println("  Next: sdl load    then read workspace/CONTEXT.md")
	return nil
}

func cmdStop() error {
	root, err := findRoot()
	if err != nil {
		return err
	}
	if _, err := requireState(root); err != nil {
		return err
	}
	return compose(root, bothProfiles, "stop")
}

func cmdRestart() error {
	root, err := findRoot()
	if err != nil {
		return err
	}
	if _, err := requireState(root); err != nil {
		return err
	}
	if err := compose(root, bothProfiles, "down", "--remove-orphans"); err != nil {
		return err
	}
	return compose(root, cacheProvider(root), "up", "-d", "--build")
}

func cmdClean() error {
	root, err := findRoot()
	if err != nil {
		return err
	}
	if _, err := requireState(root); err != nil {
		return err
	}
	return compose(root, bothProfiles, "down", "-v", "--remove-orphans")
}

func cmdLoad(args []string) error {
	fs := flag.NewFlagSet("load", flag.ExitOnError)
	scenario := fs.String("scenario", "steady-state", "steady-state | read-spike | hot-key")
	rate := fs.String("rate", "", "requests/sec (scenario default if empty)")
	duration := fs.String("duration", "", "e.g. 45s, 2m (scenario default if empty)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, err := findRoot()
	if err != nil {
		return err
	}
	if _, err := requireState(root); err != nil {
		return err
	}
	return compose(root, cacheProvider(root), "run", "--rm",
		"-e", "RATE="+*rate, "-e", "DURATION="+*duration,
		"k6", "run", "/scripts/"+*scenario+".js")
}

func cmdDashboard() error {
	url := "http://localhost:3000"
	for _, opener := range [][]string{{"xdg-open", url}, {"open", url}} {
		if exec.Command(opener[0], opener[1:]...).Run() == nil {
			return nil
		}
	}
	fmt.Printf("Open %s in your browser\n", url)
	return nil
}

func cmdTest() error {
	root, err := findRoot()
	if err != nil {
		return err
	}
	if _, err := requireState(root); err != nil {
		return err
	}
	cmd := exec.Command("go", "test", "./internal/...")
	cmd.Dir = filepath.Join(workspaceDir(root), "services")
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}
