package storage

import (
	"database/sql"
	"fmt"
	"path/filepath"

	"codeberg.org/pivpav/glow/internal/config"
	_ "modernc.org/sqlite"
)

// SQLiteStorage stores articles in a SQLite database (no CGO).
type SQLiteStorage struct {
	sqlStore
}

// NewSQLiteStorage opens (or creates) a SQLite DB for the given wiki.
func NewSQLiteStorage(wikiName string) (*SQLiteStorage, error) {
	wikiPath, err := config.GetWikiPath(wikiName)
	if err != nil {
		return nil, err
	}

	dbPath := filepath.Join(wikiPath, "articles.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite db: %w", err)
	}

	// WAL mode + foreign keys
	for _, pragma := range []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA foreign_keys=ON",
	} {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			return nil, fmt.Errorf("sqlite pragma failed: %w", err)
		}
	}

	s := &SQLiteStorage{sqlStore{db: db, ph: sqlitePH}}
	if err := s.fts5Migrate(db); err != nil {
		db.Close()
		return nil, err
	}

	return s, nil
}

func sqlitePH(_ int) string { return "?" }

// Search implements Searcher using shared FTS5 logic.
func (s *SQLiteStorage) Search(query string, filters map[string]string, limit int) (*SearchOutput, error) {
	return s.searchFTS5(query, filters, limit)
}
