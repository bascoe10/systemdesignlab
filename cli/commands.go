package main

import (
	"fmt"
	"os"
	"regexp"
)

func cmdRevealSolution() error {
	root, err := findRoot()
	if err != nil {
		return err
	}
	if _, err := requireState(root); err != nil {
		return err
	}
	if err := os.Chdir(root); err != nil {
		return err
	}
	src := "workspace/.solutions/SOLUTIONS.md"
	dst := "workspace/SOLUTIONS.md"
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("no hidden solutions on this level (Level 4 only): %w", err)
	}
	fmt.Println("Last chance: the diagnosis IS the exercise. SOLUTIONS.md is for")
	fmt.Println("checking your answers, not finding them.")
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		return err
	}
	fmt.Printf("\nRevealed %s.\n", dst)
	return nil
}

func cmdCache(args []string) error {
	if len(args) != 1 || (args[0] != "redis" && args[0] != "memcached") {
		return fmt.Errorf("usage: sdl cache <redis|memcached>")
	}
	provider := args[0]
	root, err := findRoot()
	if err != nil {
		return err
	}
	if _, err := requireState(root); err != nil {
		return err
	}
	if err := os.Chdir(root); err != nil {
		return err
	}

	raw, err := os.ReadFile("workspace/config.yaml")
	if err != nil {
		return fmt.Errorf("workspace/config.yaml not found: %w", err)
	}
	// Replace only the provider under the cache: block (first provider key).
	re := regexp.MustCompile(`(?m)^(\s*provider:\s*)(redis|memcached)(.*)$`)
	replaced := false
	out := re.ReplaceAllFunc(raw, func(m []byte) []byte {
		if replaced {
			return m
		}
		replaced = true
		sub := re.FindSubmatch(m)
		return []byte(string(sub[1]) + provider + string(sub[3]))
	})
	if !replaced {
		return fmt.Errorf("could not find cache provider line in workspace/config.yaml")
	}
	if err := os.WriteFile("workspace/config.yaml", out, 0o644); err != nil {
		return err
	}
	fmt.Printf("cache.provider → %s\n", provider)
	fmt.Println("Apply it with: sdl restart && sdl load")
	return nil
}
