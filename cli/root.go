package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Level tables — the single source of truth for level naming.
var levelDirs = map[int]string{
	1: "level-1-observe",
	2: "level-2-experiment",
	3: "level-3-build",
	4: "level-4-fix",
	5: "level-5-scratch",
}

var levelNames = map[int]string{
	1: "Observe & Understand",
	2: "Tweak & Experiment",
	3: "Build the Missing Piece",
	4: "Fix the Broken System",
	5: "Build from Scratch",
}

var levelTimes = map[int]string{
	1: "~1 hour", 2: "2-3 hours", 3: "2-4 hours", 4: "2-4 hours", 5: "6-12 hours",
}

// findRoot walks up from cwd to the repo root (the directory containing
// systems/ and infrastructure/), so sdl works from any subdirectory.
func findRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if isDir(filepath.Join(dir, "systems")) && isDir(filepath.Join(dir, "infrastructure")) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not inside a SystemDesignLab checkout (couldn't find systems/ + infrastructure/)")
		}
		dir = parent
	}
}

func isDir(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.IsDir()
}

// State is what sdl remembers between invocations: which system and level
// the workspace currently holds. Lives in .sdl/state.json at the repo root
// so it survives workspace parking and resets.
type State struct {
	System    string `json:"system"`
	Level     int    `json:"level"`
	UpdatedAt string `json:"updated_at"`
}

func sdlDir(root string) string       { return filepath.Join(root, ".sdl") }
func statePath(root string) string    { return filepath.Join(sdlDir(root), "state.json") }
func workspaceDir(root string) string { return filepath.Join(root, "workspace") }
func parkDir(root string, level int) string {
	return filepath.Join(sdlDir(root), "park", fmt.Sprintf("level-%d", level))
}

func loadState(root string) *State {
	raw, err := os.ReadFile(statePath(root))
	if err != nil {
		return nil
	}
	var s State
	if err := json.Unmarshal(raw, &s); err != nil || s.Level < 1 || s.Level > 5 || s.System == "" {
		return nil
	}
	return &s
}

func saveState(root string, s *State) error {
	s.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := os.MkdirAll(sdlDir(root), 0o755); err != nil {
		return err
	}
	raw, _ := json.MarshalIndent(s, "", "  ")
	return os.WriteFile(statePath(root), append(raw, '\n'), 0o644)
}

// requireState is used by commands that need an existing workspace.
func requireState(root string) (*State, error) {
	s := loadState(root)
	if s == nil || !isDir(workspaceDir(root)) {
		return nil, fmt.Errorf("no active workspace — run `sdl start` (or `sdl start --level N`) first")
	}
	return s, nil
}
