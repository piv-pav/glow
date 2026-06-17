package index

import (
	"fmt"
	"strings"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search/query"
)

// SearchResult represents a search result
type SearchResult struct {
	Name    string
	Fields  map[string]interface{}
	Snippet string
}

// SearchOutput holds search results with total match count.
type SearchOutput struct {
	Results []SearchResult
	Total   int
}

// Search searches the index with query and filters
func (i *Index) Search(queryStr string, filters map[string]string, limit int) (*SearchOutput, error) {
	if limit <= 0 {
		limit = 10
	}

	// Parse query string to extract embedded filters
	textQuery, embeddedFilters := parseQueryString(queryStr)
	
	// Merge embedded filters with explicit filters
	for k, v := range embeddedFilters {
		if _, exists := filters[k]; !exists {
			filters[k] = v
		}
	}

	// Build Bleve query
	var queries []query.Query

	// Add text search if present — search content + all metadata fields
	if textQuery != "" {
		var textQueries []query.Query
		
		// Use cached fields if available, otherwise enumerate
		fields := i.fields
		if len(fields) == 0 {
			var err error
			fields, err = i.index.Fields()
			if err != nil {
				return nil, fmt.Errorf("failed to list index fields: %w", err)
			}
		}

		textQueries = make([]query.Query, 0, len(fields))
		
		for _, field := range fields {
			if field == "_all" || field == "_id" || field == "_type" {
				continue
			}
			q := bleve.NewMatchQuery(textQuery)
			q.SetField(field)
			if field == "content" {
				q.SetBoost(1.0)
			} else {
				q.SetBoost(1.5)
			}
			textQueries = append(textQueries, q)
		}

		if len(textQueries) > 0 {
			textDisjunction := bleve.NewDisjunctionQuery(textQueries...)
			queries = append(queries, textDisjunction)
		}
	}

	// Add filter queries
	for field, value := range filters {
		switch field {
		case "path":
			// Path uses prefix match (keyword field)
			prefixQuery := bleve.NewPrefixQuery(value)
			prefixQuery.SetField("path")
			queries = append(queries, prefixQuery)
		case "tag", "tags":
			// Tags use term query (keyword field, no analysis)
			termQuery := bleve.NewTermQuery(value)
			termQuery.SetField("tags")
			queries = append(queries, termQuery)
		default:
			// Other fields use match query
			fieldQuery := bleve.NewMatchQuery(value)
			fieldQuery.SetField(field)
			queries = append(queries, fieldQuery)
		}
	}

	// If no queries, match all
	var finalQuery query.Query
	if len(queries) == 0 {
		finalQuery = bleve.NewMatchAllQuery()
	} else if len(queries) == 1 {
		finalQuery = queries[0]
	} else {
		// Combine with boolean AND
		boolQuery := bleve.NewConjunctionQuery(queries...)
		finalQuery = boolQuery
	}

	// Execute search
	searchRequest := bleve.NewSearchRequest(finalQuery)
	searchRequest.Size = limit
	searchRequest.Fields = []string{"*"}
	searchRequest.Highlight = bleve.NewHighlight()

	searchResult, err := i.index.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Convert results
	results := make([]SearchResult, 0, len(searchResult.Hits))
	for _, hit := range searchResult.Hits {
		result := SearchResult{
			Name:   hit.ID,
			Fields: hit.Fields,
		}

		// Extract snippet from highlights
		if len(hit.Fragments) > 0 {
			for _, fragments := range hit.Fragments {
				if len(fragments) > 0 {
					result.Snippet = fragments[0]
					break
				}
			}
		}

		results = append(results, result)
	}

	return &SearchOutput{Results: results, Total: int(searchResult.Total)}, nil
}

// parseQueryString extracts field:value filters from query string
// Returns cleaned text query and extracted filters
func parseQueryString(queryStr string) (string, map[string]string) {
	filters := make(map[string]string)
	var textParts []string

	parts := strings.Fields(queryStr)
	for _, part := range parts {
		// Check for field:value pattern
		if colonIdx := strings.Index(part, ":"); colonIdx > 0 {
			field := part[:colonIdx]
			value := part[colonIdx+1:]

			if value != "" {
				// Handle comma-separated values (e.g., tag:go,cli)
				if strings.Contains(value, ",") {
					// For now, just use first value
					// TODO: support multiple values per field
					value = strings.Split(value, ",")[0]
				}
				filters[field] = value
				continue
			}
		}

		// Not a filter, add to text query
		textParts = append(textParts, part)
	}

	textQuery := strings.Join(textParts, " ")
	return textQuery, filters
}

// SearchAll returns all documents (for listing)
func (i *Index) SearchAll(limit int) (*SearchOutput, error) {
	return i.Search("", nil, limit)
}
