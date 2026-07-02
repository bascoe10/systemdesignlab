package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// cmdSwitch moves between levels without requiring git stash fluency:
//
//	sdl switch 4    (or: make level-4)
//
// Uncommitted work is parked in a stash tagged with the branch it belongs
// to, and restored automatically the next time you switch back to that
// branch. Prefers your my-progress/<system>-level-N branch when it exists.
func cmdSwitch(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: sdl switch <1-5>")
	}
	level, err := strconv.Atoi(args[0])
	if err != nil || level < 1 || level > 5 {
		return fmt.Errorf("level must be 1-5, got %q", args[0])
	}

	system, err := currentSystem()
	if err != nil {
		return err
	}
	if _, err := gitOut("rev-parse", "--git-dir"); err != nil {
		return fmt.Errorf("not a git repository — sdl switch needs the cloned repo")
	}

	current, _ := gitOut("symbolic-ref", "--short", "HEAD")

	generated := fmt.Sprintf("%s/%s", levelDirNames[level], system)
	progress := fmt.Sprintf("my-progress/%s-level-%d", system, level)
	target := generated
	if branchExists(progress) {
		target = progress
	}
	if current == target {
		fmt.Printf("Already on %s.\n", target)
		return nil
	}

	// Park uncommitted work, tagged with the branch it belongs to.
	dirty, _ := gitOut("status", "--porcelain")
	if dirty != "" {
		if _, err := gitOut("stash", "push", "-u", "-m", "sdl:"+current); err != nil {
			return fmt.Errorf("could not stash your uncommitted work: %w", err)
		}
		fmt.Printf("Parked uncommitted work from %s (restored when you switch back).\n", current)
	}

	if _, err := gitOut("checkout", target); err != nil {
		return fmt.Errorf("checkout %s failed — does the branch exist? try `git fetch origin`: %w", target, err)
	}
	fmt.Printf("Switched to %s.\n", target)

	// Restore work previously parked for this branch.
	if ref := findStash("sdl:" + target); ref != "" {
		if _, err := gitOut("stash", "pop", ref); err != nil {
			fmt.Printf("note: could not auto-restore your parked work (%v)\n", err)
			fmt.Printf("      it is safe in the stash: git stash list\n")
		} else {
			fmt.Println("Restored your parked work for this branch.")
		}
	}

	fmt.Println("\nNext: make start   (then read system/CONTEXT.md)")
	return nil
}

var levelDirNames = map[int]string{
	1: "level-1-observe",
	2: "level-2-experiment",
	3: "level-3-build",
	4: "level-4-fix",
	5: "level-5-scratch",
}

func currentSystem() (string, error) {
	raw, err := os.ReadFile("system/system.yaml")
	if err != nil {
		return "", fmt.Errorf("system/system.yaml not found — run from a level branch root: %w", err)
	}
	var meta struct {
		Name string `yaml:"name"`
	}
	if err := yaml.Unmarshal(raw, &meta); err != nil || meta.Name == "" {
		return "", fmt.Errorf("could not read system name from system/system.yaml")
	}
	return meta.Name, nil
}

func branchExists(name string) bool {
	_, err := gitOut("rev-parse", "--verify", "--quiet", "refs/heads/"+name)
	return err == nil
}

// findStash returns the stash ref whose message carries the given tag.
func findStash(tag string) string {
	out, err := gitOut("stash", "list", "--format=%gd\t%gs")
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(out, "\n") {
		ref, msg, ok := strings.Cut(line, "\t")
		if ok && strings.HasSuffix(msg, tag) {
			return ref
		}
	}
	return ""
}

func gitOut(args ...string) (string, error) {
	out, err := exec.Command("git", args...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}
