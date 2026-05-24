package main

import (
	"fmt"

	"github.com/pavelpivovarov/glow/internal/storage"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all articles",
	Long:  `List all articles in the wiki, showing their paths.`,
	Args:  cobra.NoArgs,
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	store := storage.New(wikiName)

	articles, err := store.List()
	if err != nil {
		return err
	}

	if len(articles) == 0 {
		fmt.Printf("No articles in wiki '%s'\n", wikiName)
		return nil
	}

	fmt.Printf("Articles in wiki '%s' (%d):\n\n", wikiName, len(articles))
	for _, article := range articles {
		fmt.Printf("  %s\n", article)
	}

	return nil
}
