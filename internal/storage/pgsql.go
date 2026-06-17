package storage

import (
	"database/sql"
	"fmt"
	"strings"

	"codeberg.org/pivpav/glow/internal/config"
	_ "github.com/lib/pq"
)

// EnsurePgDatabase connects to the postgres system DB and creates dbname if it doesn't exist.
// Returns (true, nil) if it created the DB, (false, nil) if it already existed.
func EnsurePgDatabase(cfg *config.PgSQLConfig) (created bool, err error) {
	admin := *cfg
	admin.DBName = "postgres"
	db, err := sql.Open("postgres", admin.DSN())
	if err != nil {
		return false, fmt.Errorf("failed to connect to postgres: %w", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		return false, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	var exists bool
	if err = db.QueryRow(`SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=$1)`, cfg.DBName).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to query pg_database: %w", err)
	}
	if exists {
		return false, nil
	}
	if _, err := db.Exec(`CREATE DATABASE ` + cfg.DBName); err != nil {
		return false, fmt.Errorf("failed to create database %q: %w", cfg.DBName, err)
	}
	return true, nil
}

// PgSQLStorage stores articles in a PostgreSQL database.
type PgSQLStorage struct {
	sqlStore
}

func NewPgSQLStorage(dsn string) (*PgSQLStorage, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open pgsql: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to pgsql: %w", err)
	}
	if err := pgsqlMigrate(db); err != nil {
		db.Close()
		return nil, err
	}
	return &PgSQLStorage{sqlStore{db: db, ph: pgsqlPH}}, nil
}

func pgsqlPH(n int) string { return fmt.Sprintf("$%d", n) }

func pgsqlMigrate(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS articles (
		name     TEXT PRIMARY KEY,
		content  TEXT NOT NULL DEFAULT '',
		meta     TEXT NOT NULL DEFAULT '{}',
		tags     TEXT NOT NULL DEFAULT '',
		created  TEXT NOT NULL,
		modified TEXT NOT NULL,
		tsv      tsvector GENERATED ALWAYS AS (
			setweight(to_tsvector('english', coalesce(name,'')), 'A') ||
			setweight(to_tsvector('english', coalesce(tags,'')), 'A') ||
			setweight(to_tsvector('english', coalesce(content,'')), 'B')
		) STORED
	)`)
	if err != nil {
		return err
	}
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS articles_tsv_idx ON articles USING GIN(tsv)`)
	return err
}

// Search implements Searcher using PostgreSQL tsvector + GIN index.
func (s *PgSQLStorage) Search(query string, filters map[string]string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 10
	}

	conditions, args, i := s.buildSearchConditions(filters, 1, "")

	if query != "" {
		conditions = append(conditions, fmt.Sprintf(`tsv @@ plainto_tsquery('english', $%d)`, i))
		args = append(args, query)
		i++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	var sqlStr string
	if query != "" {
		sqlStr = fmt.Sprintf(`
			SELECT name, tags,
				ts_headline('english', content, plainto_tsquery('english', $%d), 'MaxFragments=1,MaxWords=20,MinWords=5') AS snippet,
				ts_rank(tsv, plainto_tsquery('english', $%d)) AS score
			FROM articles %s
			ORDER BY score DESC
			LIMIT $%d`, i, i, where, i+1)
		args = append(args, query, limit)
	} else {
		sqlStr = fmt.Sprintf(`
			SELECT name, tags, '' AS snippet, 0.0 AS score
			FROM articles %s
			ORDER BY name
			LIMIT $%d`, where, i)
		args = append(args, limit)
	}

	rows, err := s.db.Query(sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	defer rows.Close()

	return s.scanSearchResults(rows)
}
