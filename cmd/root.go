package main

import (
	"fmt"
	"os"

	"github.com/pavelpivovarov/glow/internal/config"
	"github.com/spf13/cobra"
)

var (
	wikiName string
	Version  = "dev" // Set via ldflags: -X 'github.com/pavelpivovarov/glow/cmd.Version=v1.0.0'
)

var rootCmd = &cobra.Command{
	Use:     "wiki",
	Short:   "GLOW - Go LLM-Oriented Wiki",
	Long:    `A simple CLI tool providing wiki-like access to markdown articles with full-text search and metadata management.`,
	Version: Version,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&wikiName, "wiki", "w", "default", "Wiki name to use")

	// Ensure default wiki exists
	if err := config.EnsureWikiExists("default"); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to create default wiki: %v\n", err)
	}
}

func main() {
	Execute()
}
