package tools

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"codeberg.org/pivpav/glow/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var wikiInitCmd = &cobra.Command{
	Use:   "init [name]",
	Short: "Create a new wiki (interactive)",
	Long:  `Create a new wiki and configure its storage backend (sqlite or rqlite).`,
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

var wikiDeleteCmd = &cobra.Command{
	Use:   "wiki-delete [name]",
	Short: "Delete a wiki",
	Long:  `Remove a wiki from config. For sqlite backend, also deletes local data.`,
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

	fmt.Print("Storage backend [sqlite/rqlite] (default: sqlite): ")
	backendStr, _ := r.ReadString('\n')
	backendStr = strings.TrimSpace(strings.ToLower(backendStr))

	var cfg config.WikiConfig
	switch backendStr {
	case "rqlite":
		cfg.Backend = config.BackendRqlite
		cfg.Rqlite = promptRqlite(r)
	default:
		cfg.Backend = config.BackendSQLite
	}

	if err := config.CreateWiki(name, &cfg); err != nil {
		return err
	}

	fmt.Printf("Created wiki: %s (backend: %s)\n", name, cfg.Backend)
	switch cfg.Backend {
	case config.BackendRqlite:
		fmt.Printf("Location: %s\n", cfg.Rqlite.URL)
	default:
		wikiPath, _ := config.GetWikiPath(name)
		fmt.Printf("Location: %s\n", wikiPath)
	}
	return nil
}

// readPassword prints a prompt and reads a line without echo.
// Falls back to normal stdin read when not a terminal.
func readPassword(r *bufio.Reader, prompt string) string {
	fmt.Print(prompt)
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		s, _ := r.ReadString('\n')
		return strings.TrimSpace(s)
	}
	b, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func promptRqlite(r *bufio.Reader) *config.RqliteConfig {
	ask := func(prompt string) string {
		fmt.Print(prompt)
		s, _ := r.ReadString('\n')
		return strings.TrimSpace(s)
	}
	url := ask("  URL (e.g. http://localhost:4001): ")
	if url == "" {
		url = "http://localhost:4001"
	}
	return &config.RqliteConfig{
		URL:      url,
		User:     ask("  User (optional): "),
		Password: readPassword(r, "  Password (optional): "),
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
		cfg, _ := config.GetWikiConfig(wiki)
		backend := "sqlite"
		location := ""
		if cfg != nil {
			backend = string(cfg.Backend)
			if cfg.Backend == config.BackendRqlite && cfg.Rqlite != nil {
				location = cfg.Rqlite.URL
			}
		}
		if location == "" {
			location, _ = config.GetWikiPath(wiki)
		}
		fmt.Printf("  %s  [%s]\n    %s\n", wiki, backend, location)
	}
	return nil
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

	if cfg.Backend == config.BackendSQLite {
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
