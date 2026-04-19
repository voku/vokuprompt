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
		wantPlaceholders []string
	}{
		{
			category:     "performance",
			wantPatterns: []string{"performance_scope", "performance_loop", "performance_validation", "performance_review"},
			wantPromptPieces: []string{
				"dominant workload",
				"realistic benchmarks",
				"benchmark-backed",
			},
			wantPlaceholders: []string{"SCOPE_ELEMENTS", "UNIT", "DONE_CONDITION", "VALIDATION", "STABLE_INTERFACE"},
		},
		{
			category:     "refactor",
			wantPatterns: []string{"refactor_scope", "refactor_loop", "refactor_review"},
			wantPromptPieces: []string{
				"Prefer deletion, simplification, and safe restructuring before extension",
				"contained",
				"deleting duplication before adding new code",
			},
			wantPlaceholders: []string{"SCOPE_ELEMENTS", "UNIT", "DONE_CONDITION", "STABLE_INTERFACE"},
		},
		{
			category:     "review",
			wantPatterns: []string{"review_loop", "review_analysis", "review_challenge"},
			wantPromptPieces: []string{
				"failure analysis",
				"missing",
				"challenged",
			},
			wantPlaceholders: []string{"UNIT", "DONE_CONDITION", "CONTEXT_TARGET", "STABLE_INTERFACE"},
		},
		{
			category:     "tests",
			wantPatterns: []string{"tests_scope", "tests_loop", "tests_validation", "tests_review"},
			wantPromptPieces: []string{
				"failing test",
				"Use [VALIDATION] as proof",
				"Treat tests as proof",
			},
			wantPlaceholders: []string{"SCOPE_ELEMENTS", "UNIT", "DONE_CONDITION", "VALIDATION", "STABLE_INTERFACE"},
		},
		{
			category:     "operational_contract",
			wantPatterns: []string{"ask_before_assume", "multi_pass_workflow", "missingness_detection", "continuation_rule", "verification_prompt", "double_check"},
			wantPromptPieces: []string{
				"Do not invent context",
				"named passes",
				"[PASS_DEFINITIONS]",
				"missing edge cases",
				"Do not stop early",
			},
			wantPlaceholders: []string{"CONTEXT_TARGET", "PASS_DEFINITIONS", "DONE_CONDITION", "VALIDATION", "STABLE_INTERFACE"},
		},
		{
			category:     "code_discovery",
			wantPatterns: []string{"privacy_redaction_gate", "evidence_first_memory", "crystallize_after_work", "memory_capture_gate", "writeback_required"},
			wantPromptPieces: []string{
				"Ground the memory in [EVIDENCE_PATHS]",
				"Crystallize the durable [LEARNING_TYPE] from [CODE_AREA]",
				"Do not treat task completion alone as sufficient",
				"Write or update a [MEMORY_KIND] entry",
			},
			wantPlaceholders: []string{"HANDOFF_SCOPE", "EVIDENCE_PATHS", "DISCOVERY_SCOPE", "LEARNING_TYPE", "CODE_AREA", "MEMORY_KIND", "MEMORY_TARGET_FILE", "FILES_TOUCHED", "CONFIDENCE_LEVEL"},
		},
		{
			category:     "debugging_digest",
			wantPatterns: []string{"privacy_redaction_gate", "evidence_first_memory", "crystallize_after_work", "separate_fact_from_hypothesis", "memory_capture_gate", "writeback_required"},
			wantPromptPieces: []string{
				"Ground the memory in [EVIDENCE_PATHS]",
				"Separate verified facts from hypotheses",
				"[PRIOR_ASSUMPTION]",
				"[UPDATED_ASSUMPTION]",
				"[OPEN_QUESTIONS]",
			},
			wantPlaceholders: []string{"HANDOFF_SCOPE", "EVIDENCE_PATHS", "DISCOVERY_SCOPE", "LEARNING_TYPE", "CODE_AREA", "PRIOR_ASSUMPTION", "UPDATED_ASSUMPTION", "OPEN_QUESTIONS", "MEMORY_KIND", "MEMORY_TARGET_FILE", "FILES_TOUCHED", "CONFIDENCE_LEVEL"},
		},
		{
			category:     "claim_update",
			wantPatterns: []string{"evidence_first_memory", "update_existing_memory_before_creating_new", "separate_fact_from_hypothesis", "memory_capture_gate", "writeback_required"},
			wantPromptPieces: []string{
				"Ground the memory in [EVIDENCE_PATHS]",
				"update that memory before creating a new entry",
				"Separate verified facts from hypotheses",
				"Do not treat task completion alone as sufficient",
			},
			wantPlaceholders: []string{"EVIDENCE_PATHS", "DISCOVERY_SCOPE", "MEMORY_TARGET_FILE", "MEMORY_KIND", "PRIOR_ASSUMPTION", "UPDATED_ASSUMPTION", "OPEN_QUESTIONS", "CODE_AREA", "LEARNING_TYPE", "FILES_TOUCHED", "CONFIDENCE_LEVEL"},
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
		manifestNames := make([]string, 0, len(resp.PlaceholderManifest))
		for _, entry := range resp.PlaceholderManifest {
			manifestNames = append(manifestNames, entry.Name)
		}
		for _, name := range tc.wantPlaceholders {
			if !slices.Contains(manifestNames, name) {
				t.Fatalf("%s placeholder_manifest = %v, want %q", tc.category, manifestNames, name)
			}
		}
		// review: Analysis section must appear before Validation so the prompt reads
		// execute → analyze → validate → challenge, not the reversed form.
		if tc.category == "review" {
			analysisPos := strings.Index(resp.CompiledPrompt, "Analysis:")
			validationPos := strings.Index(resp.CompiledPrompt, "Validation:")
			if analysisPos == -1 || validationPos == -1 || analysisPos >= validationPos {
				t.Fatalf("review: Analysis section must appear before Validation in compiled prompt\n%s", resp.CompiledPrompt)
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

	want := []string{
		"bugfix",
		"performance",
		"refactor",
		"review",
		"tests",
		"operational_contract",
		"code_discovery",
		"implementation_learning",
		"debugging_digest",
		"architecture_memory",
		"claim_update",
		"handoff_memory",
		"memory_review",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("categories = %v, want %v", got, want)
	}
}
