package engine

import "sort"

var rolePriority = map[string]int{
	"framing":    0,
	"constraint": 1,
	"execution":  2,
	"validation": 3,
	"analysis":   4,
	"review":     5,
	"done":       6,
	"output":     7,
}

type scoredPattern struct {
	pattern Pattern
	score   float64
}

func SelectPatterns(patterns []Pattern, category Category) []Pattern {
	indexed := IndexPatterns(patterns)
	selected := make([]Pattern, 0, len(category.RequiredPatterns)+len(category.OptionalPatterns))
	selectedNames := make(map[string]struct{})
	roleUsage := make(map[string]int)

	for _, required := range category.RequiredPatterns {
		pattern, ok := indexed[required]
		if !ok {
			continue
		}
		selected = append(selected, pattern)
		selectedNames[pattern.Name] = struct{}{}
		roleUsage[pattern.Role]++
	}

	scored := make([]scoredPattern, 0, len(category.OptionalPatterns))
	for _, optional := range category.OptionalPatterns {
		pattern, ok := indexed[optional]
		if !ok {
			continue
		}
		scored = append(scored, scoredPattern{pattern: pattern, score: pattern.Weights[category.Name]})
	}

	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].score == scored[j].score {
			return scored[i].pattern.Name < scored[j].pattern.Name
		}
		return scored[i].score > scored[j].score
	})

	for _, candidate := range scored {
		pattern := candidate.pattern
		if _, exists := selectedNames[pattern.Name]; exists {
			continue
		}
		if limit, hasLimit := category.RoleLimits[pattern.Role]; hasLimit && roleUsage[pattern.Role] >= limit {
			continue
		}
		if conflictsWithSelected(pattern, selected) {
			continue
		}
		selected = append(selected, pattern)
		selectedNames[pattern.Name] = struct{}{}
		roleUsage[pattern.Role]++
	}

	sort.SliceStable(selected, func(i, j int) bool {
		pi := rolePriority[selected[i].Role]
		pj := rolePriority[selected[j].Role]
		if pi == pj {
			return selected[i].Name < selected[j].Name
		}
		return pi < pj
	})

	return selected
}

func conflictsWithSelected(candidate Pattern, selected []Pattern) bool {
	for _, existing := range selected {
		if hasConflict(candidate.Conflicts, existing.Name) || hasConflict(existing.Conflicts, candidate.Name) {
			return true
		}
	}

	return false
}

func hasConflict(conflicts []string, name string) bool {
	for _, conflict := range conflicts {
		if conflict == name {
			return true
		}
	}
	return false
}
