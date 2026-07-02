package main

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	pass = "[pass]"
	fail = "[fail]"
	warn = "[warn]"
	skip = "[skip]"
)

// Live-metric thresholds. Ratios (hit rate, skew, error ratio) are
// machine-independent → absolute targets. Latency is machine-dependent →
// validated against this machine's own Level 1 baseline when one exists
// (see baseline.go), with a generous absolute fallback otherwise.
const (
	minHitRate         = 0.80
	fallbackP99Seconds = 0.150
	p99BaselineFactor  = 1.5   // allowed p99 = factor × your healthy baseline
	p99FloorSeconds    = 0.050 // never demand tighter than this
	maxNodeSkew        = 0.25  // relative stddev of per-node cache ops
	trafficQuery       = `sum(increase(http_requests_total{route!=""}[2h]))`
	// Numerators are wrapped in `or vector(0)`: sum() over an empty vector
	// returns no data, and "no 5xx at all" must read as 0%, not an error.
	hitRateQuery  = `(sum(rate(cache_requests_total{op="get",result="hit"}[10m])) or vector(0)) / clamp_min(sum(rate(cache_requests_total{op="get",result=~"hit|miss|bypass"}[10m])), 0.001)`
	p99Query      = `histogram_quantile(0.99, sum by (le) (rate(http_request_duration_seconds_bucket{job="api-gateway"}[10m])))`
	p50Query      = `histogram_quantile(0.50, sum by (le) (rate(http_request_duration_seconds_bucket{job="api-gateway"}[10m])))`
	nodeSkewQuery = `stddev(sum by (node) (increase(cache_node_requests_total[10m]))) / clamp_min(avg(sum by (node) (increase(cache_node_requests_total[10m]))), 1)`
	errRatioQuery = `(sum(rate(http_requests_total{job="api-gateway",code=~"5.."}[10m])) or vector(0)) / clamp_min(sum(rate(http_requests_total{job="api-gateway"}[10m])), 0.001)`
)

func cmdValidate() error {
	root, err := findRoot()
	if err != nil {
		return err
	}
	st, err := requireState(root)
	if err != nil {
		return err
	}
	if err := os.Chdir(root); err != nil {
		return err
	}

	fmt.Printf("Validating %s — Level %d (%s)\n\n", st.System, st.Level, levelNames[st.Level])

	failed := false
	switch st.Level {
	case 1:
		failed = !validateLevel1()
	case 2:
		failed = !validateLevel2()
	case 3:
		failed = !validateLevel3()
	case 4:
		failed = !validateLevel4()
	case 5:
		failed = !validateLevel5()
	}

	fmt.Println()
	if failed {
		fmt.Println("Validation did not pass yet. Keep going — the dashboards know why.")
		os.Exit(1)
	}
	fmt.Println("Validation passed. Nice work.")
	return nil
}

func check(ok bool, tag, msg string) bool {
	if ok {
		fmt.Printf("%s %s\n", pass, msg)
	} else {
		fmt.Printf("%s %s\n", tag, msg)
	}
	return ok
}

// --- Level 1: watch the system under load, then calibrate ------------------
// Level 1 doubles as calibration: it records THIS machine's healthy numbers
// into .sdl/baseline.json, and Levels 3-5 measure latency against them.

func validateLevel1() bool {
	if !promReachable() {
		fmt.Printf("%s Prometheus not reachable at localhost:9090 — run `sdl start` first\n", fail)
		return false
	}
	total, err := promQuery(trafficQuery)
	ok := check(err == nil && total > 1000, fail,
		fmt.Sprintf("load test traffic observed in the last 2h (want >1000 requests, saw %.0f)", total))

	b, err := captureBaseline()
	if err != nil {
		fmt.Printf("%s baseline calibration: %v\n", fail, err)
		fmt.Println("       (run `sdl load` to steady state, then `sdl validate` again)")
		ok = false
	} else {
		fmt.Printf("%s baseline calibrated for this machine → %s\n", pass, baselinePath)
		fmt.Printf("       p50 %.1fms · p99 %.0fms · hit rate %.0f%% · node skew %.0f%%\n",
			b.P50Seconds*1000, b.P99Seconds*1000, b.HitRate*100, b.NodeSkew*100)
		fmt.Println("       These are YOUR healthy numbers. Levels 3-5 validate against them.")
	}

	fmt.Println("\nNow answer the prompts in workspace/QUESTIONS.md — in writing.")
	fmt.Println("When you can explain the p50/p99 gap and the hit-rate warmup curve,")
	fmt.Println("you're ready for: sdl level 2")
	return ok
}

// --- Level 2: did you experiment? ------------------------------------------

func validateLevel2() bool {
	current, err1 := os.ReadFile("workspace/config.yaml")
	baseline, err2 := os.ReadFile("workspace/.config.baseline.yaml")
	if err1 != nil || err2 != nil {
		fmt.Printf("%s workspace/config.yaml or its pristine copy missing (sdl reset to repair)\n", fail)
		return false
	}
	changed := string(current) != string(baseline)
	ok := check(changed, fail, "config.yaml modified from the shipped baseline (run at least one experiment)")
	if !changed {
		fmt.Println("\nPick an experiment from workspace/EXPERIMENTS.md, edit workspace/config.yaml,")
		fmt.Println("then: sdl restart && sdl load")
	} else {
		fmt.Println("\nCompare what you predicted against what the dashboards showed.")
		fmt.Println("Restore the baseline when done: sdl reset")
	}
	return ok
}

// --- Level 3: tiered — tests, live metrics, journal ------------------------

func validateLevel3() bool {
	ok := true
	ok = check(runGoTest("./internal/ring/"), fail, "Required   ring unit tests (go test ./internal/ring/)") && ok
	ok = livePerformanceChecks(false) && ok
	ok = check(journalFilled(), fail, "Journal    my-journal.md has the required sections filled in") && ok
	return ok
}

// --- Level 4: golden signals back to YOUR baseline + journal ---------------

func validateLevel4() bool {
	if !promReachable() {
		fmt.Printf("%s Prometheus not reachable — run `sdl start && sdl load` first\n", fail)
		return false
	}
	if loadBaseline() != nil {
		fmt.Println("Target: return the system to YOUR Level 1 baseline.")
	}
	ok := livePerformanceChecks(true)
	ok = check(journalFilled(), fail, "Journal    symptom → hypothesis → evidence → fix, per issue") && ok
	return ok
}

// --- Level 5: full suite ----------------------------------------------------

func validateLevel5() bool {
	ok := true
	ok = check(runGoTest("./internal/ring/"), fail, "Required   ring unit tests") && ok
	ok = check(runGoTestTags("integration", "./integration/"), fail, "Required   integration tests against the running stack") && ok
	ok = livePerformanceChecks(false) && ok
	ok = check(journalFilled(), fail, "Journal    architecture decisions + load test numbers") && ok
	ok = check(journalHasSLO(), fail, "Journal    SLO definitions with reasoning (see SLI/SLO dashboard)") && ok
	return ok
}

// livePerformanceChecks verifies the running system. If required is false
// and Prometheus is down, checks are skipped with a warning, not failed.
func livePerformanceChecks(required bool) bool {
	if !promReachable() {
		tag := warn
		if required {
			tag = fail
		}
		fmt.Printf("%s Performance  Prometheus not reachable — start the stack and run a load test\n", tag)
		return !required
	}
	ok := true

	// Latency threshold: relative to this machine's Level 1 baseline when
	// available; generic fallback (with a warning) when not.
	p99Limit := fallbackP99Seconds
	p99Label := fmt.Sprintf("generic ≤ %.0fms", fallbackP99Seconds*1000)
	if b := loadBaseline(); b != nil {
		p99Limit = math.Max(p99BaselineFactor*b.P99Seconds, p99FloorSeconds)
		p99Label = fmt.Sprintf("≤ %.0fms = %.1f× your L1 baseline", p99Limit*1000, p99BaselineFactor)
	} else {
		fmt.Printf("%s Performance  no %s — using generic latency bounds; run Level 1's\n", warn, baselinePath)
		fmt.Println("             `sdl validate` once to calibrate for this machine")
	}

	hit, err := promQuery(hitRateQuery)
	ok = check(err == nil && hit >= minHitRate, fail,
		fmt.Sprintf("Performance  cache hit rate over last 10m: %.0f%% (want ≥ %.0f%%)", hit*100, minHitRate*100)) && ok

	p99, err := promQuery(p99Query)
	ok = check(err == nil && p99 <= p99Limit, fail,
		fmt.Sprintf("Performance  gateway p99 over last 10m: %.0fms (%s)", p99*1000, p99Label)) && ok

	skew, err := promQuery(nodeSkewQuery)
	ok = check(err == nil && skew <= maxNodeSkew, fail,
		fmt.Sprintf("Performance  cache node distribution rel. stddev: %.0f%% (want ≤ %.0f%%)", skew*100, maxNodeSkew*100)) && ok

	errRatio, err := promQuery(errRatioQuery)
	ok = check(err == nil && errRatio <= 0.01, fail,
		fmt.Sprintf("Performance  5xx ratio over last 10m: %.2f%% (want ≤ 1%%)", errRatio*100)) && ok

	if !ok {
		fmt.Printf("%s          (metrics need a recent load test: sdl load)\n", skip)
	}
	return ok
}

func runGoTest(pkg string) bool {
	return goTest(pkg)
}

func runGoTestTags(tags, pkg string) bool {
	return goTest(pkg, "-tags", tags, "-count=1")
}

func goTest(pkg string, extra ...string) bool {
	args := append([]string{"test"}, extra...)
	args = append(args, pkg)
	cmd := exec.Command("go", args...)
	cmd.Dir = filepath.Join("workspace", "services")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(strings.TrimSpace(string(out)))
	}
	return err == nil
}
