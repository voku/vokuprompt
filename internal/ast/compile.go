package ast

import "github.com/voku/vokuprompt/internal/engine"

func Compile(patterns []engine.Pattern) (string, []PlaceholderManifestEntry) {
	expanded := ExpandPatterns(patterns)
	normalized := Normalize(expanded)
	placeholders := CollectPlaceholders(normalized)
	rendered := Render(normalized)
	return rendered, placeholders
}
