package engine

import "testing"

func TestSelectPatterns_RespectsRoleLimitsAndConflicts(t *testing.T) {
	patterns := []Pattern{
		{Name: "goal", Role: "framing"},
		{Name: "verification", Role: "validation", Weights: map[string]float64{"bugfix": 0.9}},
		{Name: "step_loop", Role: "execution", Weights: map[string]float64{"bugfix": 0.8}, Conflicts: []string{"multi_pass"}},
		{Name: "multi_pass", Role: "execution", Weights: map[string]float64{"bugfix": 0.7}},
	}
	category := Category{
		Name:             "bugfix",
		RoleLimits:       map[string]int{"execution": 1, "validation": 1},
		RequiredPatterns: []string{"goal"},
		OptionalPatterns: []string{"multi_pass", "step_loop", "verification"},
	}

	selected := SelectPatterns(patterns, category)

	names := make([]string, 0, len(selected))
	for _, p := range selected {
		names = append(names, p.Name)
	}

	want := []string{"goal", "step_loop", "verification"}
	if len(names) != len(want) {
		t.Fatalf("got %v patterns, want %v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("got %v, want %v", names, want)
		}
	}
}

func TestSelectPatterns_RequiredPatternBlocksConflictingOptional(t *testing.T) {
	// optional "alt" declares a conflict with required "core"; alt must be rejected.
	patterns := []Pattern{
		{Name: "core", Role: "execution"},
		{Name: "alt", Role: "execution", Weights: map[string]float64{"test": 0.9}, Conflicts: []string{"core"}},
	}
	category := Category{
		Name:             "test",
		RequiredPatterns: []string{"core"},
		OptionalPatterns: []string{"alt"},
	}

	selected := SelectPatterns(patterns, category)

	if len(selected) != 1 || selected[0].Name != "core" {
		names := make([]string, len(selected))
		for i, p := range selected {
			names[i] = p.Name
		}
		t.Fatalf("got %v, want [core]", names)
	}
}

func TestSelectPatterns_ConflictIsSymmetric(t *testing.T) {
	// Only "beta" declares a conflict with "alpha". When beta is selected first (higher
	// score), alpha must still be rejected because the conflict check is bidirectional.
	patterns := []Pattern{
		{Name: "alpha", Role: "execution", Weights: map[string]float64{"test": 0.5}},
		{Name: "beta", Role: "execution", Weights: map[string]float64{"test": 0.9}, Conflicts: []string{"alpha"}},
	}
	category := Category{
		Name:             "test",
		OptionalPatterns: []string{"alpha", "beta"},
	}

	selected := SelectPatterns(patterns, category)

	if len(selected) != 1 || selected[0].Name != "beta" {
		names := make([]string, len(selected))
		for i, p := range selected {
			names[i] = p.Name
		}
		t.Fatalf("got %v, want [beta]", names)
	}
}

func TestSelectPatterns_RoleLimitAloneBlocks(t *testing.T) {
	// Two execution patterns with no conflict. Role limit = 1, so only the higher-scored
	// pattern is selected. This disentangles the role-limit enforcement from conflict logic.
	patterns := []Pattern{
		{Name: "exec_a", Role: "execution", Weights: map[string]float64{"test": 0.9}},
		{Name: "exec_b", Role: "execution", Weights: map[string]float64{"test": 0.4}},
	}
	category := Category{
		Name:             "test",
		RoleLimits:       map[string]int{"execution": 1},
		OptionalPatterns: []string{"exec_a", "exec_b"},
	}

	selected := SelectPatterns(patterns, category)

	if len(selected) != 1 || selected[0].Name != "exec_a" {
		names := make([]string, len(selected))
		for i, p := range selected {
			names[i] = p.Name
		}
		t.Fatalf("got %v, want [exec_a]", names)
	}
}
