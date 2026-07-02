// The SystemDesignLab CLI. Users normally reach it through make targets
// (make diagnose, make validate, …) which run `go run ./cli <cmd>` from the
// branch root.
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "diagnose":
		err = cmdDiagnose(os.Args[2:])
	case "validate":
		err = cmdValidate()
	case "journal":
		err = cmdJournal()
	case "reveal-solution":
		err = cmdRevealSolution()
	case "switch":
		err = cmdSwitch(os.Args[2:])
	case "switch-cache":
		err = cmdSwitchCache(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `usage: sdl <command>

commands:
  diagnose         5-question quiz that recommends your entry level
  validate         run level-appropriate checks for the current branch
  journal          create my-journal.md from the template (if missing)
  switch <1-5>     hop to another level; parks/restores uncommitted work
  reveal-solution  unhide SOLUTIONS.md (Level 4)
  switch-cache     switch cache provider in config.yaml (redis|memcached)`)
}
