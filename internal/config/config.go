package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/adrg/xdg"
	"gopkg.in/yaml.v3"
)

// BackendType identifies the storage backend.
type BackendType string

const (
	BackendFiles  BackendType = "files"
	BackendSQLite BackendType = "sqlite"
	BackendPgSQL  BackendType = "pgsql"
	BackendRqlite BackendType = "rqlite"
)

// WikiConfig holds per-wiki configuration.
type WikiConfig struct {
	// DataPath overrides the default data directory for this wiki.
	// Defaults to <GLOW_DATA>/<name> if empty.
	DataPath string        `yaml:"data_path,omitempty"`
	Backend  BackendType   `yaml:"backend"`
	PgSQL    *PgSQLConfig  `yaml:"pgsql,omitempty"`
	Rqlite   *RqliteConfig `yaml:"rqlite,omitempty"`
}

// RqliteConfig holds rqlite connection parameters.
type RqliteConfig struct {
	URL              string `yaml:"url"`                // e.g. "http://localhost:4001" or "https://glow.example.com"
	User             string `yaml:"user,omitempty"`
	Password         string `yaml:"password,omitempty"`
	Level            string `yaml:"level,omitempty"`             // none, weak, strong (default: weak)
	DisableDiscovery bool   `yaml:"disable_discovery,omitempty"` // disable cluster discovery (use behind reverse proxy)
}

// ConnString builds the gorqlite connection string.
func (r *RqliteConfig) ConnString() string {
	userinfo := ""
	if r.User != "" {
		userinfo = url.PathEscape(r.User)
		if r.Password != "" {
			userinfo += ":" + url.PathEscape(r.Password)
		}
		userinfo += "@"
	}
	var params []string
	if r.Level != "" {
		params = append(params, "level="+r.Level)
	}
	if r.DisableDiscovery {
		params = append(params, "disableClusterDiscovery=true")
	}
	query := ""
	if len(params) > 0 {
		query = "?" + strings.Join(params, "&")
	}
	// Strip scheme, inject userinfo, re-add scheme.
	url := r.URL
	scheme := "http"
	if strings.HasPrefix(url, "https://") {
		scheme = "https"
		url = strings.TrimPrefix(url, "https://")
	} else {
		url = strings.TrimPrefix(url, "http://")
	}
	url = strings.TrimRight(url, "/")
	return scheme + "://" + userinfo + url + "/" + query
}

// PgSQLConfig holds PostgreSQL connection parameters.
type PgSQLConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port,omitempty"`
	DBName   string `yaml:"dbname"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	SSLMode  string `yaml:"sslmode,omitempty"`
}

// DSN converts PgSQLConfig to a lib/pq connection string.
func (p *PgSQLConfig) DSN() string {
	sslmode := p.SSLMode
	if sslmode == "" {
		sslmode = "disable"
	}
	dsn := fmt.Sprintf("host=%s dbname=%s user=%s password=%s sslmode=%s",
		p.Host, p.DBName, p.User, p.Password, sslmode)
	if p.Port != 0 {
		dsn += fmt.Sprintf(" port=%d", p.Port)
	}
	return dsn
}

// AppConfig is the top-level glow.yaml structure.
type AppConfig struct {
	Wikis map[string]*WikiConfig `yaml:"wikis"`
}

// GetConfigPath returns the path to glow.yaml.
// Priority: GLOW_CONFIG env > XDG_CONFIG_HOME/glow/glow.yaml
func GetConfigPath() string {
	if p := os.Getenv("GLOW_CONFIG"); p != "" {
		return p
	}
	return filepath.Join(xdg.ConfigHome, "glow", "glow.yaml")
}

// GetWikiBasePath returns the base directory for wiki data.
// Priority: GLOW_DATA env > XDG_DATA_HOME/glow/wiki
func GetWikiBasePath() (string, error) {
	if p := os.Getenv("GLOW_DATA"); p != "" {
		return p, nil
	}
	base := filepath.Join(xdg.DataHome, "glow", "wiki")
	if err := os.MkdirAll(base, 0755); err != nil {
		return "", err
	}
	return base, nil
}

// LoadConfig reads glow.yaml, returning an empty config if the file doesn't exist yet.
func LoadConfig() (*AppConfig, error) {
	path := GetConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &AppConfig{Wikis: map[string]*WikiConfig{}}, nil
		}
		return nil, fmt.Errorf("failed to read config %s: %w", path, err)
	}
	var cfg AppConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	if cfg.Wikis == nil {
		cfg.Wikis = map[string]*WikiConfig{}
	}
	return &cfg, nil
}

// SaveConfig writes the config to glow.yaml (0600).
func SaveConfig(cfg *AppConfig) error {
	path := GetConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	return os.WriteFile(path, data, 0600)
}

// GetWikiConfig returns the WikiConfig for wikiName, or nil if not registered.
func GetWikiConfig(wikiName string) (*WikiConfig, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	if wc, ok := cfg.Wikis[wikiName]; ok {
		return wc, nil
	}
	return nil, nil
}

// GetWikiPath returns the data directory for a wiki.
func GetWikiPath(wikiName string) (string, error) {
	if wikiName == "" {
		wikiName = "default"
	}
	cfg, err := LoadConfig()
	if err != nil {
		return "", err
	}
	if wc, ok := cfg.Wikis[wikiName]; ok && wc.DataPath != "" {
		return wc.DataPath, nil
	}
	base, err := GetWikiBasePath()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, wikiName), nil
}

var validWikiName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

// CreateWiki registers a new wiki in glow.yaml and creates its data directory.
func CreateWiki(wikiName string, wc *WikiConfig) error {
	if wikiName == "" {
		return fmt.Errorf("wiki name cannot be empty")
	}
	if !validWikiName.MatchString(wikiName) {
		return fmt.Errorf("invalid wiki name %q: use only letters, digits, hyphens, underscores", wikiName)
	}

	cfg, err := LoadConfig()
	if err != nil {
		return err
	}
	if _, exists := cfg.Wikis[wikiName]; exists {
		return fmt.Errorf("wiki already exists: %s", wikiName)
	}

	// Resolve data path
	dataPath := wc.DataPath
	if dataPath == "" {
		base, err := GetWikiBasePath()
		if err != nil {
			return err
		}
		dataPath = filepath.Join(base, wikiName)
	}

	if err := os.MkdirAll(dataPath, 0755); err != nil {
		return fmt.Errorf("failed to create wiki data directory: %w", err)
	}

	cfg.Wikis[wikiName] = wc
	return SaveConfig(cfg)
}

// WikiExists reports whether a wiki is registered in glow.yaml.
func WikiExists(wikiName string) (bool, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return false, err
	}
	_, ok := cfg.Wikis[wikiName]
	return ok, nil
}

// DeleteWiki removes a wiki from glow.yaml.
func DeleteWiki(wikiName string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}
	if _, ok := cfg.Wikis[wikiName]; !ok {
		return fmt.Errorf("wiki not found: %s", wikiName)
	}
	delete(cfg.Wikis, wikiName)
	return SaveConfig(cfg)
}

// ListWikis returns all wiki names from glow.yaml.
func ListWikis() ([]string, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(cfg.Wikis))
	for name := range cfg.Wikis {
		names = append(names, name)
	}
	return names, nil
}

// GetArticlesPath returns articles directory for file-based storage.
func GetArticlesPath(wikiName string) (string, error) {
	wikiPath, err := GetWikiPath(wikiName)
	if err != nil {
		return "", err
	}
	return filepath.Join(wikiPath, "articles"), nil
}

// GetIndexPath returns the bleve index path for a wiki.
func GetIndexPath(wikiName string) (string, error) {
	wikiPath, err := GetWikiPath(wikiName)
	if err != nil {
		return "", err
	}
	return filepath.Join(wikiPath, "index.bleve"), nil
}

// ConfigExists reports whether glow.yaml exists on disk.
func ConfigExists() bool {
	_, err := os.Stat(GetConfigPath())
	return err == nil
}

// DiscoveredWiki represents a wiki found in the data directory without config.
type DiscoveredWiki struct {
	Name    string
	Path    string
	Backend BackendType
}

// DiscoverWikis scans the data directory for wiki directories not in config.
// Detects backend by checking for articles.db (sqlite) or articles/ dir (files).
func DiscoverWikis() ([]DiscoveredWiki, error) {
	base, err := GetWikiBasePath()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(base)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	var found []DiscoveredWiki
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if _, registered := cfg.Wikis[name]; registered {
			continue
		}
		if !validWikiName.MatchString(name) {
			continue
		}

		dir := filepath.Join(base, name)
		backend := detectBackend(dir)
		if backend == "" {
			continue // empty or unrecognized directory
		}

		found = append(found, DiscoveredWiki{
			Name:    name,
			Path:    dir,
			Backend: backend,
		})
	}
	return found, nil
}

// detectBackend checks a wiki directory for known artifacts.
func detectBackend(dir string) BackendType {
	if _, err := os.Stat(filepath.Join(dir, "articles.db")); err == nil {
		return BackendSQLite
	}
	if info, err := os.Stat(filepath.Join(dir, "articles")); err == nil && info.IsDir() {
		return BackendFiles
	}
	return ""
}
