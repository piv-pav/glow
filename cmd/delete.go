package main

import (
	"fmt"

	"github.com/pavelpivovarov/glow/internal/index"
	"github.com/pavelpivovarov/glow/internal/storage"
	"github.com/spf13/cobra"
)

var (
	deleteSection string
)

var deleteCmd = &cobra.Command{
	Use:   "delete [article-name]",
	Short: "Delete an article or section",
	Long:  `Delete an article and remove it from the index, or delete a specific section from an article.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runDelete,
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().StringVar(&deleteSection, "section", "", "Delete only specific section by heading")
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

	if deleteSection != "" {
		// Delete specific section
		art, err := store.Read(name)
		if err != nil {
			return err
		}

		if err := art.DeleteSection(deleteSection); err != nil {
			return err
		}

		if err := store.Update(name, art); err != nil {
			return err
		}

		if err := idx.UpdateArticle(name, art); err != nil {
			return fmt.Errorf("failed to update index: %w", err)
		}

		fmt.Printf("Deleted section: %s from article: %s\n", deleteSection, name)
		return nil
	}

	// Delete entire article
	if err := store.Delete(name); err != nil {
		return err
	}

	if err := idx.DeleteArticle(name); err != nil {
		return fmt.Errorf("failed to remove from index: %w", err)
	}

	fmt.Printf("Deleted article: %s\n", name)
	return nil
}
