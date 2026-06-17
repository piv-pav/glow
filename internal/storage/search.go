package storage

// SearchResult is a single search hit.
type SearchResult struct {
	Name    string
	Snippet string
	Tags    []string
}

// SearchOutput holds search results with total match count.
type SearchOutput struct {
	Results []SearchResult
	Total   int
}

// Searcher is implemented by backends that support native full-text search.
type Searcher interface {
	Search(query string, filters map[string]string, limit int) (*SearchOutput, error)
}
