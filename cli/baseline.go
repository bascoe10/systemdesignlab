package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Baseline is this machine's healthy Level 1 snapshot. Ratio metrics (hit
// rate, skew, error ratio) are machine-independent and keep absolute
// thresholds; latency is machine-dependent, so Levels 3-5 validate against
// a multiple of YOUR healthy p99 rather than a number tuned on someone
// else's laptop. Captured by `sdl validate` on Level 1; lives in .sdl/ so
// it survives workspace parking, level switches, and resets.
type Baseline struct {
	CapturedAt string  `json:"captured_at"`
	P50Seconds float64 `json:"p50_seconds"`
	P99Seconds float64 `json:"p99_seconds"`
	HitRate    float64 `json:"hit_rate"`
	NodeSkew   float64 `json:"node_skew"`
}

const baselinePath = ".sdl/baseline.json"

func loadBaseline() *Baseline {
	raw, err := os.ReadFile(baselinePath)
	if err != nil {
		return nil
	}
	var b Baseline
	if err := json.Unmarshal(raw, &b); err != nil || b.P99Seconds <= 0 {
		return nil
	}
	return &b
}

// captureBaseline snapshots current live metrics. It refuses to calibrate
// against an unhealthy or cold system — a bad baseline is worse than none.
func captureBaseline() (*Baseline, error) {
	hit, err := promQuery(hitRateQuery)
	if err != nil {
		return nil, fmt.Errorf("hit rate: %w", err)
	}
	if hit < 0.80 {
		return nil, fmt.Errorf("cache hit rate is %.0f%% — let the load test reach steady state (>80%%) first", hit*100)
	}
	p99, err := promQuery(p99Query)
	if err != nil || p99 <= 0 {
		return nil, fmt.Errorf("p99 unavailable: %v", err)
	}
	p50, err := promQuery(p50Query)
	if err != nil {
		return nil, fmt.Errorf("p50 unavailable: %w", err)
	}
	skew, err := promQuery(nodeSkewQuery)
	if err != nil {
		return nil, fmt.Errorf("node skew unavailable: %w", err)
	}

	b := &Baseline{
		CapturedAt: time.Now().UTC().Format(time.RFC3339),
		P50Seconds: p50,
		P99Seconds: p99,
		HitRate:    hit,
		NodeSkew:   skew,
	}
	if err := os.MkdirAll(".sdl", 0o755); err != nil {
		return nil, err
	}
	raw, _ := json.MarshalIndent(b, "", "  ")
	if err := os.WriteFile(baselinePath, append(raw, '\n'), 0o644); err != nil {
		return nil, err
	}
	return b, nil
}
