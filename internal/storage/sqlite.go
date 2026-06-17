package storage

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"

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

	if err := sqliteMigrate(db); err != nil {
		db.Close()
		return nil, err
	}

	return &SQLiteStorage{sqlStore{db: db, ph: sqlitePH}}, nil
}

func sqlitePH(_ int) string { return "?" }

func sqliteMigrate(db *sql.DB) error {
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
	// FTS5 for full-text search over content + tags
	_, err = db.Exec(`CREATE VIRTUAL TABLE IF NOT EXISTS articles_fts USING fts5(
		name, content, tags,
		content='articles', content_rowid='rowid'
	)`)
	if err != nil {
		return err
	}
	// Keep FTS in sync via triggers
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

// Search implements Searcher using SQLite FTS5.
func (s *SQLiteStorage) Search(query string, filters map[string]string, limit int) ([]SearchResult, error) {
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
