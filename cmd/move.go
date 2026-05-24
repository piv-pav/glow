package main

import (
	"fmt"

	"github.com/pavelpivovarov/glow/internal/index"
	"github.com/pavelpivovarov/glow/internal/storage"
	"github.com/spf13/cobra"
)

var moveCmd = &cobra.Command{
	Use:   "move [old-name] [new-name]",
	Short: "Move or rename an article",
	Long:  `Move or rename an article. Can move to different folder by specifying path in new name.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runMove,
}

func init() {
	rootCmd.AddCommand(moveCmd)
}

func runMove(cmd *cobra.Command, args []string) error {
	oldName := args[0]
	newName := args[1]

	// Create storage and index
	store := storage.New(wikiName)
	idx, err := index.New(wikiName)
	if err != nil {
		return fmt.Errorf("failed to open index: %w", err)
	}
	defer idx.Close()

	// Read article before move (to re-index with new name)
	art, err := store.Read(oldName)
	if err != nil {
		return err
	}

	// Move in storage
	if err := store.Move(oldName, newName); err != nil {
		return err
	}

	// Delete old from index
	if err := idx.DeleteArticle(oldName); err != nil {
		return fmt.Errorf("failed to remove old entry from index: %w", err)
	}

	// Re-read with new path metadata
	art, err = store.Read(newName)
	if err != nil {
		return err
	}

	// Index with new name
	if err := idx.IndexArticle(newName, art); err != nil {
		return fmt.Errorf("failed to index article with new name: %w", err)
	}

	fmt.Printf("Moved article: %s -> %s\n", oldName, newName)
	return nil
}
