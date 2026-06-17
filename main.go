package main

import (
	_ "embed"
	"bufio"
	"fmt"
	"os"
	"strings"

	"codeberg.org/pivpav/glow/internal/config"
	"codeberg.org/pivpav/glow/tools"
	"github.com/spf13/cobra"
)

//go:embed VERSION
var versionFile string

var (
	wikiName string
	Version  = "dev"
)

func init() {
	v := strings.TrimSpace(versionFile)
	if Version == "dev" && v != "" {
		Version = "v" + v
	}
	rootCmd.Version = Version
}

var rootCmd = &cobra.Command{
	Use:     "glow",
	Short:   "GLOW - Go LLM-Oriented Wiki",
	Long:    `A simple CLI tool providing wiki-like access to markdown articles with full-text search and tag management.`,
	PersistentPreRunE: maybeDiscoverWikis,
}

func main() {
	rootCmd.PersistentFlags().StringVarP(&wikiName, "wiki", "w", "default", "Wiki name to use")

	tools.RegisterCommands(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// maybeDiscoverWikis checks for existing wiki data when no config file exists.
// If wikis are found in the data directory, registers them automatically.
// In interactive mode (TTY), asks for confirmation first.
func maybeDiscoverWikis(cmd *cobra.Command, args []string) error {
	if config.ConfigExists() {
		return nil
	}

	wikis, err := config.DiscoverWikis()
	if err != nil || len(wikis) == 0 {
		return nil
	}

	// Check if interactive (TTY)
	interactive := false
	if fi, _ := os.Stdin.Stat(); fi != nil && fi.Mode()&os.ModeCharDevice != 0 {
		interactive = true
	}

	if interactive {
		fmt.Fprintf(os.Stderr, "Found %d existing wiki(s) in data directory:\n", len(wikis))
		for _, w := range wikis {
			fmt.Fprintf(os.Stderr, "  %s [%s] %s\n", w.Name, w.Backend, w.Path)
		}

		r := bufio.NewReader(os.Stdin)
		fmt.Fprint(os.Stderr, "\nAdd them to config? [Y/n]: ")
		ans, _ := r.ReadString('\n')
		ans = strings.ToLower(strings.TrimSpace(ans))
		if ans != "" && ans != "y" && ans != "yes" {
			return nil
		}
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}
	for _, w := range wikis {
		cfg.Wikis[w.Name] = &config.WikiConfig{Backend: w.Backend}
	}
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Registered %d wiki(s) in %s\n", len(wikis), config.GetConfigPath())
	if interactive {
		fmt.Fprintln(os.Stderr)
	}
	return nil
}
