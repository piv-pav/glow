package tools

import (
	"fmt"

	"codeberg.org/pivpav/glow/internal/article"
	"codeberg.org/pivpav/glow/internal/index"
	"codeberg.org/pivpav/glow/internal/storage"
	"github.com/spf13/cobra"
)

var (
	createTags    []string
	createContent string
	createStdin   bool
)

var createCmd = &cobra.Command{
	Use:   "create [article-name]",
	Short: "Create a new article",
	Long:  `Create a new article with optional tags. Article name can include folders (e.g., folder/article).`,
	Args:  cobra.ExactArgs(1),
	RunE:  runCreate,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if !createStdin && createContent == "" {
			return fmt.Errorf("must specify one of: --content or --stdin")
		}
		return nil
	},
}

func init() {
	createCmd.Flags().StringArrayVar(&createTags, "tag", []string{}, "Add tag (comma-separated or repeated: --tag go --tag cli)")
	createCmd.Flags().StringVar(&createContent, "content", "", "Article content")
	createCmd.Flags().BoolVar(&createStdin, "stdin", false, "Read content from stdin")
}

func runCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	wikiName := wikiNameFrom(cmd)

	content, err := readContent(createStdin, createContent)
	if err != nil {
		return err
	}

	art := article.New(content)
	if len(createTags) > 0 {
		art.AddTags(createTags...)
	}

	return withStore(wikiName, func(store storage.Store) error {
		if err := store.Create(name, art); err != nil {
			return err
		}
		if err := withIndex(wikiName, func(idx *index.Index) error {
			return idx.IndexArticle(name, art)
		}); err != nil {
			return fmt.Errorf("failed to index article: %w", err)
		}
		fmt.Printf("Created article: %s\n", name)
		return nil
	})
}
