package main

import (
	"fmt"
	"os"

	"git.netra.pivpav.com/public/glow/internal/config"
	"git.netra.pivpav.com/public/glow/tools"
	"github.com/spf13/cobra"
)

var (
	wikiName string
	Version  = "dev"
)

var rootCmd = &cobra.Command{
	Use:     "glow",
	Short:   "GLOW - Go LLM-Oriented Wiki",
	Long:    `A simple CLI tool providing wiki-like access to markdown articles with full-text search and metadata management.`,
	Version: Version,
}

func main() {
	rootCmd.PersistentFlags().StringVarP(&wikiName, "wiki", "w", "default", "Wiki name to use")

	tools.RegisterCommands(rootCmd)

	if err := config.EnsureWikiExists("default"); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to create default wiki: %v\n", err)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
