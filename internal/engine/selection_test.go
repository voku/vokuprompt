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
