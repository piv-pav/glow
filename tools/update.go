package tools

import (
	"fmt"

	"github.com/piv-pav/glow/internal/article"
	"github.com/spf13/cobra"
)

var (
	updateSection string
	updateContent string
	updateStdin   bool
	updateDiff    bool
	updateTags    []string
	updateUntags  []string
)

var updateCmd = &cobra.Command{
	Use:   "update [article-name]",
	Short: "Update an existing article",
	Long: `Update article content or specific section. Use --content or pipe via --stdin. Use --tag/--untag to manage tags.

With --diff, the diff is read from STDIN as one or more SEARCH/REPLACE blocks
(applied to the whole article), the way most AI tools emit text edits:

  <<<<<<< SEARCH
  exact existing text
  =======
  replacement text
  >>>>>>> REPLACE

--diff is its own input mode and cannot be combined with --content or --stdin.
Use --section with --diff to scope the blocks to a single section.`,
	Args: cobra.ExactArgs(1),
	RunE: runUpdate,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		hasContent := updateStdin || updateContent != ""
		hasTags := len(updateTags) > 0 || len(updateUntags) > 0
		if updateDiff {
			if hasContent {
				return fmt.Errorf("--diff cannot be combined with --content or --stdin (it reads the diff from STDIN)")
			}
			return nil
		}
		if !hasContent && !hasTags {
			return fmt.Errorf("must specify --content, --stdin, --diff, --tag, or --untag")
		}
		return nil
	},
}

func init() {
	updateCmd.Flags().StringVar(&updateSection, "section", "", "Update only specific section by heading")
	updateCmd.Flags().StringVar(&updateContent, "content", "", "New content")
	updateCmd.Flags().BoolVar(&updateStdin, "stdin", false, "Read content from stdin")
	updateCmd.Flags().BoolVar(&updateDiff, "diff", false, "Read SEARCH/REPLACE diff blocks from STDIN (own input mode)")
	updateCmd.Flags().StringArrayVar(&updateTags, "tag", []string{}, "Add tag (comma-separated or repeated: --tag go --tag cli)")
	updateCmd.Flags().StringArrayVar(&updateUntags, "untag", []string{}, "Remove tag (comma-separated or repeated: --untag go --untag cli)")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	name := args[0]
	wikiName := wikiNameFrom(cmd)

	var newContent string
	if updateDiff {
		var err error
		newContent, err = readContent(true, "")
		if err != nil {
			return err
		}
	} else if updateStdin || updateContent != "" {
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

	appliedBlocks := 0
	err := modifyArticle(wikiName, name, func(art *article.Article) error {
		switch {
		case updateDiff && updateSection != "":
			n, err := art.ApplyDiffToSection(updateSection, newContent)
			if err != nil {
				return err
			}
			appliedBlocks = n
		case updateDiff:
			result, n, err := article.ApplyDiff(art.Content, newContent)
			if err != nil {
				return err
			}
			art.Content = result
			appliedBlocks = n
		case newContent != "" && updateSection != "":
			if err := art.UpdateSection(updateSection, newContent); err != nil {
				return err
			}
		case newContent != "":
			art.Content = newContent
		}
		if len(updateTags) > 0 {
			art.AddTags(updateTags...)
		}
		if len(updateUntags) > 0 {
			art.RemoveTags(updateUntags...)
		}
		return nil
	})
	if err != nil {
		return err
	}
	if updateDiff {
		if updateSection != "" {
			msg = fmt.Sprintf("Applied %d diff block(s) to section %q in article: %s", appliedBlocks, updateSection, name)
		} else {
			msg = fmt.Sprintf("Applied %d diff block(s) to article: %s", appliedBlocks, name)
		}
	}
	fmt.Println(msg)
	return nil
}
