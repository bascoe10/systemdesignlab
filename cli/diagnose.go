package main

import (
	"bufio"
	_ "embed"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed diagnose/questions.yaml
var questionsYAML []byte

//go:embed diagnose/scoring.yaml
var scoringYAML []byte

type question struct {
	ID          string            `yaml:"id"`
	Category    string            `yaml:"category"`
	Question    string            `yaml:"question"`
	Options     map[string]string `yaml:"options"`
	Correct     string            `yaml:"correct"`
	Explanation string            `yaml:"explanation"`
	MapsToLevel int               `yaml:"maps_to_level"`
}

type questionBank struct {
	Questions []question `yaml:"questions"`
}

type bucket struct {
	QuestionsTagged int `yaml:"questions_tagged"`
	Threshold       int `yaml:"threshold"`
}

type scoringConfig struct {
	Scoring struct {
		LevelBuckets          map[string]bucket `yaml:"level_buckets"`
		DefaultRecommendation string            `yaml:"default_recommendation"`
	} `yaml:"scoring"`
}

var levelPitch = map[int]string{
	2: "Level 2 will build your intuition for config trade-offs — eviction policies, memory limits, pool sizes — with a dashboard showing you every consequence.",
	3: "Level 3 will challenge you to build a consistent hashing ring inside an otherwise-working system.",
	4: "Level 4 drops you into a misconfigured system that you diagnose from dashboards alone.",
	5: "Level 5 hands you contracts and tests — you build the entire system from scratch.",
}

func cmdDiagnose(args []string) error {
	fs := flag.NewFlagSet("diagnose", flag.ExitOnError)
	answersFlag := fs.String("answers", "", "comma-separated answers (a-d) for non-interactive use, e.g. b,c,b,c,b")
	if err := fs.Parse(args); err != nil {
		return err
	}

	var bank questionBank
	if err := yaml.Unmarshal(questionsYAML, &bank); err != nil {
		return fmt.Errorf("parse questions.yaml: %w", err)
	}
	var scoring scoringConfig
	if err := yaml.Unmarshal(scoringYAML, &scoring); err != nil {
		return fmt.Errorf("parse scoring.yaml: %w", err)
	}

	var preset []string
	if *answersFlag != "" {
		preset = strings.Split(*answersFlag, ",")
		if len(preset) != len(bank.Questions) {
			return fmt.Errorf("--answers needs %d answers, got %d", len(bank.Questions), len(preset))
		}
	}

	fmt.Println("SystemDesignLab — Diagnostic Quiz")
	fmt.Println("=================================")
	fmt.Printf("%d questions, multiple choice. No trick questions — answer honestly;\nthe recommendation is only as good as your answers.\n", len(bank.Questions))

	reader := bufio.NewReader(os.Stdin)
	correctByLevel := map[int]int{}
	strengths, gaps := []string{}, []string{}

	for i, q := range bank.Questions {
		fmt.Printf("\n--- Question %d of %d [%s] ---\n\n%s\n", i+1, len(bank.Questions), q.Category, strings.TrimSpace(q.Question))
		for _, key := range sortedKeys(q.Options) {
			fmt.Printf("  %s) %s\n", key, q.Options[key])
		}

		var answer string
		if preset != nil {
			answer = strings.TrimSpace(strings.ToLower(preset[i]))
			fmt.Printf("\nYour answer: %s\n", answer)
		} else {
			for {
				fmt.Print("\nYour answer (a-d): ")
				line, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("read answer: %w", err)
				}
				answer = strings.TrimSpace(strings.ToLower(line))
				if _, ok := q.Options[answer]; ok {
					break
				}
				fmt.Println("Please answer a, b, c, or d.")
			}
		}

		if answer == q.Correct {
			fmt.Println("✓ Correct.")
			correctByLevel[q.MapsToLevel]++
			strengths = append(strengths, q.Category)
		} else {
			fmt.Printf("✗ The answer was %s.\n", q.Correct)
			gaps = append(gaps, q.Category)
		}
		fmt.Printf("\n%s\n", indent(strings.TrimSpace(q.Explanation), "  "))
	}

	// Recommended level = lowest bucket below threshold; all pass → 5.
	recommended := 5
	for _, lvl := range []int{2, 3, 4, 5} {
		b, ok := scoring.Scoring.LevelBuckets[fmt.Sprintf("level-%d", lvl)]
		if !ok {
			continue
		}
		if correctByLevel[lvl] < b.Threshold {
			recommended = lvl
			break
		}
	}

	fmt.Println("\n=================================")
	fmt.Printf("Suggested starting point: Level %d.\n", recommended)
	fmt.Println("(Five questions are a signal, not a verdict — when in doubt, start lower.)")
	fmt.Println()
	if len(strengths) > 0 {
		fmt.Printf("  Solid: %s.\n", strings.Join(dedupe(strengths), ", "))
	}
	if len(gaps) > 0 {
		fmt.Printf("  Gaps:  %s.\n", strings.Join(dedupe(gaps), ", "))
	}
	if pitch, ok := levelPitch[recommended]; ok {
		fmt.Printf("\n  %s\n", pitch)
	}
	fmt.Printf("\n  Run: sdl start --level %d\n", recommended)
	if recommended > 1 {
		fmt.Println("\n  Tip: Even experienced engineers benefit from Level 1. It takes")
		fmt.Println("  ~15 minutes, and its `sdl validate` calibrates the performance")
		fmt.Println("  baseline your machine is judged against at Levels 3-5.")
	}
	return nil
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func indent(s, prefix string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = prefix + l
	}
	return strings.Join(lines, "\n")
}

func dedupe(in []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}
