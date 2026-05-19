package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

// GetWikiBasePath returns base path for all wikis
// Uses WIKI_DATA env var if set, otherwise XDG_DATA_HOME/glow/wiki
func GetWikiBasePath() (string, error) {
	if customPath := os.Getenv("WIKI_DATA"); customPath != "" {
		return customPath, nil
	}

	basePath := filepath.Join(xdg.DataHome, "glow", "wiki")
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return "", err
	}

	return basePath, nil
}

// GetWikiPath returns path for specific wiki
// Does NOT auto-create, only returns path
func GetWikiPath(wikiName string) (string, error) {
	if wikiName == "" {
		wikiName = "default"
	}

	basePath, err := GetWikiBasePath()
	if err != nil {
		return "", err
	}

	wikiPath := filepath.Join(basePath, wikiName)
	return wikiPath, nil
}

// CreateWiki creates a new wiki directory structure
func CreateWiki(wikiName string) error {
	if wikiName == "" {
		return fmt.Errorf("wiki name cannot be empty")
	}

	wikiPath, err := GetWikiPath(wikiName)
	if err != nil {
		return err
	}

	// Check if already exists
	if _, err := os.Stat(wikiPath); err == nil {
		return fmt.Errorf("wiki already exists: %s", wikiName)
	}

	// Create wiki directories
	articlesPath := filepath.Join(wikiPath, "articles")
	if err := os.MkdirAll(articlesPath, 0755); err != nil {
		return fmt.Errorf("failed to create wiki directories: %w", err)
	}

	return nil
}

// EnsureWikiExists creates wiki if it doesn't exist (used for default)
func EnsureWikiExists(wikiName string) error {
	if wikiName == "" {
		wikiName = "default"
	}

	wikiPath, err := GetWikiPath(wikiName)
	if err != nil {
		return err
	}

	// Create if doesn't exist
	if _, err := os.Stat(wikiPath); os.IsNotExist(err) {
		articlesPath := filepath.Join(wikiPath, "articles")
		if err := os.MkdirAll(articlesPath, 0755); err != nil {
			return fmt.Errorf("failed to create wiki directories: %w", err)
		}
	}

	return nil
}

// WikiExists checks if wiki exists
func WikiExists(wikiName string) (bool, error) {
	wikiPath, err := GetWikiPath(wikiName)
	if err != nil {
		return false, err
	}

	if _, err := os.Stat(wikiPath); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

// ListWikis returns all available wikis
func ListWikis() ([]string, error) {
	basePath, err := GetWikiBasePath()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(basePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var wikis []string
	for _, entry := range entries {
		if entry.IsDir() {
			wikis = append(wikis, entry.Name())
		}
	}

	return wikis, nil
}

// GetArticlesPath returns articles directory for wiki
func GetArticlesPath(wikiName string) (string, error) {
	wikiPath, err := GetWikiPath(wikiName)
	if err != nil {
		return "", err
	}
	return filepath.Join(wikiPath, "articles"), nil
}

// GetIndexPath returns bleve index path for wiki
func GetIndexPath(wikiName string) (string, error) {
	wikiPath, err := GetWikiPath(wikiName)
	if err != nil {
		return "", err
	}
	return filepath.Join(wikiPath, "index.bleve"), nil
}
