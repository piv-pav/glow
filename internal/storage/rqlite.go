package storage

import (
	"database/sql"
	"fmt"

	_ "github.com/rqlite/gorqlite/stdlib"
)

// RqliteStorage stores articles in a rqlite cluster (distributed SQLite, no CGO).
type RqliteStorage struct {
	sqlStore
}

// NewRqliteStorage connects to a rqlite cluster and ensures the schema exists.
// url is the connection string, e.g. "http://localhost:4001/" or "https://user:pass@host:4001/?level=weak".
func NewRqliteStorage(url string) (*RqliteStorage, error) {
	db, err := sql.Open("rqlite", url)
	if err != nil {
		return nil, fmt.Errorf("failed to open rqlite connection: %w", err)
	}

	// Verify connectivity.
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("rqlite ping failed: %w", err)
	}

	s := &RqliteStorage{sqlStore{db: db, ph: sqlitePH}}
	if err := s.fts5Migrate(db); err != nil {
		db.Close()
		return nil, err
	}

	return s, nil
}

// Search implements Searcher using shared FTS5 logic.
func (s *RqliteStorage) Search(query string, filters map[string]string, limit int) (*SearchOutput, error) {
	return s.searchFTS5(query, filters, limit)
}
