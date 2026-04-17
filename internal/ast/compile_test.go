package ast

import (
	"testing"

	"github.com/voku/vokuprompt/internal/engine"
)

func TestCompile_ExactOutputAndDeterminism(t *testing.T) {
	patterns := []engine.Pattern{
		{Name: "goal", Role: "framing", Prompt: "[TASK]\n\nContext:\n[CONTEXT_TARGET]", Placeholders: []string{"TASK", "CONTEXT_TARGET"}},
		{Name: "scope", Role: "constraint", Prompt: "Do not change [SCOPE]", Placeholders: []string{"SCOPE"}},
		{Name: "verify", Role: "validation", Prompt: "Run [VALIDATION]", Placeholders: []string{"VALIDATION"}},
		{Name: "verify_dup", Role: "validation", Prompt: "Run [VALIDATION]", Placeholders: []string{"VALIDATION"}},
	}

	const wantPrompt = "Goal:\n[TASK]\n\nContext:\n[CONTEXT_TARGET]\n\nConstraints:\n- Do not change [SCOPE]\n\nValidation:\n- Run [VALIDATION]"

	// Run twice to prove determinism.
	for run := range 2 {
		compiled, manifest := Compile(patterns)

		if compiled != wantPrompt {
			t.Fatalf("run %d: compiled prompt mismatch\ngot:\n%s\nwant:\n%s", run+1, compiled, wantPrompt)
		}

		if len(manifest) != 4 {
			t.Fatalf("run %d: got %d placeholders, want 4", run+1, len(manifest))
		}

		// Manifest must be sorted by name.
		wantNames := []string{"CONTEXT_TARGET", "SCOPE", "TASK", "VALIDATION"}
		for i, entry := range manifest {
			if entry.Name != wantNames[i] {
				t.Fatalf("run %d: manifest[%d] = %q, want %q", run+1, i, entry.Name, wantNames[i])
			}
			if !entry.Required {
				t.Fatalf("run %d: manifest[%d].Required should be true", run+1, i)
			}
		}

		// VALIDATION appears in two source patterns; both should be captured.
		validationEntry := manifest[3]
		if len(validationEntry.SourcePatterns) != 2 {
			t.Fatalf("run %d: VALIDATION source_patterns = %v, want [verify verify_dup]", run+1, validationEntry.SourcePatterns)
		}
	}
}

func TestCompile_FramingHeaderNotDoubled(t *testing.T) {
	// Regression: if a framing prompt embeds the section title the renderer must not add it again.
	// The fix is that patterns must NOT embed the "Goal:" header — the renderer adds it.
	patterns := []engine.Pattern{
		{Name: "framing", Role: "framing", Prompt: "[TASK]", Placeholders: []string{"TASK"}},
	}

	compiled, _ := Compile(patterns)
	const want = "Goal:\n[TASK]"
	if compiled != want {
		t.Fatalf("framing rendered as:\n%s\nwant:\n%s", compiled, want)
	}
}
