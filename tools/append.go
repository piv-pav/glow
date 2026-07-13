package tools

import (
	"fmt"

	"codeberg.org/pivpav/glow/internal/article"
	"github.com/spf13/cobra"
)

var (
	appendSection string
	appendContent string
	appendStdin   bool
	appendTags    []string
	appendUntags  []string
)

var appendCmd = &cobra.Command{
	Use:   "append [article-name]",
	Short: "Append content to an article",
	Long:  `Append content to an article or specific section.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runAppend,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		hasContent := appendStdin || appendContent != ""
		hasTags := len(appendTags) > 0 || len(appendUntags) > 0
		if !hasContent && !hasTags {
			return fmt.Errorf("must specify one of: --content, --stdin, --tag, or --untag")
		}
		return nil
	},
}

func init() {
	appendCmd.Flags().StringVar(&appendSection, "section", "", "Append to specific section by heading")
	appendCmd.Flags().StringVar(&appendContent, "content", "", "Content to append")
	appendCmd.Flags().BoolVar(&appendStdin, "stdin", false, "Read content from stdin")
	appendCmd.Flags().StringArrayVar(&appendTags, "tag", []string{}, "Add tag (comma-separated or repeated: --tag go --tag cli)")
	appendCmd.Flags().StringArrayVar(&appendUntags, "untag", []string{}, "Remove tag (comma-separated or repeated: --untag go --untag cli)")
}

func runAppend(cmd *cobra.Command, args []string) error {
	name := args[0]
	wikiName := wikiNameFrom(cmd)

	content, err := readContent(appendStdin, appendContent)
	if err != nil {
		return err
	}

	msg := fmt.Sprintf("Appended to article: %s", name)
	if appendSection != "" {
		msg = fmt.Sprintf("Appended to section %q in article: %s", appendSection, name)
	}

	err = modifyArticle(wikiName, name, func(art *article.Article) error {
		if content != "" {
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
		}
		if len(appendTags) > 0 {
			art.AddTags(appendTags...)
		}
		if len(appendUntags) > 0 {
			art.RemoveTags(appendUntags...)
		}
		return nil
	})
	if err != nil {
		return err
	}
	fmt.Println(msg)
	return nil
}
