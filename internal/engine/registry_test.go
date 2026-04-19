package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadPatterns(t *testing.T) {
	path := filepath.Join(t.TempDir(), "patterns.json")
	content := `[{"name":"goal","role":"framing","weights":{"bugfix":1},"conflicts":[],"prompt":"[TASK]","placeholders":["TASK"]}]`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write patterns.json: %v", err)
	}

	patterns, err := LoadPatterns(path)
	if err != nil {
		t.Fatalf("LoadPatterns returned error: %v", err)
	}
	if len(patterns) != 1 || patterns[0].Name != "goal" {
		t.Fatalf("LoadPatterns returned %v, want one goal pattern", patterns)
	}
}

func TestLoadPatterns_ReadAndParseErrors(t *testing.T) {
	_, err := LoadPatterns(filepath.Join(t.TempDir(), "missing.json"))
	if err == nil || !strings.Contains(err.Error(), "read patterns") {
		t.Fatalf("LoadPatterns read error = %v, want read patterns wrapper", err)
	}

	badPath := filepath.Join(t.TempDir(), "patterns.json")
	if err := os.WriteFile(badPath, []byte("{not-json"), 0o644); err != nil {
		t.Fatalf("write invalid patterns.json: %v", err)
	}
	_, err = LoadPatterns(badPath)
	if err == nil || !strings.Contains(err.Error(), "parse patterns") {
		t.Fatalf("LoadPatterns parse error = %v, want parse patterns wrapper", err)
	}
}

func TestLoadCategories(t *testing.T) {
	path := filepath.Join(t.TempDir(), "categories.json")
	content := `[{"name":"bugfix","description":"Fix a bug","role_limits":{"execution":1},"required_patterns":["goal"],"optional_patterns":["review"]}]`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write categories.json: %v", err)
	}

	categories, err := LoadCategories(path)
	if err != nil {
		t.Fatalf("LoadCategories returned error: %v", err)
	}
	if len(categories) != 1 || categories[0].Name != "bugfix" {
		t.Fatalf("LoadCategories returned %v, want one bugfix category", categories)
	}
}

func TestLoadCategories_ReadAndParseErrors(t *testing.T) {
	_, err := LoadCategories(filepath.Join(t.TempDir(), "missing.json"))
	if err == nil || !strings.Contains(err.Error(), "read categories") {
		t.Fatalf("LoadCategories read error = %v, want read categories wrapper", err)
	}

	badPath := filepath.Join(t.TempDir(), "categories.json")
	if err := os.WriteFile(badPath, []byte("{not-json"), 0o644); err != nil {
		t.Fatalf("write invalid categories.json: %v", err)
	}
	_, err = LoadCategories(badPath)
	if err == nil || !strings.Contains(err.Error(), "parse categories") {
		t.Fatalf("LoadCategories parse error = %v, want parse categories wrapper", err)
	}
}

func TestFindCategory(t *testing.T) {
	categories := []Category{
		{Name: "bugfix"},
		{Name: "performance"},
	}

	got, ok := FindCategory(categories, "performance")
	if !ok {
		t.Fatal("FindCategory(performance) = false, want true")
	}
	if got.Name != "performance" {
		t.Fatalf("FindCategory(performance) = %q, want performance", got.Name)
	}

	_, ok = FindCategory(categories, "missing")
	if ok {
		t.Fatal("FindCategory(missing) = true, want false")
	}
}
