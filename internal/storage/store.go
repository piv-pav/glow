package storage

import "codeberg.org/pivpav/glow/internal/article"

// Store is the backend-agnostic storage interface.
type Store interface {
	Create(name string, art *article.Article) error
	Read(name string) (*article.Article, error)
	Update(name string, art *article.Article) error
	Delete(name string) error
	Move(oldName, newName string) error
	List() ([]string, error)
	Close() error
}
