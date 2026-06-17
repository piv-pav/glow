package tools

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"codeberg.org/pivpav/glow/internal/article"
	"codeberg.org/pivpav/glow/internal/config"
	"codeberg.org/pivpav/glow/internal/index"
	"codeberg.org/pivpav/glow/internal/storage"
	"github.com/spf13/cobra"
)

var wikiInitCmd = &cobra.Command{
	Use:   "init [name]",
	Short: "Create a new wiki (interactive)",
	Long:  `Create a new wiki and configure its storage backend (sqlite or pgsql).`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runWikiInit,
}

var wikiCreateCmd = &cobra.Command{
	Use:   "wiki-create [name]",
	Short: "Create a new wiki (sqlite, non-interactive)",
	Long:  `Create a new wiki with sqlite backend. Use 'init' for interactive backend selection.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runWikiCreate,
}

var wikiListCmd = &cobra.Command{
	Use:   "wiki-list",
	Short: "List all wikis",
	Args:  cobra.NoArgs,
	RunE:  runWikiList,
}

var wikiRebuildCmd = &cobra.Command{
	Use:   "rebuild",
	Short: "Rebuild wiki index",
	Long:  `Completely rebuild the wiki index from all articles. Use when index is corrupted.`,
	Args:  cobra.NoArgs,
	RunE:  runWikiRebuild,
}

var wikiDeleteCmd = &cobra.Command{
	Use:   "wiki-delete [name]",
	Short: "Delete a wiki",
	Long:  `Remove a wiki from config. For files/sqlite backends, also deletes local data.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runWikiDelete,
}

// runWikiInit is the interactive wiki creator.
func runWikiInit(cmd *cobra.Command, args []string) error {
	name := "default"
	if len(args) > 0 {
		name = args[0]
	}

	r := bufio.NewReader(os.Stdin)

	fmt.Print("Storage backend [sqlite/pgsql/files] (default: sqlite): ")
	backendStr, _ := r.ReadString('\n')
	backendStr = strings.TrimSpace(strings.ToLower(backendStr))

	var cfg config.WikiConfig
	switch backendStr {
	case "pgsql", "postgres", "postgresql":
		cfg.Backend = config.BackendPgSQL
		cfg.PgSQL = promptPgSQL(r)
		if err := ensureDatabase(r, cfg.PgSQL.DBName, cfg.PgSQL.Host, func() (bool, error) {
			return storage.EnsurePgDatabase(cfg.PgSQL)
		}); err != nil {
			return err
		}
	case "files":
		cfg.Backend = config.BackendFiles
	default:
		cfg.Backend = config.BackendSQLite
	}

	if err := config.CreateWiki(name, &cfg); err != nil {
		return err
	}

	// Initialize Bleve index for files backend
	if err := withIndex(name, func(idx *index.Index) error {
		return nil
	}); err != nil {
		return err
	}

	fmt.Printf("Created wiki: %s (backend: %s)\n", name, cfg.Backend)
	wikiPath, _ := config.GetWikiPath(name)
	fmt.Printf("Location: %s\n", wikiPath)
	return nil
}

// ensureDatabase checks if the target DB exists and offers to create it if not.
func ensureDatabase(r *bufio.Reader, dbname, host string, create func() (bool, error)) error {
	fmt.Printf("  Checking database %q on %s...", dbname, host)
	created, err := create()
	if err != nil {
		fmt.Printf(" failed: %v\n", err)
		fmt.Print("  Continue anyway? [y/N]: ")
		ans, _ := r.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(ans)) != "y" {
			return fmt.Errorf("aborted")
		}
		return nil
	}
	if created {
		fmt.Printf(" created.\n")
	} else {
		fmt.Printf(" exists.\n")
	}
	return nil
}

// promptPgSQL reads PgSQL connection details interactively.
func promptPgSQL(r *bufio.Reader) *config.PgSQLConfig {
	ask := func(prompt string) string {
		fmt.Print(prompt)
		s, _ := r.ReadString('\n')
		return strings.TrimSpace(s)
	}
	host := ask("  Host (default: localhost): ")
	if host == "" {
		host = "localhost"
	}
	return &config.PgSQLConfig{
		Host:     host,
		DBName:   ask("  Database name: "),
		User:     ask("  User: "),
		Password: ask("  Password: "),
	}
}

func runWikiCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	if err := config.CreateWiki(name, &config.WikiConfig{Backend: config.BackendSQLite}); err != nil {
		return err
	}
	fmt.Printf("Created wiki: %s\n", name)
	wikiPath, _ := config.GetWikiPath(name)
	fmt.Printf("Location: %s\n", wikiPath)
	return nil
}

func runWikiList(cmd *cobra.Command, args []string) error {
	wikis, err := config.ListWikis()
	if err != nil {
		return err
	}
	if len(wikis) == 0 {
		fmt.Println("No wikis found")
		return nil
	}
	fmt.Printf("Available wikis (%d):\n\n", len(wikis))
	for _, wiki := range wikis {
		wikiPath, _ := config.GetWikiPath(wiki)
		cfg, _ := config.GetWikiConfig(wiki)
		backend := "sqlite"
		if cfg != nil {
			backend = string(cfg.Backend)
		}
		fmt.Printf("  %s  [%s]\n    %s\n", wiki, backend, wikiPath)
	}
	return nil
}

func runWikiRebuild(cmd *cobra.Command, args []string) error {
	wikiName := wikiNameFrom(cmd)

	cfg, err := config.GetWikiConfig(wikiName)
	if err != nil {
		return err
	}
	if cfg == nil {
		return fmt.Errorf("wiki does not exist: %s", wikiName)
	}
	if cfg.Backend != config.BackendFiles {
		fmt.Printf("Wiki '%s' uses %s backend with native search — no index to rebuild\n", wikiName, cfg.Backend)
		return nil
	}

	fmt.Printf("Rebuilding index for wiki '%s'...\n", wikiName)

	return withStore(wikiName, func(store storage.Store) error {
		articleNames, err := store.List()
		if err != nil {
			return fmt.Errorf("failed to list articles: %w", err)
		}

		articles := make(map[string]*article.Article)
		for _, name := range articleNames {
			art, err := store.Read(name)
			if err != nil {
				fmt.Printf("Warning: failed to read article %s: %v\n", name, err)
				continue
			}
			articles[name] = art
		}

		return withIndex(wikiName, func(idx *index.Index) error {
			if err := idx.Rebuild(articles); err != nil {
				return fmt.Errorf("failed to rebuild index: %w", err)
			}
			fmt.Printf("Successfully rebuilt index for wiki '%s'\n", wikiName)
			fmt.Printf("Indexed %d articles\n", len(articles))
			return nil
		})
	})
}

func runWikiDelete(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := config.GetWikiConfig(name)
	if err != nil {
		return err
	}
	if cfg == nil {
		return fmt.Errorf("wiki not found: %s", name)
	}

	// Remove local data for file-based backends
	if cfg.Backend != config.BackendPgSQL {
		wikiPath, err := config.GetWikiPath(name)
		if err != nil {
			return err
		}
		if err := os.RemoveAll(wikiPath); err != nil {
			return fmt.Errorf("failed to remove wiki data: %w", err)
		}
		fmt.Printf("Removed data: %s\n", wikiPath)
	}

	if err := config.DeleteWiki(name); err != nil {
		return err
	}
	fmt.Printf("Deleted wiki: %s\n", name)
	return nil
}
