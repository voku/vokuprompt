package ast

import "sort"

var sectionOrder = map[NodeType]int{
	NodeTypeFraming:     0,
	NodeTypeConstraints: 1,
	NodeTypeExecution:   2,
	NodeTypeValidation:  3,
	NodeTypeAnalysis:    4,
	NodeTypeReview:      5,
	NodeTypeDone:        6,
	NodeTypeOutput:      7,
}

func Normalize(nodes []*Node) *Node {
	grouped := make(map[NodeType][]*Node)
	for _, node := range nodes {
		grouped[node.Type] = append(grouped[node.Type], node)
	}

	sections := make([]*Node, 0, len(grouped))
	for nodeType, sectionNodes := range grouped {
		dedup := make(map[string]struct{})
		children := make([]*Node, 0, len(sectionNodes))
		for _, node := range sectionNodes {
			key := node.Text + "\x00" + node.SourcePattern
			if _, exists := dedup[key]; exists {
				continue
			}
			dedup[key] = struct{}{}
			children = append(children, node)
		}
		sort.SliceStable(children, func(i, j int) bool {
			return children[i].SourcePattern < children[j].SourcePattern
		})
		sections = append(sections, &Node{Type: nodeType, Children: children})
	}

	sort.SliceStable(sections, func(i, j int) bool {
		return sectionOrder[sections[i].Type] < sectionOrder[sections[j].Type]
	})

	return &Node{Type: NodeTypeRoot, Children: sections}
}
