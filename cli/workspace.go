package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// materialize assembles a self-contained level workspace:
//
//	systems/<sys>/shared/  +  systems/<sys>/level-N-*/ overlay
//	+ infrastructure/observability (so dashboards are IDENTICAL everywhere)
//
// This is the same assembly v1's branch generator performed; the output is
// now a disposable local directory instead of a git branch (ADR 0002).
func materialize(root, system string, level int, dst string) error {
	sysDir := filepath.Join(root, "systems", system)
	shared := filepath.Join(sysDir, "shared")
	lvlDir := filepath.Join(sysDir, levelDirs[level])
	if !isDir(lvlDir) {
		return fmt.Errorf("%s does not exist — unknown system or level", lvlDir)
	}

	if err := os.RemoveAll(dst); err != nil {
		return err
	}

	// Shared system content.
	for _, d := range []string{"services", "load-tests", "db"} {
		if err := copyTree(filepath.Join(shared, d), filepath.Join(dst, d)); err != nil {
			return err
		}
	}
	for _, f := range []string{"docker-compose.yml", "README.md"} {
		if err := copyFile(filepath.Join(shared, f), filepath.Join(dst, f)); err != nil {
			return err
		}
	}
	if err := copyFile(filepath.Join(sysDir, "JOURNAL_TEMPLATE.md"), filepath.Join(dst, "JOURNAL_TEMPLATE.md")); err != nil {
		return err
	}

	// Observability — copied verbatim so every level shows the same panels.
	if err := copyTree(filepath.Join(root, "infrastructure", "observability"), filepath.Join(dst, "observability")); err != nil {
		return err
	}

	// system.yaml rendered from the template (for humans; sdl reads state.json).
	if err := renderSystemYAML(filepath.Join(shared, "system.yaml.tmpl"), filepath.Join(dst, "system.yaml"), level); err != nil {
		return err
	}

	// Level overlay: docs to the workspace root, config + pristine baseline.
	entries, err := os.ReadDir(lvlDir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			if err := copyFile(filepath.Join(lvlDir, e.Name()), filepath.Join(dst, e.Name())); err != nil {
				return err
			}
		}
	}
	if err := copyFile(filepath.Join(lvlDir, "config.yaml"), filepath.Join(dst, "config.yaml")); err != nil {
		return err
	}
	if err := copyFile(filepath.Join(lvlDir, "config.yaml"), filepath.Join(dst, ".config.baseline.yaml")); err != nil {
		return err
	}

	// Level-specific surgery.
	switch level {
	case 3:
		if err := copyFile(filepath.Join(lvlDir, "stubs", "ring.go"),
			filepath.Join(dst, "services", "internal", "ring", "ring.go")); err != nil {
			return err
		}
	case 4:
		if err := copyTree(filepath.Join(lvlDir, ".solutions"), filepath.Join(dst, ".solutions")); err != nil {
			return err
		}
	case 5:
		// Stub mains ship as .go.tmpl so the repo's root module doesn't try
		// to resolve their imports; they become real main.go in the workspace.
		for _, svc := range []string{"api-gateway", "shortener", "redirector"} {
			if err := copyFile(filepath.Join(lvlDir, "stubs", svc, "main.go.tmpl"),
				filepath.Join(dst, "services", svc, "main.go")); err != nil {
				return err
			}
		}
		// The ring is part of the scratch build too — reuse the L3 stub.
		if err := copyFile(filepath.Join(sysDir, levelDirs[3], "stubs", "ring.go"),
			filepath.Join(dst, "services", "internal", "ring", "ring.go")); err != nil {
			return err
		}
		if err := copyTree(filepath.Join(lvlDir, "contracts"), filepath.Join(dst, "contracts")); err != nil {
			return err
		}
	}
	return nil
}

func renderSystemYAML(tmplPath, dstPath string, level int) error {
	raw, err := os.ReadFile(tmplPath)
	if err != nil {
		return err
	}
	out := strings.NewReplacer(
		"__LEVEL__", strconv.Itoa(level),
		"__LEVEL_NAME__", levelNames[level],
		"__ESTIMATED_TIME__", levelTimes[level],
		"__GENERATED_FROM__", "workspace",
	).Replace(string(raw))
	return os.WriteFile(dstPath, []byte(out), 0o644)
}

// ensureWorkspace makes workspace/ hold (system, level), parking or
// restoring as needed. Returns true if it changed anything.
func ensureWorkspace(root, system string, level int) (bool, error) {
	ws := workspaceDir(root)
	st := loadState(root)

	if st != nil && st.System == system && st.Level == level && isDir(ws) {
		return false, nil
	}

	// Park the current workspace so user edits survive round-trips.
	if st != nil && isDir(ws) {
		park := parkDir(root, st.Level)
		if err := os.RemoveAll(park); err != nil {
			return false, err
		}
		if err := os.MkdirAll(filepath.Dir(park), 0o755); err != nil {
			return false, err
		}
		if err := os.Rename(ws, park); err != nil {
			return false, err
		}
		fmt.Printf("Parked your Level %d workspace (restored when you come back).\n", st.Level)
	}

	// Restore a previously parked workspace for the target, else materialize.
	park := parkDir(root, level)
	if st != nil && st.System == system && isDir(park) {
		if err := os.Rename(park, ws); err != nil {
			return false, err
		}
		fmt.Printf("Restored your parked Level %d workspace.\n", level)
	} else {
		if err := materialize(root, system, level, ws); err != nil {
			return false, err
		}
		fmt.Printf("Materialized %s Level %d (%s) → workspace/\n", system, level, levelNames[level])
	}

	return true, saveState(root, &State{System: system, Level: level})
}

func cmdLevel(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: sdl level <1-5>")
	}
	level, err := strconv.Atoi(args[0])
	if err != nil || level < 1 || level > 5 {
		return fmt.Errorf("level must be 1-5, got %q", args[0])
	}
	root, err := findRoot()
	if err != nil {
		return err
	}
	system := "url-shortener"
	if st := loadState(root); st != nil {
		system = st.System
	}
	changed, err := ensureWorkspace(root, system, level)
	if err != nil {
		return err
	}
	if !changed {
		fmt.Printf("Already on Level %d.\n", level)
		return nil
	}
	fmt.Printf("\nNext: sdl start   (then read workspace/CONTEXT.md)\n")
	return nil
}

func cmdReset() error {
	root, err := findRoot()
	if err != nil {
		return err
	}
	st, err := requireState(root)
	if err != nil {
		return err
	}
	if err := materialize(root, st.System, st.Level, workspaceDir(root)); err != nil {
		return err
	}
	fmt.Printf("Re-materialized %s Level %d fresh. (Your journal and baseline were untouched.)\n", st.System, st.Level)
	fmt.Println("Apply it to the running stack with: sdl restart")
	return nil
}

// cmdMaterialize is the authoring/CI entrypoint: assemble any level into
// any directory without touching workspace state.
func cmdMaterialize(args []string) error {
	fs := flag.NewFlagSet("materialize", flag.ExitOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	rest := fs.Args()
	if len(rest) != 3 {
		return fmt.Errorf("usage: sdl materialize <system> <level> <dir>")
	}
	level, err := strconv.Atoi(rest[1])
	if err != nil || level < 1 || level > 5 {
		return fmt.Errorf("level must be 1-5")
	}
	root, err := findRoot()
	if err != nil {
		return err
	}
	dst, err := filepath.Abs(rest[2])
	if err != nil {
		return err
	}
	if err := materialize(root, rest[0], level, dst); err != nil {
		return err
	}
	fmt.Printf("materialized %s level %d -> %s\n", rest[0], level, dst)
	return nil
}

// --- file helpers -----------------------------------------------------------

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(p, target)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	fi, err := in.Stat()
	if err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, fi.Mode().Perm())
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
