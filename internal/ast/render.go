package ast

import (
	"strings"
)

func Render(root *Node) string {
	var b strings.Builder

	for sectionIndex, section := range root.Children {
		title := sectionTitle(section.Type)
		if title != "" {
			b.WriteString(title)
			b.WriteString(":\n")
		}
		for _, node := range section.Children {
			if section.Type == NodeTypeFraming || section.Type == NodeTypeDone {
				b.WriteString(node.Text)
				b.WriteString("\n")
			} else {
				b.WriteString("- ")
				b.WriteString(node.Text)
				b.WriteString("\n")
			}
		}
		if sectionIndex < len(root.Children)-1 {
			b.WriteString("\n")
		}
	}

	return strings.TrimSpace(b.String())
}

func sectionTitle(nodeType NodeType) string {
	switch nodeType {
	case NodeTypeFraming:
		return "Goal"
	case NodeTypeConstraints:
		return "Constraints"
	case NodeTypeExecution:
		return "Execution"
	case NodeTypeValidation:
		return "Validation"
	case NodeTypeAnalysis:
		return "Analysis"
	case NodeTypeReview:
		return "Review"
	case NodeTypeDone:
		return "Done when"
	case NodeTypeOutput:
		return "Output"
	default:
		return ""
	}
}
