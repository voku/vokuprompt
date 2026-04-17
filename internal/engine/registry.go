package engine

import (
	"encoding/json"
	"fmt"
	"os"
)

func LoadPatterns(path string) ([]Pattern, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read patterns: %w", err)
	}

	var patterns []Pattern
	if err := json.Unmarshal(data, &patterns); err != nil {
		return nil, fmt.Errorf("parse patterns: %w", err)
	}

	return patterns, nil
}

func LoadCategories(path string) ([]Category, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read categories: %w", err)
	}

	var categories []Category
	if err := json.Unmarshal(data, &categories); err != nil {
		return nil, fmt.Errorf("parse categories: %w", err)
	}

	return categories, nil
}

func indexPatterns(patterns []Pattern) map[string]Pattern {
	indexed := make(map[string]Pattern, len(patterns))
	for _, pattern := range patterns {
		indexed[pattern.Name] = pattern
	}

	return indexed
}

func FindCategory(categories []Category, name string) (Category, bool) {
	for _, category := range categories {
		if category.Name == name {
			return category, true
		}
	}

	return Category{}, false
}
