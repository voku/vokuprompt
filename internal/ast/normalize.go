package ast

import (
	"sort"

	"github.com/voku/vokuprompt/internal/engine"
)

func Normalize(nodes []*Node) *Node {
	grouped := make(map[NodeType][]*Node)
	for _, node := range nodes {
		grouped[node.Type] = append(grouped[node.Type], node)
	}

	sections := make([]*Node, 0, len(grouped))
	for nodeType, sectionNodes := range grouped {
		dedup := make(map[string]*Node)
		children := make([]*Node, 0, len(sectionNodes))
		for _, node := range sectionNodes {
			if existing, exists := dedup[node.Text]; exists {
				existing.SourcePatterns = append(existing.SourcePatterns, node.SourcePatterns...)
				continue
			}
			dedup[node.Text] = node
			children = append(children, node)
		}
		sort.SliceStable(children, func(i, j int) bool {
			return children[i].SourcePatterns[0] < children[j].SourcePatterns[0]
		})
		sections = append(sections, &Node{Type: nodeType, Children: children})
	}

	sort.SliceStable(sections, func(i, j int) bool {
		ri := engine.RoleOrder[roleForNodeType(sections[i].Type)]
		rj := engine.RoleOrder[roleForNodeType(sections[j].Type)]
		return ri < rj
	})

	return &Node{Type: NodeTypeRoot, Children: sections}
}

func roleForNodeType(t NodeType) string {
	switch t {
	case NodeTypeFraming:
		return "framing"
	case NodeTypeConstraints:
		return "constraint"
	case NodeTypeExecution:
		return "execution"
	case NodeTypeValidation:
		return "validation"
	case NodeTypeAnalysis:
		return "analysis"
	case NodeTypeReview:
		return "review"
	case NodeTypeDone:
		return "done"
	case NodeTypeOutput:
		return "output"
	default:
		return ""
	}
}
