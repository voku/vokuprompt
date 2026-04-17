package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/voku/vokuprompt/internal/ast"
	"github.com/voku/vokuprompt/internal/engine"
	"github.com/voku/vokuprompt/internal/output"
)

type exampleFixture struct {
	Name                       string                         `json:"name"`
	OriginalWeakPrompt         string                         `json:"original_weak_prompt"`
	ChosenCategory             string                         `json:"chosen_category"`
	OptimizeCommand            string                         `json:"optimize_command"`
	ReturnedCompiledPrompt     string                         `json:"returned_compiled_prompt"`
	PlaceholderManifestExcerpt []ast.PlaceholderManifestEntry `json:"placeholder_manifest_excerpt"`
	PlaceholderResolution      map[string]string              `json:"placeholder_resolution"`
	FinalExecutablePrompt      string                         `json:"final_executable_prompt"`
}

var placeholderPattern = regexp.MustCompile(`\[[A-Z_]+\]`)

func TestExampleFixtures_MatchOptimizeOutputAndResolveExecutablePrompt(t *testing.T) {
	root := repoRoot(t)

	categories, err := engine.LoadCategories(filepath.Join(root, "categories.json"))
	if err != nil {
		t.Fatalf("load categories.json: %v", err)
	}

	registry := loadPlaceholderRegistry(t)
	registryNames := make([]string, 0, len(registry))
	for _, entry := range registry {
		registryNames = append(registryNames, entry.Name)
	}
	slices.Sort(registryNames)

	entries, err := os.ReadDir(filepath.Join(root, "examples"))
	if err != nil {
		t.Fatalf("read examples directory: %v", err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		names = append(names, entry.Name())
	}
	slices.Sort(names)

	wantNames := []string{
		"bugfix-flow.json",
		"performance-flow.json",
		"refactor-flow.json",
		"review-flow.json",
		"tests-flow.json",
	}
	if !slices.Equal(names, wantNames) {
		t.Fatalf("example fixtures = %v, want %v", names, wantNames)
	}

	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			example := loadExampleFixture(t, filepath.Join(root, "examples", name))

			if example.OriginalWeakPrompt == "" {
				t.Fatal("original_weak_prompt must not be empty")
			}
			if example.OptimizeCommand != "vokuprompt optimize --category "+example.ChosenCategory {
				t.Fatalf("optimize_command = %q, want real CLI invocation", example.OptimizeCommand)
			}
			if _, ok := engine.FindCategory(categories, example.ChosenCategory); !ok {
				t.Fatalf("chosen_category = %q, not present in categories.json", example.ChosenCategory)
			}

			var out bytes.Buffer
			if err := ExecuteOptimize([]string{"--category", example.ChosenCategory, "--patterns", filepath.Join(root, "patterns.json"), "--categories", filepath.Join(root, "categories.json")}, &out); err != nil {
				t.Fatalf("ExecuteOptimize: %v", err)
			}

			var resp output.OptimizeResponse
			if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
				t.Fatalf("parse optimize output: %v", err)
			}

			if example.ReturnedCompiledPrompt != resp.CompiledPrompt {
				t.Fatalf("returned_compiled_prompt mismatch\ngot:\n%s\nwant:\n%s", example.ReturnedCompiledPrompt, resp.CompiledPrompt)
			}
			if !reflect.DeepEqual(example.PlaceholderManifestExcerpt, resp.PlaceholderManifest) {
				t.Fatalf("placeholder_manifest_excerpt mismatch\ngot:  %+v\nwant: %+v", example.PlaceholderManifestExcerpt, resp.PlaceholderManifest)
			}

			resolutionNames := make([]string, 0, len(example.PlaceholderResolution))
			for name, value := range example.PlaceholderResolution {
				if value == "" {
					t.Fatalf("placeholder_resolution[%q] must not be empty", name)
				}
				resolutionNames = append(resolutionNames, name)
				if _, ok := registry[name]; !ok {
					t.Fatalf("placeholder_resolution[%q] missing from placeholders.json", name)
				}
			}
			slices.Sort(resolutionNames)

			manifestNames := make([]string, 0, len(resp.PlaceholderManifest))
			for _, entry := range resp.PlaceholderManifest {
				manifestNames = append(manifestNames, entry.Name)
			}
			slices.Sort(manifestNames)

			if !slices.Equal(resolutionNames, manifestNames) {
				t.Fatalf("placeholder_resolution keys = %v, want %v", resolutionNames, manifestNames)
			}

			finalPrompt := resolveExecutablePrompt(resp.CompiledPrompt, example.PlaceholderResolution, resp.ExecutionRequest)
			if finalPrompt != example.FinalExecutablePrompt {
				t.Fatalf("final_executable_prompt mismatch\ngot:\n%s\nwant:\n%s", finalPrompt, example.FinalExecutablePrompt)
			}
			if placeholderPattern.MatchString(finalPrompt) {
				t.Fatalf("final_executable_prompt still contains unresolved placeholders: %s", finalPrompt)
			}
			lowerRequest := strings.ToLower(resp.ExecutionRequest)
			if !strings.Contains(lowerRequest, "selected category") || !strings.Contains(lowerRequest, "final executable prompt") || !strings.Contains(lowerRequest, "execute the final prompt") {
				t.Fatalf("execution_request must instruct category framing + final prompt build + execute, got %q", resp.ExecutionRequest)
			}
		})
	}
}

func loadExampleFixture(t *testing.T, path string) exampleFixture {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	var example exampleFixture
	if err := json.Unmarshal(data, &example); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}

	return example
}

func loadPlaceholderRegistry(t *testing.T) map[string]placeholderRegistryEntry {
	t.Helper()

	root := repoRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "placeholders.json"))
	if err != nil {
		t.Fatalf("read placeholders.json: %v", err)
	}

	var entries []placeholderRegistryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		t.Fatalf("parse placeholders.json: %v", err)
	}

	registry := make(map[string]placeholderRegistryEntry, len(entries))
	for _, entry := range entries {
		registry[entry.Name] = entry
	}

	return registry
}

func resolveExecutablePrompt(compiled string, resolutions map[string]string, executionRequest string) string {
	replacements := make([]string, 0, len(resolutions)*2)
	for name, value := range resolutions {
		replacements = append(replacements, "["+name+"]", value)
	}

	resolved := strings.NewReplacer(replacements...).Replace(compiled)
	return resolved + "\n\n" + executionRequest
}
