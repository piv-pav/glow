package storage

// SearchResult is a single search hit.
type SearchResult struct {
	Name    string
	Snippet string
	Tags    []string
}

// Searcher is implemented by DB-backed stores that can search natively.
// File-backed wikis fall back to Bleve.
type Searcher interface {
	Search(query string, filters map[string]string, limit int) ([]SearchResult, error)
}
