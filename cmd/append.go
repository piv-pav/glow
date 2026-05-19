package cmd

import (
	"fmt"

	"github.com/pavelpivovarov/glow/internal/index"
	"github.com/pavelpivovarov/glow/internal/storage"
	"github.com/spf13/cobra"
)

var (
	appendSection string
)

var appendCmd = &cobra.Command{
	Use:   "append [article-name] [content]",
	Short: "Append content to an article",
	Long:  `Append content to an article or specific section.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runAppend,
}

func init() {
	rootCmd.AddCommand(appendCmd)
	appendCmd.Flags().StringVar(&appendSection, "section", "", "Append to specific section by heading")
}

func runAppend(cmd *cobra.Command, args []string) error {
	name := args[0]
	content := args[1]

	// Create storage and index
	store := storage.New(wikiName)
	idx, err := index.New(wikiName)
	if err != nil {
		return fmt.Errorf("failed to open index: %w", err)
	}
	defer idx.Close()

	// Read existing article
	art, err := store.Read(name)
	if err != nil {
		return err
	}

	// Append content
	if appendSection != "" {
		if err := art.AppendToSection(appendSection, content); err != nil {
			return err
		}
	} else {
		if art.Content != "" && art.Content[len(art.Content)-1] != '\n' {
			art.Content += "\n"
		}
		art.Content += "\n" + content
	}

	// Save
	if err := store.Update(name, art); err != nil {
		return err
	}

	// Update index
	if err := idx.UpdateArticle(name, art); err != nil {
		return fmt.Errorf("failed to update index: %w", err)
	}

	fmt.Printf("Appended to article: %s\n", name)
	if appendSection != "" {
		fmt.Printf("Section: %s\n", appendSection)
	}

	return nil
}
