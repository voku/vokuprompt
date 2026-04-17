package output

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/voku/vokuprompt/internal/ast"
)

type OptimizeResponse struct {
	Category            string                         `json:"category"`
	SelectedPatterns    []string                       `json:"selected_patterns"`
	CompiledPrompt      string                         `json:"compiled_prompt"`
	PlaceholderManifest []ast.PlaceholderManifestEntry `json:"placeholder_manifest"`
	ExecutionRequest    string                         `json:"execution_request"`
	Explanation         *OptimizeExplanation           `json:"explanation,omitempty"`
}

type OptimizeExplanation struct {
	RejectedByConflict   []OptimizeConflictRejection  `json:"rejected_by_conflict,omitempty"`
	RejectedByRoleLimits []OptimizeRoleLimitRejection `json:"rejected_by_role_limits,omitempty"`
	RequiredPlaceholders []string                     `json:"required_placeholders"`
}

type OptimizeConflictRejection struct {
	Name          string   `json:"name"`
	ConflictsWith []string `json:"conflicts_with"`
}

type OptimizeRoleLimitRejection struct {
	Name  string `json:"name"`
	Role  string `json:"role"`
	Limit int    `json:"limit"`
}

func WriteJSON(w io.Writer, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}
