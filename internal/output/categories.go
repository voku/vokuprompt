package output

type CategoryItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CategoriesResponse struct {
	Categories []CategoryItem `json:"categories"`
}
