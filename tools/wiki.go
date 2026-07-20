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

// wikiCreateCmd flags
var (
	wikiCreateBackend       string
	wikiCreateURL           string
	wikiCreateUser          string
	wikiCreatePassword      string
	wikiCreatePasswordStdin bool
	wikiCreateLevel         string
	wikiCreateInteractive   bool
)

var wikiCreateCmd = &cobra.Command{
	Use:   "wiki-create <name>",
	Short: "Create a new wiki",
	Long: `Create a new wiki.

Non-interactive (-b / --backend required):
  glow wiki-create <name> -b sqlite
  glow wiki-create <name> -b rqlite --url http://localhost:4001
  glow wiki-create <name> -b rqlite --url http://localhost:4001 --user foo --password bar
  glow wiki-create <name> -b rqlite --url http://localhost:4001 --user foo --password-stdin

Interactive (-i / --interactive prompts for backend and connection details):
  glow wiki-create <name> -i`,
	Args: cobra.ExactArgs(1),
	RunE: runWikiCreate,
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

func init() {
	wikiCreateCmd.Flags().StringVarP(&wikiCreateBackend, "backend", "b", "", "Storage backend: sqlite or rqlite (required for non-interactive)")
	wikiCreateCmd.Flags().StringVar(&wikiCreateURL, "url", "", "rqlite URL (e.g. http://localhost:4001)")
	wikiCreateCmd.Flags().StringVar(&wikiCreateUser, "user", "", "rqlite username (optional)")
	wikiCreateCmd.Flags().StringVar(&wikiCreatePassword, "password", "", "rqlite password (optional; use --password-stdin to read from stdin)")
	wikiCreateCmd.Flags().BoolVar(&wikiCreatePasswordStdin, "password-stdin", false, "Read rqlite password from stdin")
	wikiCreateCmd.Flags().StringVar(&wikiCreateLevel, "level", "", "rqlite consistency level: none, weak, strong (default: weak)")
	wikiCreateCmd.Flags().BoolVarP(&wikiCreateInteractive, "interactive", "i", false, "Run in interactive mode (prompts for backend and connection details)")
}

func runWikiCreate(cmd *cobra.Command, args []string) error {
	name := args[0]

	if wikiCreateInteractive && wikiCreateBackend != "" {
		return fmt.Errorf("--interactive / -i and --backend / -b are mutually exclusive")
	}
	if !wikiCreateInteractive && wikiCreateBackend == "" {
		return fmt.Errorf("one of --interactive / -i or --backend / -b is required")
	}

	if wikiCreateInteractive {
		return runWikiCreateInteractive(name)
	}

	// Non-interactive path
	var cfg config.WikiConfig
	switch strings.ToLower(wikiCreateBackend) {
	case "sqlite":
		cfg.Backend = config.BackendSQLite
	case "rqlite":
		cfg.Backend = config.BackendRqlite

		password := wikiCreatePassword
		if wikiCreatePasswordStdin {
			if wikiCreatePassword != "" {
				return fmt.Errorf("cannot use both --password and --password-stdin")
			}
			r := bufio.NewReader(os.Stdin)
			line, _ := r.ReadString('\n')
			password = strings.TrimSpace(line)
		}

		cfg.Rqlite = &config.RqliteConfig{
			URL:      wikiCreateURL,
			User:     wikiCreateUser,
			Password: password,
			Level:    wikiCreateLevel,
		}
		if cfg.Rqlite.URL == "" {
			return fmt.Errorf("--url is required for rqlite backend")
		}
	default:
		return fmt.Errorf("unknown backend %q: use sqlite or rqlite", wikiCreateBackend)
	}

	if err := config.CreateWiki(name, &cfg); err != nil {
		return err
	}

	fmt.Printf("Created wiki: %s\n", name)
	switch cfg.Backend {
	case config.BackendRqlite:
		fmt.Printf("Location: %s\n", cfg.Rqlite.URL)
	default:
		wikiPath, _ := config.GetWikiDBPath(name)
		fmt.Printf("Location: %s\n", wikiPath)
	}
	return nil
}

// runWikiCreateInteractive prompts for backend and connection details.
// name is pre-filled from the positional arg; the name prompt is skipped.
func runWikiCreateInteractive(name string) error {
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
		wikiPath, _ := config.GetWikiDBPath(name)
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
			location, _ = config.GetWikiDBPath(wiki)
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
		wikiPath, err := config.GetWikiDBPath(name)
		if err != nil {
			return err
		}
		if err := os.Remove(wikiPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove wiki db: %w", err)
		}
		fmt.Printf("Removed: %s\n", wikiPath)
	}

	if err := config.DeleteWiki(name); err != nil {
		return err
	}
	fmt.Printf("Deleted wiki: %s\n", name)
	return nil
}
