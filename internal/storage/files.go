package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"codeberg.org/pivpav/glow/internal/article"
	"codeberg.org/pivpav/glow/internal/config"
)

// FileStorage handles article file operations.
type FileStorage struct {
	WikiName string
}

// NewFileStorage creates a file-based storage instance for a wiki.
func NewFileStorage(wikiName string) *FileStorage {
	if wikiName == "" {
		wikiName = "default"
	}
	return &FileStorage{WikiName: wikiName}
}

// Close is a no-op for file storage (satisfies Store interface).
func (s *FileStorage) Close() error { return nil }

// articlePath returns full file path for article name
// Supports nested paths like "folder/subfolder/article"
func (s *FileStorage) articlePath(name string) (string, error) {
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
func (s *FileStorage) ensureDir(articlePath string) error {
	dir := filepath.Dir(articlePath)
	return os.MkdirAll(dir, 0755)
}

// Create creates a new article
func (s *FileStorage) Create(name string, article *article.Article) error {
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

	// Add path to frontmatter
	article.Frontmatter["path"] = s.normalizeArticleName(name)

	return s.write(path, article)
}

// Read reads an article
func (s *FileStorage) Read(name string) (*article.Article, error) {
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
func (s *FileStorage) Update(name string, article *article.Article) error {
	path, err := s.articlePath(name)
	if err != nil {
		return err
	}

	// Check if exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("article not found: %s", name)
	}

	// Update path in frontmatter
	article.Frontmatter["path"] = s.normalizeArticleName(name)

	return s.write(path, article)
}

// Delete deletes an article
func (s *FileStorage) Delete(name string) error {
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
func (s *FileStorage) Move(oldName, newName string) error {
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

	// Read article to update path in frontmatter
	art, err := s.Read(oldName)
	if err != nil {
		return err
	}

	// Update path in frontmatter
	art.Frontmatter["path"] = s.normalizeArticleName(newName)

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
func (s *FileStorage) List() ([]string, error) {
	articlesPath, err := config.GetArticlesPath(s.WikiName)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(articlesPath); os.IsNotExist(err) {
		return nil, nil
	}

	var articles []string

	err = filepath.WalkDir(articlesPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && filepath.Ext(path) == ".md" {
			relPath, err := filepath.Rel(articlesPath, path)
			if err != nil {
				return err
			}

			name := strings.TrimSuffix(relPath, ".md")
			articles = append(articles, name)
		}

		return nil
	})

	return articles, err
}

// write writes article to file, updating timestamps
func (s *FileStorage) write(path string, article *article.Article) error {
	// Ensure created timestamp exists
	if _, exists := article.Frontmatter["created"]; !exists {
		article.Frontmatter["created"] = time.Now().Format(time.RFC3339)
	}
	// Always update modified timestamp
	article.Frontmatter["modified"] = time.Now().Format(time.RFC3339)

	data, err := article.Serialize()
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// normalizeArticleName removes .md extension and cleans path
func (s *FileStorage) normalizeArticleName(name string) string {
	name = strings.TrimSuffix(name, ".md")
	return filepath.ToSlash(name)
}

// cleanupEmptyDirs removes empty parent directories up to articles root
func (s *FileStorage) cleanupEmptyDirs(dir string) {
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
