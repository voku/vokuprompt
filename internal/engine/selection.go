package engine

import "sort"

// RoleOrder defines the canonical ordering of pattern roles.
// The ast package uses this same ordering for section layout in the rendered prompt.
// analysis is ordered before validation so review prompts read: execute → analyze → validate → challenge.
var RoleOrder = map[string]int{
	"framing":    0,
	"constraint": 1,
	"execution":  2,
	"analysis":   3,
	"validation": 4,
	"review":     5,
	"done":       6,
	"output":     7,
}

type scoredPattern struct {
	pattern Pattern
	score   float64
}

type ConflictRejection struct {
	Name          string
	ConflictsWith []string
}

type RoleLimitRejection struct {
	Name      string
	Role      string
	Limit     int
	BlockedBy string // name of the pattern that already occupies this role slot
}

type SelectionTrace struct {
	Selected             []Pattern
	RequiredSelected     []string // names of required patterns that were found and selected
	RejectedByConflict   []ConflictRejection
	RejectedByRoleLimits []RoleLimitRejection
}

func SelectPatterns(patterns []Pattern, category Category) []Pattern {
	return SelectPatternsWithTrace(patterns, category).Selected
}

func SelectPatternsWithTrace(patterns []Pattern, category Category) SelectionTrace {
	indexed := indexPatterns(patterns)
	selected := make([]Pattern, 0, len(category.RequiredPatterns)+len(category.OptionalPatterns))
	selectedNames := make(map[string]struct{})
	roleUsage := make(map[string]int)
	roleOccupant := make(map[string]string) // first pattern name selected per role
	trace := SelectionTrace{}

	for _, required := range category.RequiredPatterns {
		pattern, ok := indexed[required]
		if !ok {
			continue
		}
		selected = append(selected, pattern)
		selectedNames[pattern.Name] = struct{}{}
		roleUsage[pattern.Role]++
		if roleOccupant[pattern.Role] == "" {
			roleOccupant[pattern.Role] = pattern.Name
		}
		trace.RequiredSelected = append(trace.RequiredSelected, pattern.Name)
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
			trace.RejectedByRoleLimits = append(trace.RejectedByRoleLimits, RoleLimitRejection{
				Name:      pattern.Name,
				Role:      pattern.Role,
				Limit:     limit,
				BlockedBy: roleOccupant[pattern.Role],
			})
			continue
		}
		if conflicts := conflictsWithSelected(pattern, selected); len(conflicts) > 0 {
			trace.RejectedByConflict = append(trace.RejectedByConflict, ConflictRejection{
				Name:          pattern.Name,
				ConflictsWith: conflicts,
			})
			continue
		}
		selected = append(selected, pattern)
		selectedNames[pattern.Name] = struct{}{}
		roleUsage[pattern.Role]++
		if roleOccupant[pattern.Role] == "" {
			roleOccupant[pattern.Role] = pattern.Name
		}
	}

	sort.SliceStable(selected, func(i, j int) bool {
		pi := RoleOrder[selected[i].Role]
		pj := RoleOrder[selected[j].Role]
		if pi == pj {
			return selected[i].Name < selected[j].Name
		}
		return pi < pj
	})

	trace.Selected = selected
	return trace
}

func conflictsWithSelected(candidate Pattern, selected []Pattern) []string {
	var conflicts []string
	for _, existing := range selected {
		if hasConflict(candidate.Conflicts, existing.Name) || hasConflict(existing.Conflicts, candidate.Name) {
			conflicts = append(conflicts, existing.Name)
		}
	}

	return conflicts
}

func hasConflict(conflicts []string, name string) bool {
	for _, conflict := range conflicts {
		if conflict == name {
			return true
		}
	}
	return false
}
