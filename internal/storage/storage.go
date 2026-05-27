package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"codeberg.org/pivpav/glow/internal/article"
	"codeberg.org/pivpav/glow/internal/config"
)

// Storage handles article file operations
type Storage struct {
	WikiName string
}

// New creates a new storage instance for a wiki
func New(wikiName string) *Storage {
	if wikiName == "" {
		wikiName = "default"
	}
	return &Storage{WikiName: wikiName}
}

// articlePath returns full file path for article name
// Supports nested paths like "folder/subfolder/article"
func (s *Storage) articlePath(name string) (string, error) {
	articlesPath, err := config.GetArticlesPath(s.WikiName)
	if err != nil {
		return "", err
	}

	// Ensure .md extension
	if !strings.HasSuffix(name, ".md") {
		name = name + ".md"
	}

	fullPath := filepath.Join(articlesPath, name)
	
	return fullPath, nil
}

// ensureDir creates parent directories for article path
func (s *Storage) ensureDir(articlePath string) error {
	dir := filepath.Dir(articlePath)
	return os.MkdirAll(dir, 0755)
}

// Create creates a new article
func (s *Storage) Create(name string, article *article.Article) error {
	path, err := s.articlePath(name)
	if err != nil {
		return err
	}

	// Check if already exists
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("article already exists: %s", name)
	}

	// Ensure parent directory exists
	if err := s.ensureDir(path); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Add path metadata
	article.Metadata["path"] = s.normalizeArticleName(name)

	return s.write(path, article)
}

// Read reads an article
func (s *Storage) Read(name string) (*article.Article, error) {
	path, err := s.articlePath(name)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("article not found: %s", name)
		}
		return nil, err
	}

	art, err := article.Parse(data)
	if err != nil {
		return nil, err
	}

	art.FilePath = path
	return art, nil
}

// Update updates an existing article
func (s *Storage) Update(name string, article *article.Article) error {
	path, err := s.articlePath(name)
	if err != nil {
		return err
	}

	// Check if exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("article not found: %s", name)
	}

	// Update path metadata
	article.Metadata["path"] = s.normalizeArticleName(name)

	return s.write(path, article)
}

// Delete deletes an article
func (s *Storage) Delete(name string) error {
	path, err := s.articlePath(name)
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("article not found: %s", name)
		}
		return err
	}

	// Clean up empty parent directories
	s.cleanupEmptyDirs(filepath.Dir(path))

	return nil
}

// Move renames/moves an article
func (s *Storage) Move(oldName, newName string) error {
	oldPath, err := s.articlePath(oldName)
	if err != nil {
		return err
	}

	newPath, err := s.articlePath(newName)
	if err != nil {
		return err
	}

	// Check old exists
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return fmt.Errorf("article not found: %s", oldName)
	}

	// Check new doesn't exist
	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("destination already exists: %s", newName)
	}

	// Ensure new parent directory exists
	if err := s.ensureDir(newPath); err != nil {
		return fmt.Errorf("failed to create destination directories: %w", err)
	}

	// Read article to update path metadata
	art, err := s.Read(oldName)
	if err != nil {
		return err
	}

	// Update path metadata
	art.Metadata["path"] = s.normalizeArticleName(newName)

	// Write to new location
	if err := s.write(newPath, art); err != nil {
		return err
	}

	// Delete old
	if err := os.Remove(oldPath); err != nil {
		// Rollback new file
		os.Remove(newPath)
		return err
	}

	// Clean up empty directories
	s.cleanupEmptyDirs(filepath.Dir(oldPath))

	return nil
}

// List returns all article names in the wiki
func (s *Storage) List() ([]string, error) {
	articlesPath, err := config.GetArticlesPath(s.WikiName)
	if err != nil {
		return nil, err
	}

	var articles []string

	err = filepath.Walk(articlesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".md") {
			// Get relative path from articles directory
			relPath, err := filepath.Rel(articlesPath, path)
			if err != nil {
				return err
			}

			// Remove .md extension
			name := strings.TrimSuffix(relPath, ".md")
			articles = append(articles, name)
		}

		return nil
	})

	return articles, err
}

// write writes article to file
func (s *Storage) write(path string, article *article.Article) error {
	data, err := article.Serialize()
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// normalizeArticleName removes .md extension and cleans path
func (s *Storage) normalizeArticleName(name string) string {
	name = strings.TrimSuffix(name, ".md")
	return filepath.ToSlash(name)
}

// cleanupEmptyDirs removes empty parent directories up to articles root
func (s *Storage) cleanupEmptyDirs(dir string) {
	articlesPath, err := config.GetArticlesPath(s.WikiName)
	if err != nil {
		return
	}

	// Don't delete articles root
	if dir == articlesPath {
		return
	}

	// Check if empty
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) > 0 {
		return
	}

	// Remove if empty
	os.Remove(dir)

	// Recurse to parent
	s.cleanupEmptyDirs(filepath.Dir(dir))
}
