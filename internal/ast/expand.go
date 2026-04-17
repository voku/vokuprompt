package ast

import "github.com/voku/vokuprompt/internal/engine"

func ExpandPatterns(patterns []engine.Pattern) []*Node {
	nodes := make([]*Node, 0, len(patterns))
	for _, pattern := range patterns {
		nodes = append(nodes, &Node{
			Type:          roleToNodeType(pattern.Role),
			Key:           pattern.Name,
			Text:          pattern.Prompt,
			Placeholders:  append([]string(nil), pattern.Placeholders...),
			SourcePattern: pattern.Name,
		})
	}

	return nodes
}

func roleToNodeType(role string) NodeType {
	switch role {
	case "framing":
		return NodeTypeFraming
	case "constraint":
		return NodeTypeConstraints
	case "execution":
		return NodeTypeExecution
	case "validation":
		return NodeTypeValidation
	case "analysis":
		return NodeTypeAnalysis
	case "review":
		return NodeTypeReview
	case "done":
		return NodeTypeDone
	case "output":
		return NodeTypeOutput
	default:
		return NodeTypeExecution
	}
}
