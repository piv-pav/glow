package tools

import (
	"fmt"

	"codeberg.org/pivpav/glow/internal/index"
	"codeberg.org/pivpav/glow/internal/storage"
	"github.com/spf13/cobra"
)

var (
	appendSection string
	appendContent string
	appendStdin   bool
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
	appendCmd.Flags().StringVar(&appendSection, "section", "", "Append to specific section by heading")
	appendCmd.Flags().StringVar(&appendContent, "content", "", "Content to append")
	appendCmd.Flags().BoolVar(&appendStdin, "stdin", false, "Read content from stdin")
}

func runAppend(cmd *cobra.Command, args []string) error {
	name := args[0]
	wikiName := wikiNameFrom(cmd)

	content, err := readContent(appendStdin, appendContent)
	if err != nil {
		return err
	}

	store := storage.New(wikiName)
	art, err := store.Read(name)
	if err != nil {
		return err
	}

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

	if err := store.Update(name, art); err != nil {
		return err
	}

	return withIndex(wikiName, func(idx *index.Index) error {
		if err := idx.UpdateArticle(name, art); err != nil {
			return fmt.Errorf("failed to update index: %w", err)
		}
		fmt.Printf("Appended to article: %s\n", name)
		if appendSection != "" {
			fmt.Printf("Section: %s\n", appendSection)
		}
		return nil
	})
}
