package cli

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/voku/vokuprompt/internal/output"
)

func TestRealCategories_ProduceDistinctCompiledPrompts(t *testing.T) {
	root := repoRoot(t)
	patternsPath := filepath.Join(root, "patterns.json")
	categoriesPath := filepath.Join(root, "categories.json")

	cases := []struct {
		category         string
		wantPatterns     []string
		wantPromptPieces []string
	}{
		{
			category:     "performance",
			wantPatterns: []string{"performance_scope", "performance_loop", "performance_validation", "performance_review"},
			wantPromptPieces: []string{
				"dominant workload",
				"realistic benchmarks",
				"benchmark-backed",
			},
		},
		{
			category:     "refactor",
			wantPatterns: []string{"refactor_scope", "refactor_loop", "refactor_review"},
			wantPromptPieces: []string{
				"Prefer deletion, simplification, and safe restructuring before extension",
				"contained",
				"deleting duplication before adding new code",
			},
		},
		{
			category:     "review",
			wantPatterns: []string{"review_loop", "review_analysis", "review_challenge"},
			wantPromptPieces: []string{
				"failure analysis",
				"missing",
				"challenged",
			},
		},
		{
			category:     "tests",
			wantPatterns: []string{"tests_scope", "tests_loop", "tests_validation", "tests_review"},
			wantPromptPieces: []string{
				"failing test",
				"Use [VALIDATION] as proof",
				"Treat tests as proof",
			},
		},
	}

	prompts := make(map[string]string, len(cases))
	for _, tc := range cases {
		var out bytes.Buffer
		if err := ExecuteOptimize([]string{"--category", tc.category, "--patterns", patternsPath, "--categories", categoriesPath}, &out); err != nil {
			t.Fatalf("ExecuteOptimize(%s): %v", tc.category, err)
		}

		var resp output.OptimizeResponse
		if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
			t.Fatalf("parse %s optimize output: %v", tc.category, err)
		}

		prompts[tc.category] = resp.CompiledPrompt
		for _, name := range tc.wantPatterns {
			if !slices.Contains(resp.SelectedPatterns, name) {
				t.Fatalf("%s selected_patterns = %v, want %q", tc.category, resp.SelectedPatterns, name)
			}
		}
		for _, fragment := range tc.wantPromptPieces {
			if !strings.Contains(resp.CompiledPrompt, fragment) {
				t.Fatalf("%s compiled_prompt missing %q\n%s", tc.category, fragment, resp.CompiledPrompt)
			}
		}
	}

	seen := map[string]string{}
	for category, prompt := range prompts {
		if existing, ok := seen[prompt]; ok {
			t.Fatalf("compiled_prompt for %s unexpectedly matches %s", category, existing)
		}
		seen[prompt] = category
	}
}

func TestExecuteCategories_RealConfigListsNewCategories(t *testing.T) {
	root := repoRoot(t)

	var out bytes.Buffer
	if err := ExecuteCategories([]string{"--categories", filepath.Join(root, "categories.json")}, &out); err != nil {
		t.Fatalf("ExecuteCategories: %v", err)
	}

	var resp struct {
		Categories []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"categories"`
	}
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("parse categories output: %v", err)
	}

	got := make([]string, 0, len(resp.Categories))
	for _, category := range resp.Categories {
		if category.Description == "" {
			t.Fatalf("category %q missing description", category.Name)
		}
		got = append(got, category.Name)
	}

	want := []string{"bugfix", "performance", "refactor", "review", "tests"}
	if !slices.Equal(got, want) {
		t.Fatalf("categories = %v, want %v", got, want)
	}
}
