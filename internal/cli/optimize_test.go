package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
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

	if resp.Category != "bugfix" {
		t.Errorf("category = %q, want bugfix", resp.Category)
	}

	wantPatterns := []string{"goal", "scope", "verify"}
	if len(resp.SelectedPatterns) != len(wantPatterns) {
		t.Fatalf("selected_patterns = %v, want %v", resp.SelectedPatterns, wantPatterns)
	}
	for i, name := range wantPatterns {
		if resp.SelectedPatterns[i] != name {
			t.Errorf("selected_patterns[%d] = %q, want %q", i, resp.SelectedPatterns[i], name)
		}
	}

	const wantPrompt = "Goal:\n[TASK]\n\nConstraints:\n- Do not modify [SCOPE].\n\nValidation:\n- Run [VALIDATION]."
	if resp.CompiledPrompt != wantPrompt {
		t.Fatalf("compiled_prompt:\ngot:  %q\nwant: %q", resp.CompiledPrompt, wantPrompt)
	}

	if len(resp.PlaceholderManifest) != 3 {
		t.Fatalf("placeholder_manifest len = %d, want 3", len(resp.PlaceholderManifest))
	}
	// Must be sorted: SCOPE, TASK, VALIDATION
	for i, want := range []string{"SCOPE", "TASK", "VALIDATION"} {
		if resp.PlaceholderManifest[i].Name != want {
			t.Errorf("manifest[%d] = %q, want %q", i, resp.PlaceholderManifest[i].Name, want)
		}
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
