package index

import (
	"fmt"
	"os"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"git.netra.pivpav.com/public/glow/internal/article"
	"git.netra.pivpav.com/public/glow/internal/config"
)

// Index handles Bleve search indexing
type Index struct {
	WikiName string
	index    bleve.Index
}

// New creates or opens an index for a wiki
func New(wikiName string) (*Index, error) {
	if wikiName == "" {
		wikiName = "default"
	}

	indexPath, err := config.GetIndexPath(wikiName)
	if err != nil {
		return nil, err
	}

	var idx bleve.Index

	// Try to open existing index
	idx, err = bleve.Open(indexPath)
	if err == bleve.ErrorIndexPathDoesNotExist {
		// Create new index
		idx, err = bleve.New(indexPath, createIndexMapping())
		if err != nil {
			return nil, fmt.Errorf("failed to create index: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to open index: %w", err)
	}

	return &Index{
		WikiName: wikiName,
		index:    idx,
	}, nil
}

// createIndexMapping creates Bleve index mapping
func createIndexMapping() mapping.IndexMapping {
	mapping := bleve.NewIndexMapping()
	return mapping
}

// Close closes the index
func (i *Index) Close() error {
	if i.index != nil {
		return i.index.Close()
	}
	return nil
}

// IndexArticle indexes or updates an article
func (i *Index) IndexArticle(name string, article *article.Article) error {
	// Build document for indexing
	doc := map[string]interface{}{
		"content": article.Content,
	}

	// Add all metadata fields (flattened for search)
	for key, val := range article.GetAllMetadataForIndex() {
		doc[key] = val
	}

	// Ensure path is set
	if _, ok := doc["path"]; !ok {
		doc["path"] = name
	}

	return i.index.Index(name, doc)
}

// DeleteArticle removes article from index
func (i *Index) DeleteArticle(name string) error {
	return i.index.Delete(name)
}

// UpdateArticle updates article in index (alias for IndexArticle)
func (i *Index) UpdateArticle(name string, article *article.Article) error {
	return i.IndexArticle(name, article)
}

// Rebuild completely rebuilds the index from all articles
func (i *Index) Rebuild(articles map[string]*article.Article) error {
	// Close existing index
	if err := i.Close(); err != nil {
		return fmt.Errorf("failed to close existing index: %w", err)
	}

	// Delete old index
	indexPath, err := config.GetIndexPath(i.WikiName)
	if err != nil {
		return err
	}

	if err := os.RemoveAll(indexPath); err != nil {
		return fmt.Errorf("failed to remove old index: %w", err)
	}

	// Create new index
	idx, err := bleve.New(indexPath, createIndexMapping())
	if err != nil {
		return fmt.Errorf("failed to create new index: %w", err)
	}
	i.index = idx

	// Index all articles
	batch := i.index.NewBatch()
	batchSize := 100

	count := 0
	for name, art := range articles {
		doc := map[string]interface{}{
			"content": art.Content,
		}

		for key, val := range art.GetAllMetadataForIndex() {
			doc[key] = val
		}

		if _, ok := doc["path"]; !ok {
			doc["path"] = name
		}

		if err := batch.Index(name, doc); err != nil {
			return fmt.Errorf("failed to add article to batch: %w", err)
		}

		count++

		// Execute batch every batchSize articles
		if count%batchSize == 0 {
			if err := i.index.Batch(batch); err != nil {
				return fmt.Errorf("failed to execute batch: %w", err)
			}
			batch = i.index.NewBatch()
		}
	}

	// Execute remaining batch
	if batch.Size() > 0 {
		if err := i.index.Batch(batch); err != nil {
			return fmt.Errorf("failed to execute final batch: %w", err)
		}
	}

	return nil
}

// Verify checks index health and returns stats
func (i *Index) Verify() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get document count
	docCount, err := i.index.DocCount()
	if err != nil {
		return nil, fmt.Errorf("failed to get document count: %w", err)
	}

	stats["document_count"] = docCount

	// Try a basic search to verify functionality
	query := bleve.NewMatchAllQuery()
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Size = 1

	searchResult, err := i.index.Search(searchRequest)
	if err != nil {
		stats["searchable"] = false
		stats["error"] = err.Error()
		return stats, fmt.Errorf("index search failed: %w", err)
	}

	stats["searchable"] = true
	stats["total_hits"] = searchResult.Total

	return stats, nil
}
