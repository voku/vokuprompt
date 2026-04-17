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

		// VALIDATION appears in two source patterns; check both name and order.
		validationEntry := manifest[3]
		wantSP := []string{"verify", "verify_dup"}
		if len(validationEntry.SourcePatterns) != len(wantSP) {
			t.Fatalf("run %d: VALIDATION source_patterns = %v, want %v", run+1, validationEntry.SourcePatterns, wantSP)
		}
		for i, sp := range wantSP {
			if validationEntry.SourcePatterns[i] != sp {
				t.Fatalf("run %d: VALIDATION source_patterns[%d] = %q, want %q", run+1, i, validationEntry.SourcePatterns[i], sp)
			}
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

func TestCompile_PlaceholderInMultipleSectionsMergesNodeTypes(t *testing.T) {
	// The same placeholder name appears in two different sections (framing + execution).
	// The manifest entry must list both NodeTypes, sorted.
	patterns := []engine.Pattern{
		{Name: "p1", Role: "framing", Prompt: "Fix [TARGET]", Placeholders: []string{"TARGET"}},
		{Name: "p2", Role: "execution", Prompt: "Run against [TARGET]", Placeholders: []string{"TARGET"}},
	}

	_, manifest := Compile(patterns)

	if len(manifest) != 1 {
		t.Fatalf("got %d manifest entries, want 1", len(manifest))
	}
	entry := manifest[0]
	if entry.Name != "TARGET" {
		t.Fatalf("manifest[0].Name = %q, want TARGET", entry.Name)
	}

	wantNodeTypes := []string{"Execution", "Framing"}
	if len(entry.NodeTypes) != len(wantNodeTypes) {
		t.Fatalf("NodeTypes = %v, want %v", entry.NodeTypes, wantNodeTypes)
	}
	for i, nt := range wantNodeTypes {
		if entry.NodeTypes[i] != nt {
			t.Fatalf("NodeTypes[%d] = %q, want %q", i, entry.NodeTypes[i], nt)
		}
	}

	wantSourcePatterns := []string{"p1", "p2"}
	if len(entry.SourcePatterns) != len(wantSourcePatterns) {
		t.Fatalf("SourcePatterns = %v, want %v", entry.SourcePatterns, wantSourcePatterns)
	}
	for i, sp := range wantSourcePatterns {
		if entry.SourcePatterns[i] != sp {
			t.Fatalf("SourcePatterns[%d] = %q, want %q", i, entry.SourcePatterns[i], sp)
		}
	}
}

func TestCompile_SectionOrderIsCanonicalRegardlessOfInputOrder(t *testing.T) {
	// Patterns given in non-canonical order: execution, framing, validation.
	// Compiled prompt must still appear in canonical section order.
	patterns := []engine.Pattern{
		{Name: "exec", Role: "execution", Prompt: "Execute"},
		{Name: "frame", Role: "framing", Prompt: "Frame"},
		{Name: "validate", Role: "validation", Prompt: "Validate"},
	}

	compiled, _ := Compile(patterns)
	const want = "Goal:\nFrame\n\nExecution:\n- Execute\n\nValidation:\n- Validate"
	if compiled != want {
		t.Fatalf("section order wrong\ngot:\n%s\nwant:\n%s", compiled, want)
	}
}
