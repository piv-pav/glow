package index

import (
	"fmt"
	"os"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"codeberg.org/pivpav/glow/internal/article"
	"codeberg.org/pivpav/glow/internal/config"
)

// Index handles Bleve search indexing
type Index struct {
	WikiName string
	index    bleve.Index
	fields   []string
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

	i := &Index{
		WikiName: wikiName,
		index:    idx,
	}

	// Cache fields for search optimization
	if fields, err := idx.Fields(); err == nil {
		i.fields = fields
	}

	return i, nil
}

// createIndexMapping creates Bleve index mapping with keyword analyzers for structured fields
func createIndexMapping() mapping.IndexMapping {
	indexMapping := bleve.NewIndexMapping()

	// Article document mapping
	articleMapping := bleve.NewDocumentMapping()

	// Tags: keyword (no stemming/tokenization) for exact match
	keywordField := bleve.NewTextFieldMapping()
	keywordField.Analyzer = "keyword"
	articleMapping.AddFieldMappingsAt("tags", keywordField)

	// Path: keyword for exact prefix matching
	pathField := bleve.NewTextFieldMapping()
	pathField.Analyzer = "keyword"
	articleMapping.AddFieldMappingsAt("path", pathField)

	indexMapping.DefaultMapping = articleMapping

	return indexMapping
}

// Close closes the index
func (i *Index) Close() error {
	if i.index != nil {
		return i.index.Close()
	}
	return nil
}

// articleToDoc converts an article to an indexable document.
func articleToDoc(name string, art *article.Article) map[string]interface{} {
	doc := map[string]interface{}{
		"content": art.Content,
	}

	for key, val := range art.Frontmatter {
		switch v := val.(type) {
		case []interface{}:
			strs := make([]string, 0, len(v))
			for _, item := range v {
				if str, ok := item.(string); ok {
					strs = append(strs, str)
				}
			}
			// Keep as array for keyword fields (e.g. tags) — Bleve indexes each element
			doc[key] = strs
		case []string:
			doc[key] = v
		default:
			doc[key] = val
		}
	}

	if _, ok := doc["path"]; !ok {
		doc["path"] = name
	}

	return doc
}

// IndexArticle indexes or updates an article
func (i *Index) IndexArticle(name string, art *article.Article) error {
	return i.index.Index(name, articleToDoc(name, art))
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
		if err := batch.Index(name, articleToDoc(name, art)); err != nil {
			return fmt.Errorf("failed to add article to batch: %w", err)
		}

		count++

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

