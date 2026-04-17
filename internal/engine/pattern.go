package engine

type Pattern struct {
	Name         string             `json:"name"`
	Role         string             `json:"role"`
	Description  string             `json:"description,omitempty"`
	Prompt       string             `json:"prompt"`
	Placeholders []string           `json:"placeholders,omitempty"`
	Weights      map[string]float64 `json:"weights,omitempty"`
	Conflicts    []string           `json:"conflicts,omitempty"`
}
