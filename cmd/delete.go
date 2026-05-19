package cmd

import (
	"fmt"

	"github.com/pavelpivovarov/glow/internal/index"
	"github.com/pavelpivovarov/glow/internal/storage"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [article-name]",
	Short: "Delete an article",
	Long:  `Delete an article and remove it from the index.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runDelete,
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Create storage and index
	store := storage.New(wikiName)
	idx, err := index.New(wikiName)
	if err != nil {
		return fmt.Errorf("failed to open index: %w", err)
	}
	defer idx.Close()

	// Delete from storage
	if err := store.Delete(name); err != nil {
		return err
	}

	// Delete from index
	if err := idx.DeleteArticle(name); err != nil {
		return fmt.Errorf("failed to remove from index: %w", err)
	}

	fmt.Printf("Deleted article: %s\n", name)
	return nil
}
