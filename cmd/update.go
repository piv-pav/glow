package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/pavelpivovarov/glow/internal/index"
	"github.com/pavelpivovarov/glow/internal/storage"
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
	Long:  `Update article content or specific section. Opens editor for content modification.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runUpdate,
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().StringVar(&updateSection, "section", "", "Update only specific section by heading")
	updateCmd.Flags().StringVar(&updateContent, "content", "", "New content (skips editor)")
	updateCmd.Flags().BoolVar(&updateStdin, "stdin", false, "Read content from stdin (skips editor)")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	name := args[0]

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

	var initialContent string
	if updateSection != "" {
		// Get section content
		section := art.FindSection(updateSection)
		if section == nil {
			return fmt.Errorf("section not found: %s", updateSection)
		}
		// Get content without heading line
		lines := splitLines(section.Content)
		if len(lines) > 1 {
			initialContent = joinLines(lines[1:])
		}
	} else {
		initialContent = art.Content
	}

	// Get content from flag, stdin, or editor
	var newContent string
	if updateStdin {
		// Read from stdin
		data, err := os.ReadFile("/dev/stdin")
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}
		newContent = string(data)
	} else if updateContent != "" {
		// Use content from flag
		newContent = updateContent
	} else {
		// Open editor
		var err error
		newContent, err = openEditor(initialContent)
		if err != nil {
			return fmt.Errorf("failed to open editor: %w", err)
		}
	}

	// Update article
	if updateSection != "" {
		if err := art.UpdateSection(updateSection, newContent); err != nil {
			return err
		}
	} else {
		art.Content = newContent
	}

	// Save
	if err := store.Update(name, art); err != nil {
		return err
	}

	// Update index
	if err := idx.UpdateArticle(name, art); err != nil {
		return fmt.Errorf("failed to update index: %w", err)
	}

	fmt.Printf("Updated article: %s\n", name)
	if updateSection != "" {
		fmt.Printf("Section: %s\n", updateSection)
	}

	return nil
}

func splitLines(s string) []string {
	return strings.Split(s, "\n")
}

func joinLines(lines []string) string {
	return strings.Join(lines, "\n")
}
