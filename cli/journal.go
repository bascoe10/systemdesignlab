package main

import (
	"fmt"
	"os"
	"strings"
)

const journalPath = "my-journal.md"
const templatePath = "system/JOURNAL_TEMPLATE.md"

func cmdJournal() error {
	if _, err := os.Stat(journalPath); err == nil {
		fmt.Printf("%s already exists — open it in your editor.\n", journalPath)
		return nil
	}
	tmpl, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("journal template not found at %s: %w", templatePath, err)
	}
	if err := os.WriteFile(journalPath, tmpl, 0o644); err != nil {
		return err
	}
	fmt.Printf("Created %s from the template (gitignored — it's yours).\n", journalPath)
	fmt.Println("Fill in constraints and decisions BEFORE you start building.")
	return nil
}

// requiredSections must exist in my-journal.md and differ from the template
// (i.e. actually be filled in). Structural check only — semantic evaluation
// is a Phase 2 concern.
var requiredSections = []string{
	"### Constraints I Identified",
	"### Component Decisions",
	"### What Surprised Me",
	"### Load Test Results",
}

func journalFilled() bool {
	journal, err := os.ReadFile(journalPath)
	if err != nil {
		fmt.Printf("       (no %s — run `make journal` to create it)\n", journalPath)
		return false
	}
	tmpl, _ := os.ReadFile(templatePath)
	for _, section := range requiredSections {
		js := extractSection(string(journal), section)
		if js == "" {
			fmt.Printf("       (journal section %q missing)\n", section)
			return false
		}
		if strings.TrimSpace(js) == strings.TrimSpace(extractSection(string(tmpl), section)) {
			fmt.Printf("       (journal section %q is still the empty template)\n", section)
			return false
		}
	}
	return true
}

func journalHasSLO() bool {
	journal, err := os.ReadFile(journalPath)
	if err != nil {
		return false
	}
	section := extractSection(string(journal), "### SLOs I Would Set")
	tmpl, _ := os.ReadFile(templatePath)
	return section != "" && strings.TrimSpace(section) != strings.TrimSpace(extractSection(string(tmpl), "### SLOs I Would Set"))
}

// extractSection returns the content between a heading (prefix match) and
// the next heading of the same or higher level.
func extractSection(doc, heading string) string {
	lines := strings.Split(doc, "\n")
	var out []string
	in := false
	for _, line := range lines {
		if strings.HasPrefix(line, heading) {
			in = true
			continue
		}
		if in && (strings.HasPrefix(line, "### ") || strings.HasPrefix(line, "## ")) {
			break
		}
		if in {
			out = append(out, line)
		}
	}
	if !in {
		return ""
	}
	return strings.Join(out, "\n")
}
