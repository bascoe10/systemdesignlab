// sdl — the SystemDesignLab CLI, and the product's spine.
//
// It materializes level workspaces from systems/ sources, drives the
// Docker Compose stack, runs load tests and chaos, and validates progress.
// Users never need git beyond `clone`, and never need make or bash.
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
	cmd, args := os.Args[1], os.Args[2:]

	var err error
	switch cmd {
	case "start":
		err = cmdStart(args)
	case "stop":
		err = cmdStop()
	case "restart":
		err = cmdRestart()
	case "clean":
		err = cmdClean()
	case "load":
		err = cmdLoad(args)
	case "dashboard":
		err = cmdDashboard()
	case "level":
		err = cmdLevel(args)
	case "reset":
		err = cmdReset()
	case "status":
		err = cmdStatus()
	case "validate":
		err = cmdValidate()
	case "test":
		err = cmdTest()
	case "journal":
		err = cmdJournal()
	case "diagnose":
		err = cmdDiagnose(args)
	case "reveal-solution":
		err = cmdRevealSolution()
	case "cache":
		err = cmdCache(args)
	case "chaos":
		err = cmdChaos(args)
	case "materialize":
		err = cmdMaterialize(args) // authoring/CI: assemble a level anywhere
	case "help", "--help", "-h":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", cmd)
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `sdl — SystemDesignLab

everyday:
  start [--system S] [--level N]   materialize workspace/ (if needed) and start the stack
  load [--scenario s] [--rate r] [--duration d]
                                   run k6 (steady-state | read-spike | hot-key)
  dashboard                        open Grafana (http://localhost:3000)
  validate                         run level-appropriate checks
  status                           where you are, what's running, what's next
  journal                          create my-journal.md from the template

moving around:
  level <1-5>                      switch levels; your work is parked & restored
  reset                            re-materialize the current level fresh
  diagnose                         5-question quiz → suggested starting level

stack:
  stop | restart | clean           stop / apply config changes / tear down (+volumes)
  test                             go test the workspace services
  cache <redis|memcached>          switch cache provider (then: sdl restart)

breaking things:
  chaos kill-cache [--outage 60]   stop all cache nodes, then restore
  chaos lag-db [--delay 300ms]     add network latency to the database
  chaos overload [--rate 5000] [--duration 2m]
  chaos restore                    undo all chaos

Level 4 only:
  reveal-solution                  unhide SOLUTIONS.md (you sure?)`)
}
