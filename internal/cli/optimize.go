package cli

import (
	"flag"
	"fmt"
	"io"

	"github.com/voku/vokuprompt/internal/ast"
	"github.com/voku/vokuprompt/internal/engine"
	"github.com/voku/vokuprompt/internal/output"
)

const executionRequest = "Analyze the original prompt, improve it, and execute the improved prompt now.\n\nReturn:\n1. failure analysis\n2. improved prompt\n3. execution result"

func ExecuteOptimize(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("optimize", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	categoryName := fs.String("category", "", "category name")
	explain := fs.Bool("explain", false, "include pattern selection explanation")
	patternsPath := fs.String("patterns", "patterns.json", "path to patterns.json")
	categoriesPath := fs.String("categories", "categories.json", "path to categories.json")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *categoryName == "" {
		return fmt.Errorf("--category is required")
	}

	patterns, err := engine.LoadPatterns(*patternsPath)
	if err != nil {
		return err
	}
	categories, err := engine.LoadCategories(*categoriesPath)
	if err != nil {
		return err
	}

	category, ok := engine.FindCategory(categories, *categoryName)
	if !ok {
		return fmt.Errorf("unknown category: %s", *categoryName)
	}

	trace := engine.SelectPatternsWithTrace(patterns, category)
	selected := trace.Selected
	selectedNames := make([]string, 0, len(selected))
	for _, pattern := range selected {
		selectedNames = append(selectedNames, pattern.Name)
	}

	compiledPrompt, manifest := ast.Compile(selected)
	requiredPlaceholders := make([]string, 0, len(manifest))
	for _, entry := range manifest {
		requiredPlaceholders = append(requiredPlaceholders, entry.Name)
	}

	var explanation *output.OptimizeExplanation
	if *explain {
		explanation = &output.OptimizeExplanation{
			RejectedByConflict:   make([]output.OptimizeConflictRejection, 0, len(trace.RejectedByConflict)),
			RejectedByRoleLimits: make([]output.OptimizeRoleLimitRejection, 0, len(trace.RejectedByRoleLimits)),
			RequiredPlaceholders: requiredPlaceholders,
		}
		for _, rejection := range trace.RejectedByConflict {
			explanation.RejectedByConflict = append(explanation.RejectedByConflict, output.OptimizeConflictRejection{
				Name:          rejection.Name,
				ConflictsWith: rejection.ConflictsWith,
			})
		}
		for _, rejection := range trace.RejectedByRoleLimits {
			explanation.RejectedByRoleLimits = append(explanation.RejectedByRoleLimits, output.OptimizeRoleLimitRejection{
				Name:  rejection.Name,
				Role:  rejection.Role,
				Limit: rejection.Limit,
			})
		}
	}

	return output.WriteJSON(out, output.OptimizeResponse{
		Category:            category.Name,
		SelectedPatterns:    selectedNames,
		CompiledPrompt:      compiledPrompt,
		PlaceholderManifest: manifest,
		ExecutionRequest:    executionRequest,
		Explanation:         explanation,
	})
}
