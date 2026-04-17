package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/voku/vokuprompt/internal/output"
)

const testPatterns = `[
  {"name":"goal","role":"framing","prompt":"[TASK]","placeholders":["TASK"],"weights":{"bugfix":1.0}},
  {"name":"scope","role":"constraint","prompt":"Do not modify [SCOPE].","placeholders":["SCOPE"],"weights":{"bugfix":0.8}},
  {"name":"verify","role":"validation","prompt":"Run [VALIDATION].","placeholders":["VALIDATION"],"weights":{"bugfix":0.9}}
]`

const testCategories = `[
  {
    "name":"bugfix",
    "description":"Fix a bug with minimal patch.",
    "role_limits":{"framing":1,"constraint":1,"validation":1},
    "required_patterns":["goal"],
    "optional_patterns":["scope","verify"]
  }
]`

const explainPatterns = `[
  {"name":"goal","role":"framing","prompt":"[TASK]","placeholders":["TASK"]},
  {"name":"exec_fast","role":"execution","prompt":"Do [UNIT] fast.","placeholders":["UNIT"],"weights":{"bugfix":0.9}},
  {"name":"exec_conflict","role":"execution","prompt":"Do [UNIT] another way.","placeholders":["UNIT"],"weights":{"bugfix":0.8},"conflicts":["exec_fast"]},
  {"name":"verify_primary","role":"validation","prompt":"Run [VALIDATION].","placeholders":["VALIDATION"],"weights":{"bugfix":0.7}},
  {"name":"verify_backup","role":"validation","prompt":"Run [VALIDATION] again.","placeholders":["VALIDATION"],"weights":{"bugfix":0.6}}
]`

const explainCategories = `[
  {
    "name":"bugfix",
    "description":"Fix a bug with minimal patch.",
    "role_limits":{"framing":1,"execution":2,"validation":1},
    "required_patterns":["goal"],
    "optional_patterns":["exec_fast","exec_conflict","verify_primary","verify_backup"]
  }
]`

func writeTemp(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0600); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return p
}

func TestExecuteOptimize_RequiresCategory(t *testing.T) {
	var out bytes.Buffer
	err := ExecuteOptimize(nil, &out)
	if err == nil {
		t.Fatal("expected error for missing --category")
	}
}

func TestExecuteOptimize_UnknownCategory(t *testing.T) {
	dir := t.TempDir()
	pp := writeTemp(t, dir, "patterns.json", testPatterns)
	cp := writeTemp(t, dir, "categories.json", testCategories)

	var out bytes.Buffer
	err := ExecuteOptimize([]string{"--category", "nonexistent", "--patterns", pp, "--categories", cp}, &out)
	if err == nil || !strings.Contains(err.Error(), "unknown category") {
		t.Fatalf("expected unknown category error, got: %v", err)
	}
}

func TestExecuteOptimize_ProducesCorrectContract(t *testing.T) {
	dir := t.TempDir()
	pp := writeTemp(t, dir, "patterns.json", testPatterns)
	cp := writeTemp(t, dir, "categories.json", testCategories)

	var out bytes.Buffer
	if err := ExecuteOptimize([]string{"--category", "bugfix", "--patterns", pp, "--categories", cp}, &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp output.OptimizeResponse
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out.String())
	}

	const want = `{
  "category": "bugfix",
  "selected_patterns": [
    "goal",
    "scope",
    "verify"
  ],
  "compiled_prompt": "Goal:\n[TASK]\n\nConstraints:\n- Do not modify [SCOPE].\n\nValidation:\n- Run [VALIDATION].",
  "placeholder_manifest": [
    {
      "name": "SCOPE",
      "required": true,
      "node_types": [
        "Constraints"
      ],
      "source_patterns": [
        "scope"
      ]
    },
    {
      "name": "TASK",
      "required": true,
      "node_types": [
        "Framing"
      ],
      "source_patterns": [
        "goal"
      ]
    },
    {
      "name": "VALIDATION",
      "required": true,
      "node_types": [
        "Validation"
      ],
      "source_patterns": [
        "verify"
      ]
    }
  ],
  "execution_request": "Use the selected category as the deterministic task frame.\n1. Build the final executable prompt by resolving every required placeholder in compiled_prompt from repository facts and the current task context.\n2. Keep the selected category and compiled structure intact; do not silently rewrite the contract.\n3. If a required placeholder cannot be resolved safely, stop and ask for the missing input.\n4. After placeholder resolution, execute the final prompt.\n\nReturn:\n1. category confirmation\n2. placeholder resolution summary\n3. final executable prompt\n4. execution result"
}
`
	if out.String() != want {
		t.Fatalf("optimize JSON contract mismatch\ngot:\n%s\nwant:\n%s", out.String(), want)
	}
}

func TestExecuteOptimize_ExplainIncludesSelectionTrace(t *testing.T) {
	dir := t.TempDir()
	pp := writeTemp(t, dir, "patterns.json", explainPatterns)
	cp := writeTemp(t, dir, "categories.json", explainCategories)

	var out bytes.Buffer
	if err := ExecuteOptimize([]string{"--category", "bugfix", "--explain", "--patterns", pp, "--categories", cp}, &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp output.OptimizeResponse
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out.String())
	}

	if resp.Explanation == nil {
		t.Fatal("expected explanation for --explain")
	}

	// required_patterns: only "goal" was in required_patterns for this category.
	if len(resp.Explanation.RequiredPatterns) != 1 || resp.Explanation.RequiredPatterns[0] != "goal" {
		t.Fatalf("required_patterns = %v, want [goal]", resp.Explanation.RequiredPatterns)
	}

	if len(resp.Explanation.RejectedByConflict) != 1 || resp.Explanation.RejectedByConflict[0].Name != "exec_conflict" {
		t.Fatalf("rejected_by_conflict = %+v", resp.Explanation.RejectedByConflict)
	}
	if len(resp.Explanation.RejectedByConflict[0].ConflictsWith) != 1 || resp.Explanation.RejectedByConflict[0].ConflictsWith[0] != "exec_fast" {
		t.Fatalf("unexpected conflict details: %+v", resp.Explanation.RejectedByConflict[0])
	}
	if len(resp.Explanation.RejectedByRoleLimits) != 1 || resp.Explanation.RejectedByRoleLimits[0].Name != "verify_backup" {
		t.Fatalf("rejected_by_role_limits = %+v", resp.Explanation.RejectedByRoleLimits)
	}
	if resp.Explanation.RejectedByRoleLimits[0].Role != "validation" || resp.Explanation.RejectedByRoleLimits[0].Limit != 1 {
		t.Fatalf("unexpected role limit details: %+v", resp.Explanation.RejectedByRoleLimits[0])
	}
	// blocked_by must name the winner that took the validation slot.
	if resp.Explanation.RejectedByRoleLimits[0].BlockedBy != "verify_primary" {
		t.Fatalf("blocked_by = %q, want \"verify_primary\"", resp.Explanation.RejectedByRoleLimits[0].BlockedBy)
	}

	wantPlaceholders := []string{"TASK", "UNIT", "VALIDATION"}
	if !slices.Equal(resp.Explanation.RequiredPlaceholders, wantPlaceholders) {
		t.Fatalf("required_placeholders = %v, want %v", resp.Explanation.RequiredPlaceholders, wantPlaceholders)
	}
}

func TestExecuteCategories_ListsAll(t *testing.T) {
	dir := t.TempDir()
	cp := writeTemp(t, dir, "categories.json", testCategories)

	var out bytes.Buffer
	if err := ExecuteCategories([]string{"--categories", cp}, &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp struct {
		Categories []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"categories"`
	}
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(resp.Categories) != 1 || resp.Categories[0].Name != "bugfix" {
		t.Fatalf("categories = %+v, want [{bugfix ...}]", resp.Categories)
	}
}

func TestExecuteCategories_JSONSnapshot(t *testing.T) {
	dir := t.TempDir()
	cp := writeTemp(t, dir, "categories.json", testCategories)

	var out bytes.Buffer
	if err := ExecuteCategories([]string{"--categories", cp}, &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	const want = `{
  "categories": [
    {
      "name": "bugfix",
      "description": "Fix a bug with minimal patch."
    }
  ]
}
`
	if out.String() != want {
		t.Fatalf("categories JSON contract mismatch\ngot:\n%s\nwant:\n%s", out.String(), want)
	}
}
