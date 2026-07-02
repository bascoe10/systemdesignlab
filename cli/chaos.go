package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func cmdChaos(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: sdl chaos <kill-cache|lag-db|overload|restore>")
	}
	root, err := findRoot()
	if err != nil {
		return err
	}
	if _, err := requireState(root); err != nil {
		return err
	}

	switch args[0] {
	case "kill-cache":
		return chaosKillCache(root, args[1:])
	case "lag-db":
		return chaosLagDB(root, args[1:])
	case "overload":
		return chaosOverload(root, args[1:])
	case "restore":
		return chaosRestore(root)
	default:
		return fmt.Errorf("unknown chaos %q (kill-cache | lag-db | overload | restore)", args[0])
	}
}

func banner(what string) {
	fmt.Printf("\n=== CHAOS: %s ===\n", what)
	fmt.Println("Watch it happen: Grafana → Chaos Impact dashboard (http://localhost:3000)")
	fmt.Println()
}

// cacheServices lists the cache node services of whichever provider exists.
func cacheServices(root string) ([]string, error) {
	out, err := composeOut(root, bothProfiles, "ps", "--services", "--all")
	if err != nil {
		return nil, err
	}
	var svcs []string
	for _, s := range strings.Fields(out) {
		if strings.HasPrefix(s, "redis-") || strings.HasPrefix(s, "memcached-") {
			svcs = append(svcs, s)
		}
	}
	if len(svcs) == 0 {
		return nil, fmt.Errorf("no cache containers found — is the stack running? (sdl start)")
	}
	return svcs, nil
}

func chaosKillCache(root string, args []string) error {
	fs := flag.NewFlagSet("kill-cache", flag.ExitOnError)
	outage := fs.Int("outage", 60, "seconds before restore")
	if err := fs.Parse(args); err != nil {
		return err
	}
	svcs, err := cacheServices(root)
	if err != nil {
		return err
	}
	banner(fmt.Sprintf("killing all cache nodes for %ds", *outage))
	if err := compose(root, bothProfiles, append([]string{"stop"}, svcs...)...); err != nil {
		return err
	}
	fmt.Println("Cache is DOWN. Questions while you wait:")
	fmt.Println("  - Did the error rate move, or only latency? Why?")
	fmt.Println("  - Where is the read traffic going now? (Deep Dive → DB queries/sec)")
	for left := *outage; left > 0; left -= 10 {
		fmt.Printf("  restoring in %ds...\n", left)
		wait := 10
		if left < 10 {
			wait = left
		}
		time.Sleep(time.Duration(wait) * time.Second)
	}
	if err := compose(root, bothProfiles, append([]string{"start"}, svcs...)...); err != nil {
		return err
	}
	fmt.Println("\nCache restored. Watch the recovery: how long until the hit rate is")
	fmt.Println("back above 85%? See docs/resilience-challenges/KILL_CACHE.md")
	return nil
}

func chaosLagDB(root string, args []string) error {
	fs := flag.NewFlagSet("lag-db", flag.ExitOnError)
	delay := fs.String("delay", "300ms", "injected latency")
	if err := fs.Parse(args); err != nil {
		return err
	}
	banner(fmt.Sprintf("adding %s latency to the database", *delay))
	dbID, err := composeOut(root, bothProfiles, "ps", "-q", "db")
	if err != nil || dbID == "" {
		return fmt.Errorf("database container not running — sdl start first")
	}
	// tc netem, run from a sidecar in the DB container's network namespace.
	cmd := exec.Command("docker", "run", "--rm", "--network", "container:"+dbID,
		"--cap-add", "NET_ADMIN", "nicolaka/netshoot",
		"tc", "qdisc", "add", "dev", "eth0", "root", "netem", "delay", *delay)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, `
Could not inject latency. tc netem needs the Linux kernel netem module;
on Docker Desktop (Mac/Windows) this may be unavailable.

Manual fallback: set database.pool_size: 1 in workspace/config.yaml and
re-run the load test — queueing produces a similar tail-latency signature.`)
		return err
	}
	fmt.Printf("\nDone. Every DB round-trip now costs an extra %s.\n", *delay)
	fmt.Println("  - Which requests feel it: cache hits or misses? Check p50 vs p99.")
	fmt.Println("  - How well does the cache absorb it? Compare hit rate to latency.")
	fmt.Println("Restore with: sdl chaos restore")
	return nil
}

func chaosOverload(root string, args []string) error {
	fs := flag.NewFlagSet("overload", flag.ExitOnError)
	rate := fs.String("rate", "5000", "requests/sec (~5x steady state)")
	duration := fs.String("duration", "2m", "how long")
	if err := fs.Parse(args); err != nil {
		return err
	}
	banner(fmt.Sprintf("overload — %s rps for %s", *rate, *duration))
	fmt.Println("While it runs, keep these panels open:")
	fmt.Println("  - Golden Signals: which signal degrades FIRST?")
	fmt.Println("  - Saturation panels: which resource is the ceiling?")
	fmt.Println()
	if err := compose(root, cacheProvider(root), "run", "--rm",
		"-e", "RATE="+*rate, "-e", "DURATION="+*duration,
		"k6", "run", "/scripts/steady-state.js"); err != nil {
		return err
	}
	fmt.Println("\nPost-mortem prompts (journal them):")
	fmt.Println("  - What was the breaking point in rps?")
	fmt.Println("  - What would you scale first, and what evidence says so?")
	return nil
}

func chaosRestore(root string) error {
	banner("restoring healthy state")
	if dbID, err := composeOut(root, bothProfiles, "ps", "-q", "db"); err == nil && dbID != "" {
		if exec.Command("docker", "run", "--rm", "--network", "container:"+dbID,
			"--cap-add", "NET_ADMIN", "nicolaka/netshoot",
			"tc", "qdisc", "del", "dev", "eth0", "root").Run() == nil {
			fmt.Println("removed DB network latency")
		} else {
			fmt.Println("no injected DB latency found (fine)")
		}
	}
	if svcs, err := cacheServices(root); err == nil {
		if err := compose(root, bothProfiles, append([]string{"start"}, svcs...)...); err != nil {
			return err
		}
	}
	if err := compose(root, cacheProvider(root), "up", "-d"); err != nil {
		return err
	}
	fmt.Println("\nAll services running. Verify recovery on the Chaos Impact dashboard:")
	fmt.Println("how long does each signal take to return to baseline?")
	return nil
}
