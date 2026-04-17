package engine

type Category struct {
	Name             string         `json:"name"`
	Description      string         `json:"description"`
	RoleLimits       map[string]int `json:"role_limits"`
	RequiredPatterns []string       `json:"required_patterns"`
	OptionalPatterns []string       `json:"optional_patterns"`
	Focus            []string       `json:"focus,omitempty"`
}
