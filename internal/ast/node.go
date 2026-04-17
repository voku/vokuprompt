package ast

type NodeType string

const (
	NodeTypeRoot        NodeType = "Root"
	NodeTypeFraming     NodeType = "Framing"
	NodeTypeConstraints NodeType = "Constraints"
	NodeTypeExecution   NodeType = "Execution"
	NodeTypeValidation  NodeType = "Validation"
	NodeTypeAnalysis    NodeType = "Analysis"
	NodeTypeReview      NodeType = "Review"
	NodeTypeDone        NodeType = "Done"
	NodeTypeOutput      NodeType = "Output"
)

type Node struct {
	Type           NodeType
	Text           string
	Placeholders   []string
	Children       []*Node
	SourcePatterns []string
}

type PlaceholderManifestEntry struct {
	Name           string   `json:"name"`
	Description    string   `json:"description,omitempty"`
	Examples       []string `json:"examples,omitempty"`
	Required       bool     `json:"required"`
	NodeTypes      []string `json:"node_types"`
	SourcePatterns []string `json:"source_patterns"`
}
