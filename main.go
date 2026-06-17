package main

import (
	_ "embed"
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
func maybeDiscoverWikis(cmd *cobra.Command, args []string) error {
	if config.ConfigExists() {
		return nil
	}

	wikis, err := config.DiscoverWikis()
	if err != nil || len(wikis) == 0 {
		return nil
	}

	fmt.Fprintf(os.Stderr, "Detected %d existing wiki(s), adding to config:\n", len(wikis))
	for _, w := range wikis {
		fmt.Fprintf(os.Stderr, "  + %s [%s]\n", w.Name, w.Backend)
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
	fmt.Fprintf(os.Stderr, "Saved %s\n", config.GetConfigPath())
	return nil
}
