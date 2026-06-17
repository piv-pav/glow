package storage

import (
	"database/sql"
	"fmt"
	"strings"

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

	if err := rqliteMigrate(db); err != nil {
		db.Close()
		return nil, err
	}

	return &RqliteStorage{sqlStore{db: db, ph: sqlitePH}}, nil
}

// rqliteMigrate creates the schema — identical to SQLite (same engine under the hood).
func rqliteMigrate(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS articles (
		name     TEXT PRIMARY KEY,
		content  TEXT NOT NULL DEFAULT '',
		meta     TEXT NOT NULL DEFAULT '{}',
		tags     TEXT NOT NULL DEFAULT '',
		created  TEXT NOT NULL,
		modified TEXT NOT NULL
	)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE VIRTUAL TABLE IF NOT EXISTS articles_fts USING fts5(
		name, content, tags,
		content='articles', content_rowid='rowid'
	)`)
	if err != nil {
		return err
	}

	for _, trigger := range []string{
		`CREATE TRIGGER IF NOT EXISTS articles_ai AFTER INSERT ON articles BEGIN
			INSERT INTO articles_fts(rowid, name, content, tags) VALUES (new.rowid, new.name, new.content, new.tags);
		END`,
		`CREATE TRIGGER IF NOT EXISTS articles_ad AFTER DELETE ON articles BEGIN
			INSERT INTO articles_fts(articles_fts, rowid, name, content, tags) VALUES('delete', old.rowid, old.name, old.content, old.tags);
		END`,
		`CREATE TRIGGER IF NOT EXISTS articles_au AFTER UPDATE ON articles BEGIN
			INSERT INTO articles_fts(articles_fts, rowid, name, content, tags) VALUES('delete', old.rowid, old.name, old.content, old.tags);
			INSERT INTO articles_fts(rowid, name, content, tags) VALUES (new.rowid, new.name, new.content, new.tags);
		END`,
	} {
		if _, err := db.Exec(trigger); err != nil {
			return err
		}
	}
	return nil
}

// Search implements Searcher using FTS5 (same as SQLite — rqlite is SQLite underneath).
func (s *RqliteStorage) Search(query string, filters map[string]string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 10
	}

	conditions, args, _ := s.buildSearchConditions(filters, 1, "a.")

	var sqlStr string
	if query != "" {
		conditions = append(conditions, `a.rowid IN (SELECT rowid FROM articles_fts WHERE articles_fts MATCH ?)`)
		args = append(args, query)

		where := "WHERE " + strings.Join(conditions, " AND ")
		sqlStr = `SELECT a.name, a.tags,
			snippet(articles_fts, 1, '<b>', '</b>', '...', 20) AS snippet,
			0.0 AS score
			FROM articles a
			JOIN articles_fts ON articles_fts.rowid = a.rowid
			` + where + `
			LIMIT ?`
	} else {
		where := ""
		if len(conditions) > 0 {
			where = "WHERE " + strings.Join(conditions, " AND ")
		}
		sqlStr = `SELECT a.name, a.tags, '' AS snippet, 0.0 AS score
			FROM articles a ` + where + ` ORDER BY a.name LIMIT ?`
	}
	args = append(args, limit)

	rows, err := s.db.Query(sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	defer rows.Close()

	return s.scanSearchResults(rows)
}
