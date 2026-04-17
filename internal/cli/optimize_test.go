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
  "execution_request": "Analyze the original prompt, improve it, and execute the improved prompt now.\n\nReturn:\n1. failure analysis\n2. improved prompt\n3. execution result"
}
`
	if out.String() != want {
		t.Fatalf("optimize JSON contract mismatch\ngot:\n%s\nwant:\n%s", out.String(), want)
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
