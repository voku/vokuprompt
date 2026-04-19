package main

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRun_UsageWhenNoCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run(nil, &stdout, &stderr)
	if exitCode != 1 {
		t.Fatalf("run(nil) exit code = %d, want 1", exitCode)
	}
	if got := stderr.String(); !strings.Contains(got, "usage: vokuprompt <categories|optimize>") {
		t.Fatalf("stderr = %q, want usage message", got)
	}
}

func TestRun_UnknownCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"unknown"}, &stdout, &stderr)
	if exitCode != 1 {
		t.Fatalf("run(unknown) exit code = %d, want 1", exitCode)
	}
	if got := stderr.String(); !strings.Contains(got, "unknown command: unknown") {
		t.Fatalf("stderr = %q, want unknown command message", got)
	}
}

func TestRun_Categories(t *testing.T) {
	root := repoRoot(t)
	restore := chdir(t, root)
	defer restore()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"categories"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("run(categories) exit code = %d, want 0; stderr=%q", exitCode, stderr.String())
	}
	if got := stdout.String(); !strings.Contains(got, `"categories"`) {
		t.Fatalf("stdout = %q, want categories JSON", got)
	}
}

func TestRun_Optimize(t *testing.T) {
	root := repoRoot(t)
	restore := chdir(t, root)
	defer restore()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"optimize", "--category", "bugfix"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("run(optimize) exit code = %d, want 0; stderr=%q", exitCode, stderr.String())
	}
	if got := stdout.String(); !strings.Contains(got, `"category": "bugfix"`) {
		t.Fatalf("stdout = %q, want bugfix optimize output", got)
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
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
