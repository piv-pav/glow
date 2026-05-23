package cmd

import (
	"fmt"
	"os"

	"github.com/pavelpivovarov/glow/internal/index"
	"github.com/pavelpivovarov/glow/internal/storage"
	"github.com/spf13/cobra"
)

var (
	appendSection  string
	appendContent  string
	appendStdin    bool
)

var appendCmd = &cobra.Command{
	Use:   "append [article-name]",
	Short: "Append content to an article",
	Long:  `Append content to an article or specific section.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runAppend,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if !appendStdin && appendContent == "" {
			return fmt.Errorf("must specify one of: --content or --stdin")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(appendCmd)
	appendCmd.Flags().StringVar(&appendSection, "section", "", "Append to specific section by heading")
	appendCmd.Flags().StringVar(&appendContent, "content", "", "Content to append")
	appendCmd.Flags().BoolVar(&appendStdin, "stdin", false, "Read content from stdin")
}

func runAppend(cmd *cobra.Command, args []string) error {
	name := args[0]
	var content string

	if appendStdin {
		data, err := os.ReadFile("/dev/stdin")
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}
		content = string(data)
	} else {
		content = appendContent
	}

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
