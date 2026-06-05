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

	msg := fmt.Sprintf("Appended to article: %s", name)
	if appendSection != "" {
		msg = fmt.Sprintf("Appended to section %q in article: %s", appendSection, name)
	}

	return modifyArticle(wikiName, name, func(art *article.Article) error {
		if appendSection != "" {
			return art.AppendToSection(appendSection, content)
		}
		if art.Content != "" && art.Content[len(art.Content)-1] != '\n' {
			art.Content += "\n"
		}
		art.Content += "\n" + content
		return nil
	}, msg)
}
