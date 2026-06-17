package storage

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"codeberg.org/pivpav/glow/internal/article"
)

// placeholder defines how a SQL backend generates parameter placeholders.
type placeholder func(n int) string

// sqlStore is the shared implementation for SQL-backed stores (sqlite, pgsql).
type sqlStore struct {
	db *sql.DB
	ph placeholder // returns $1/$2/... or ?/?/...
}

func (s *sqlStore) Close() error { return s.db.Close() }

func (s *sqlStore) Create(name string, art *article.Article) error {
	now := time.Now().Format(time.RFC3339)
	if _, ok := art.Frontmatter["created"]; !ok {
		art.Frontmatter["created"] = now
	}
	art.Frontmatter["modified"] = now
	art.Frontmatter["path"] = name

	meta, err := marshalMeta(art.Frontmatter)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(
		`INSERT INTO articles(name,content,meta,tags,created,modified) VALUES(`+s.ph(1)+`,`+s.ph(2)+`,`+s.ph(3)+`,`+s.ph(4)+`,`+s.ph(5)+`,`+s.ph(6)+`)`,
		name, art.Content, meta, tagsFromFrontmatter(art.Frontmatter),
		art.Frontmatter["created"], now,
	)
	if err != nil {
		return fmt.Errorf("article already exists: %s", name)
	}
	return nil
}

func (s *sqlStore) Read(name string) (*article.Article, error) {
	var content, meta, created, modified string
	err := s.db.QueryRow(
		`SELECT content, meta, created, modified FROM articles WHERE name=`+s.ph(1), name,
	).Scan(&content, &meta, &created, &modified)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("article not found: %s", name)
	}
	if err != nil {
		return nil, err
	}

	fm, err := unmarshalMeta(meta)
	if err != nil {
		return nil, err
	}
	fm["created"] = created
	fm["modified"] = modified
	fm["path"] = name

	return &article.Article{Frontmatter: fm, Content: content}, nil
}

func (s *sqlStore) Update(name string, art *article.Article) error {
	now := time.Now().Format(time.RFC3339)
	art.Frontmatter["modified"] = now
	art.Frontmatter["path"] = name

	meta, err := marshalMeta(art.Frontmatter)
	if err != nil {
		return err
	}

	res, err := s.db.Exec(
		`UPDATE articles SET content=`+s.ph(1)+`, meta=`+s.ph(2)+`, tags=`+s.ph(3)+`, modified=`+s.ph(4)+` WHERE name=`+s.ph(5),
		art.Content, meta, tagsFromFrontmatter(art.Frontmatter), now, name,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("article not found: %s", name)
	}
	return nil
}

func (s *sqlStore) Delete(name string) error {
	res, err := s.db.Exec(`DELETE FROM articles WHERE name=`+s.ph(1), name)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("article not found: %s", name)
	}
	return nil
}

func (s *sqlStore) Move(oldName, newName string) error {
	art, err := s.Read(oldName)
	if err != nil {
		return err
	}

	art.Frontmatter["path"] = newName
	meta, err := marshalMeta(art.Frontmatter)
	if err != nil {
		return err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	if _, err := tx.Exec(
		`INSERT INTO articles(name,content,meta,tags,created,modified) SELECT `+s.ph(1)+`,content,`+s.ph(2)+`,tags,created,modified FROM articles WHERE name=`+s.ph(3),
		newName, meta, oldName,
	); err != nil {
		return fmt.Errorf("destination already exists: %s", newName)
	}
	if _, err := tx.Exec(`DELETE FROM articles WHERE name=`+s.ph(1), oldName); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *sqlStore) List() ([]string, error) {
	rows, err := s.db.Query(`SELECT name FROM articles ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		names = append(names, n)
	}
	return names, rows.Err()
}

// buildSearchConditions builds WHERE conditions and args from filters.
// argStart is the starting placeholder number (1-based).
// prefix is the table alias prefix (e.g. "a." or "").
func (s *sqlStore) buildSearchConditions(filters map[string]string, argStart int, prefix string) ([]string, []interface{}, int) {
	var conditions []string
	var args []interface{}
	i := argStart

	if tag, ok := filters["tag"]; ok {
		conditions = append(conditions, prefix+`tags LIKE `+s.ph(i))
		args = append(args, "%"+tag+"%")
		i++
	}
	if path, ok := filters["path"]; ok {
		conditions = append(conditions, prefix+`name LIKE `+s.ph(i))
		args = append(args, path+"%")
		i++
	}
	return conditions, args, i
}

func (s *sqlStore) scanSearchResults(rows *sql.Rows) (*SearchOutput, error) {
	out := &SearchOutput{}
	for rows.Next() {
		var r SearchResult
		var tagsStr string
		var score float64 // scanned but discarded; used for ORDER BY in query
		var total int
		if err := rows.Scan(&r.Name, &tagsStr, &r.Snippet, &score, &total); err != nil {
			return nil, err
		}
		if tagsStr != "" {
			r.Tags = strings.Fields(tagsStr)
		}
		out.Results = append(out.Results, r)
		out.Total = total
	}
	return out, rows.Err()
}
