package ast

import (
	"strings"
	"testing"

	"github.com/voku/vokuprompt/internal/engine"
)

func TestCompile_RendersPromptAndManifestDeterministically(t *testing.T) {
	patterns := []engine.Pattern{
		{Name: "goal", Role: "framing", Prompt: "Fix [TASK]", Placeholders: []string{"TASK"}},
		{Name: "scope", Role: "constraint", Prompt: "Do not change [SCOPE]", Placeholders: []string{"SCOPE"}},
		{Name: "verify", Role: "validation", Prompt: "Run [VALIDATION]", Placeholders: []string{"VALIDATION"}},
		{Name: "verify_dup", Role: "validation", Prompt: "Run [VALIDATION]", Placeholders: []string{"VALIDATION"}},
	}

	compiled, manifest := Compile(patterns)

	if !strings.Contains(compiled, "Goal:\nFix [TASK]") {
		t.Fatalf("compiled prompt missing goal section: %s", compiled)
	}
	if !strings.Contains(compiled, "Validation:\n- Run [VALIDATION]") {
		t.Fatalf("compiled prompt missing validation section: %s", compiled)
	}
	if len(manifest) != 3 {
		t.Fatalf("got %d placeholders, want 3", len(manifest))
	}
	if manifest[2].Name != "VALIDATION" {
		t.Fatalf("expected sorted placeholders, got %+v", manifest)
	}
}
