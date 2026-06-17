package storage

import (
	"encoding/json"
	"fmt"
	"strings"

	"codeberg.org/pivpav/glow/internal/config"
)

// New opens the correct Store for wikiName based on its config.
func New(wikiName string) (Store, error) {
	if wikiName == "" {
		wikiName = "default"
	}
	cfg, err := config.GetWikiConfig(wikiName)
	if err != nil {
		return nil, fmt.Errorf("failed to read wiki config: %w", err)
	}
	if cfg == nil {
		return nil, fmt.Errorf("wiki %q not found — run: glow init %s", wikiName, wikiName)
	}

	switch cfg.Backend {
	case config.BackendRqlite:
		if cfg.Rqlite == nil {
			return nil, fmt.Errorf("rqlite backend requires [rqlite] config block in glow.yaml")
		}
		return NewRqliteStorage(cfg.Rqlite.ConnString())
	default: // sqlite (and empty / unknown)
		return NewSQLiteStorage(wikiName)
	}
}

// tagsFromFrontmatter extracts tags as a space-separated string for the tags column.
func tagsFromFrontmatter(fm map[string]interface{}) string {
	switch v := fm["tags"].(type) {
	case []interface{}:
		parts := make([]string, 0, len(v))
		for _, t := range v {
			if s, ok := t.(string); ok {
				parts = append(parts, s)
			}
		}
		return strings.Join(parts, " ")
	case []string:
		return strings.Join(v, " ")
	case string:
		return v
	}
	return ""
}

// created/modified/path are stored in dedicated columns, so we skip them here.
func marshalMeta(fm map[string]interface{}) (string, error) {
	m := make(map[string]interface{}, len(fm))
	for k, v := range fm {
		if k == "created" || k == "modified" || k == "path" {
			continue
		}
		m[k] = v
	}
	b, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("failed to marshal meta: %w", err)
	}
	return string(b), nil
}

// unmarshalMeta deserialises JSON back to a frontmatter map.
func unmarshalMeta(meta string) (map[string]interface{}, error) {
	fm := make(map[string]interface{})
	if meta == "" || meta == "{}" {
		return fm, nil
	}
	if err := json.Unmarshal([]byte(meta), &fm); err != nil {
		return nil, fmt.Errorf("failed to unmarshal meta: %w", err)
	}
	return fm, nil
}
