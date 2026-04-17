package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/voku/vokuprompt/internal/engine"
)

type placeholderRegistryEntry struct {
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	Examples         []string `json:"examples"`
	Required         bool     `json:"required"`
	Resolution       string   `json:"resolution"`
	ExpectedFormat   string   `json:"expected_format"`
	PreferredSources []string `json:"preferred_sources"`
}

func repoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func TestPlaceholderRegistry_MatchesPatternPlaceholders(t *testing.T) {
	root := repoRoot(t)
	patterns, err := engine.LoadPatterns(filepath.Join(root, "patterns.json"))
	if err != nil {
		t.Fatalf("load patterns.json: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, "placeholders.json"))
	if err != nil {
		t.Fatalf("read placeholders.json: %v", err)
	}

	var registry []placeholderRegistryEntry
	if err := json.Unmarshal(data, &registry); err != nil {
		t.Fatalf("parse placeholders.json: %v", err)
	}

	got := make([]string, 0, len(registry))
	for _, entry := range registry {
		got = append(got, entry.Name)
		if entry.Description == "" {
			t.Fatalf("placeholder %q missing description", entry.Name)
		}
		if len(entry.Examples) == 0 {
			t.Fatalf("placeholder %q missing examples", entry.Name)
		}
		if entry.Resolution == "" {
			t.Fatalf("placeholder %q missing resolution guidance", entry.Name)
		}
		if entry.ExpectedFormat == "" {
			t.Fatalf("placeholder %q missing expected_format", entry.Name)
		}
		if len(entry.PreferredSources) == 0 {
			t.Fatalf("placeholder %q missing preferred_sources", entry.Name)
		}
		if !entry.Required {
			t.Fatalf("placeholder %q should be marked required when emitted", entry.Name)
		}
	}
	slices.Sort(got)

	wantSet := map[string]struct{}{}
	for _, pattern := range patterns {
		for _, placeholder := range pattern.Placeholders {
			wantSet[placeholder] = struct{}{}
		}
	}

	want := make([]string, 0, len(wantSet))
	for name := range wantSet {
		want = append(want, name)
	}
	slices.Sort(want)

	if !slices.Equal(got, want) {
		t.Fatalf("placeholders.json names = %v, want %v", got, want)
	}
}

func TestSkillDoc_DescribesRealWorkflow(t *testing.T) {
	root := repoRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "skills", "vokuprompt", "SKILL.md"))
	if err != nil {
		t.Fatalf("read SKILL.md: %v", err)
	}

	doc := string(data)
	for _, fragment := range []string{
		"vokuprompt categories",
		"vokuprompt optimize --category bugfix",
		"performance",
		"refactor",
		"review",
		"tests",
		"placeholder manifest",
		"placeholders.json",
		"failure analysis",
		"improved prompt",
		"execution result",
		"## Example 1",
		"## Example 2",
	} {
		if !strings.Contains(doc, fragment) {
			t.Fatalf("SKILL.md missing %q", fragment)
		}
	}
}
