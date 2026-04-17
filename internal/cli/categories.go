package cli

import (
	"flag"
	"io"

	"github.com/voku/vokuprompt/internal/engine"
	"github.com/voku/vokuprompt/internal/output"
)

func ExecuteCategories(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("categories", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	categoriesPath := fs.String("categories", "categories.json", "path to categories.json")
	if err := fs.Parse(args); err != nil {
		return err
	}

	categories, err := engine.LoadCategories(resolveAssetPath(*categoriesPath))
	if err != nil {
		return err
	}

	items := make([]output.CategoryItem, 0, len(categories))
	for _, category := range categories {
		items = append(items, output.CategoryItem{Name: category.Name, Description: category.Description})
	}

	return output.WriteJSON(out, output.CategoriesResponse{Categories: items})
}
