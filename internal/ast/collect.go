package ast

import "sort"

func CollectPlaceholders(root *Node) []PlaceholderManifestEntry {
	type aggregate struct {
		nodeTypes      map[string]struct{}
		sourcePatterns map[string]struct{}
	}

	byName := make(map[string]*aggregate)
	for _, section := range root.Children {
		for _, node := range section.Children {
			for _, placeholder := range node.Placeholders {
				entry, exists := byName[placeholder]
				if !exists {
					entry = &aggregate{
						nodeTypes:      map[string]struct{}{},
						sourcePatterns: map[string]struct{}{},
					}
					byName[placeholder] = entry
				}
				entry.nodeTypes[string(section.Type)] = struct{}{}
				for _, sp := range node.SourcePatterns {
					entry.sourcePatterns[sp] = struct{}{}
				}
			}
		}
	}

	result := make([]PlaceholderManifestEntry, 0, len(byName))
	for name, aggregate := range byName {
		nodeTypes := toSortedList(aggregate.nodeTypes)
		sourcePatterns := toSortedList(aggregate.sourcePatterns)
		result = append(result, PlaceholderManifestEntry{
			Name:           name,
			Required:       true,
			NodeTypes:      nodeTypes,
			SourcePatterns: sourcePatterns,
		})
	}

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result
}

func toSortedList(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
