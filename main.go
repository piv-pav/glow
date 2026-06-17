package main

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

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
	PersistentPreRunE: nil,
}

func main() {
	rootCmd.PersistentFlags().StringVarP(&wikiName, "wiki", "w", "default", "Wiki name to use")

	tools.RegisterCommands(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

