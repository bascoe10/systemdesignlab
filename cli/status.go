package main

import (
	"fmt"
	"os"
	"strings"
)

// cmdStatus answers "where am I and what's next" — the narrative glue
// between sessions.
func cmdStatus() error {
	root, err := findRoot()
	if err != nil {
		return err
	}
	st := loadState(root)
	if st == nil || !isDir(workspaceDir(root)) {
		fmt.Println("No active workspace yet.")
		fmt.Println("\n  sdl diagnose        not sure where to start?")
		fmt.Println("  sdl start           begin at Level 1 (url-shortener)")
		return nil
	}
	if err := os.Chdir(root); err != nil {
		return err
	}

	fmt.Printf("System   %s\n", st.System)
	fmt.Printf("Level    %d — %s (est. %s)\n", st.Level, levelNames[st.Level], levelTimes[st.Level])

	running, _ := composeOut(root, bothProfiles, "ps", "--services", "--status", "running")
	n := 0
	if running != "" {
		n = len(strings.Fields(running))
	}
	if n > 0 {
		fmt.Printf("Stack    running (%d services) — Grafana http://localhost:3000\n", n)
	} else {
		fmt.Printf("Stack    stopped — sdl start\n")
	}

	if b := loadBaseline(); b != nil {
		fmt.Printf("Baseline calibrated %s (p99 %.0fms, hit %.0f%%)\n", b.CapturedAt, b.P99Seconds*1000, b.HitRate*100)
	} else {
		fmt.Printf("Baseline not captured — Level 1's `sdl validate` records your healthy numbers\n")
	}

	if _, err := os.Stat(journalPath); err == nil {
		fmt.Printf("Journal  my-journal.md exists\n")
	} else if st.Level >= 3 {
		fmt.Printf("Journal  missing — required at this level: sdl journal\n")
	}

	fmt.Printf("\nRead workspace/CONTEXT.md for where you are; sdl validate when you think you're done.\n")
	return nil
}
