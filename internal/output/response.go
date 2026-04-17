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
}

func WriteJSON(w io.Writer, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}
