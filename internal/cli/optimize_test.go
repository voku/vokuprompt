package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestExecuteOptimize_RequiresCategory(t *testing.T) {
	var out bytes.Buffer
	err := ExecuteOptimize(nil, &out)
	if err == nil {
		t.Fatal("expected error for missing category")
	}
}

func TestExecuteCategories_ReadsConfiguredFile(t *testing.T) {
	var out bytes.Buffer
	err := ExecuteCategories([]string{"--categories", "../../categories.json"}, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "\"name\": \"bugfix\"") {
		t.Fatalf("missing bugfix category in output: %s", out.String())
	}
}
