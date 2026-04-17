package cli

import (
	"flag"
	"fmt"
	"io"

	"github.com/voku/vokuprompt/internal/ast"
	"github.com/voku/vokuprompt/internal/engine"
	"github.com/voku/vokuprompt/internal/output"
)

const executionRequest = "Treat the selected category as fixed input chosen from the deterministic category registry.\n1. Use placeholder_manifest as the source of truth for which placeholders must be resolved now.\n2. Build the final executable prompt by resolving every required placeholder in compiled_prompt from repository facts and the current task context.\n3. Keep the selected category and compiled structure intact; do not silently rewrite the contract.\n4. If a required placeholder cannot be resolved safely, stop and ask for the missing input.\n5. After placeholder resolution, execute the final prompt.\n\nReturn:\n1. category confirmation\n2. placeholder resolution summary\n3. final executable prompt\n4. execution result"

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
			RequiredPatterns:     trace.RequiredSelected,
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
				Name:      rejection.Name,
				Role:      rejection.Role,
				Limit:     rejection.Limit,
				BlockedBy: rejection.BlockedBy,
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
