package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecuteCategories_UsesBundleAssetNextToExecutable(t *testing.T) {
	dir := t.TempDir()
	bundleDir := filepath.Join(dir, "bundle")
	if err := os.MkdirAll(bundleDir, 0o755); err != nil {
		t.Fatalf("mkdir bundle dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(bundleDir, "categories.json"), []byte(`[
		{"name":"bugfix","description":"Fix a bug"}
	]`), 0o644); err != nil {
		t.Fatalf("write categories.json: %v", err)
	}

	restoreExecutable := stubExecutablePath(t, filepath.Join(bundleDir, "vokuprompt"))
	defer restoreExecutable()

	restoreCwd := chdir(t, dir)
	defer restoreCwd()

	var out bytes.Buffer
	if err := ExecuteCategories(nil, &out); err != nil {
		t.Fatalf("ExecuteCategories: %v", err)
	}

	if got := out.String(); !strings.Contains(got, `"name": "bugfix"`) {
		t.Fatalf("ExecuteCategories output = %s, want bugfix category", got)
	}
}

func TestExecuteOptimize_UsesBundleAssetsNextToExecutable(t *testing.T) {
	dir := t.TempDir()
	bundleDir := filepath.Join(dir, "bundle")
	if err := os.MkdirAll(bundleDir, 0o755); err != nil {
		t.Fatalf("mkdir bundle dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(bundleDir, "patterns.json"), []byte(`[
		{
			"name":"goal_context_constraints_done",
			"role":"framing",
			"weights":{"bugfix":100},
			"conflicts":[],
			"prompt":"[TASK]\n\nContext:\n[CONTEXT_TARGET]\n\nConstraints:\n- Do not modify unrelated [SCOPE_ELEMENTS].\n\nExecution:\n- Work step by step over [UNIT] and continue until [DONE_CONDITION].",
			"placeholders":["TASK","CONTEXT_TARGET","SCOPE_ELEMENTS","UNIT","DONE_CONDITION"]
		}
	]`), 0o644); err != nil {
		t.Fatalf("write patterns.json: %v", err)
	}

	if err := os.WriteFile(filepath.Join(bundleDir, "categories.json"), []byte(`[
		{"name":"bugfix","description":"Fix a bug","role_limits":{},"required_patterns":["goal_context_constraints_done"],"optional_patterns":[]}
	]`), 0o644); err != nil {
		t.Fatalf("write categories.json: %v", err)
	}

	restoreExecutable := stubExecutablePath(t, filepath.Join(bundleDir, "vokuprompt"))
	defer restoreExecutable()

	restoreCwd := chdir(t, dir)
	defer restoreCwd()

	var out bytes.Buffer
	if err := ExecuteOptimize([]string{"--category", "bugfix"}, &out); err != nil {
		t.Fatalf("ExecuteOptimize: %v", err)
	}

	if got := out.String(); !strings.Contains(got, `"category": "bugfix"`) {
		t.Fatalf("ExecuteOptimize output = %s, want bugfix category", got)
	}
}

func stubExecutablePath(t *testing.T, path string) func() {
	t.Helper()

	previous := executablePath
	executablePath = func() (string, error) {
		return path, nil
	}

	return func() {
		executablePath = previous
	}
}

func chdir(t *testing.T, dir string) func() {
	t.Helper()

	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir to %s: %v", dir, err)
	}

	return func() {
		if err := os.Chdir(previous); err != nil {
			t.Fatalf("restore cwd to %s: %v", previous, err)
		}
	}
}
