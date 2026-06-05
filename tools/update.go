package tools

import (
	"fmt"

	"codeberg.org/pivpav/glow/internal/index"
	"codeberg.org/pivpav/glow/internal/storage"
	"github.com/spf13/cobra"
)

var (
	updateSection string
	updateContent string
	updateStdin   bool
)

var updateCmd = &cobra.Command{
	Use:   "update [article-name]",
	Short: "Update an existing article",
	Long:  `Update article content or specific section. Use --content or pipe via --stdin. For metadata changes use 'glow meta'.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runUpdate,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if !updateStdin && updateContent == "" {
			return fmt.Errorf("must specify one of: --content or --stdin")
		}
		return nil
	},
}

func init() {
	updateCmd.Flags().StringVar(&updateSection, "section", "", "Update only specific section by heading")
	updateCmd.Flags().StringVar(&updateContent, "content", "", "New content")
	updateCmd.Flags().BoolVar(&updateStdin, "stdin", false, "Read content from stdin")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	name := args[0]
	wikiName := wikiNameFrom(cmd)

	newContent, err := readContent(updateStdin, updateContent)
	if err != nil {
		return err
	}

	store := storage.New(wikiName)
	art, err := store.Read(name)
	if err != nil {
		return err
	}

	if updateSection != "" {
		if err := art.UpdateSection(updateSection, newContent); err != nil {
			return err
		}
	} else {
		art.Content = newContent
	}

	if err := store.Update(name, art); err != nil {
		return err
	}

	return withIndex(wikiName, func(idx *index.Index) error {
		if err := idx.UpdateArticle(name, art); err != nil {
			return fmt.Errorf("failed to update index: %w", err)
		}
		fmt.Printf("Updated article: %s\n", name)
		if updateSection != "" {
			fmt.Printf("Section: %s\n", updateSection)
		}
		return nil
	})
}
