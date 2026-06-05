package tools

import (
	"fmt"

	"codeberg.org/pivpav/glow/internal/article"
	"github.com/spf13/cobra"
)

var (
	updateSection string
	updateContent string
	updateStdin   bool
	updateTags    []string
	updateUntags  []string
)

var updateCmd = &cobra.Command{
	Use:   "update [article-name]",
	Short: "Update an existing article",
	Long:  `Update article content or specific section. Use --content or pipe via --stdin. Use --tag/--untag to manage tags.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runUpdate,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		hasContent := updateStdin || updateContent != ""
		hasTags := len(updateTags) > 0 || len(updateUntags) > 0
		if !hasContent && !hasTags {
			return fmt.Errorf("must specify --content, --stdin, --tag, or --untag")
		}
		return nil
	},
}

func init() {
	updateCmd.Flags().StringVar(&updateSection, "section", "", "Update only specific section by heading")
	updateCmd.Flags().StringVar(&updateContent, "content", "", "New content")
	updateCmd.Flags().BoolVar(&updateStdin, "stdin", false, "Read content from stdin")
	updateCmd.Flags().StringArrayVar(&updateTags, "tag", []string{}, "Add tag (comma-separated or repeated: --tag go --tag cli)")
	updateCmd.Flags().StringArrayVar(&updateUntags, "untag", []string{}, "Remove tag (comma-separated or repeated: --untag go --untag cli)")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	name := args[0]
	wikiName := wikiNameFrom(cmd)

	var newContent string
	if updateStdin || updateContent != "" {
		var err error
		newContent, err = readContent(updateStdin, updateContent)
		if err != nil {
			return err
		}
	}

	msg := fmt.Sprintf("Updated article: %s", name)
	if updateSection != "" {
		msg = fmt.Sprintf("Updated section %q in article: %s", updateSection, name)
	}

	return modifyArticle(wikiName, name, func(art *article.Article) error {
		if newContent != "" {
			if updateSection != "" {
				if err := art.UpdateSection(updateSection, newContent); err != nil {
					return err
				}
			} else {
				art.Content = newContent
			}
		}
		if len(updateTags) > 0 {
			art.AddTags(updateTags...)
		}
		if len(updateUntags) > 0 {
			art.RemoveTags(updateUntags...)
		}
		return nil
	}, msg)
}
