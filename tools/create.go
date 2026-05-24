package tools

import (
	"fmt"
	"os"
	"strings"

	"git.netra.pivpav.com/public/glow/internal/article"
	"git.netra.pivpav.com/public/glow/internal/index"
	"git.netra.pivpav.com/public/glow/internal/storage"
	"github.com/spf13/cobra"
)

var (
	createMeta    []string
	createContent string
	createStdin   bool
)

var createCmd = &cobra.Command{
	Use:   "create [article-name]",
	Short: "Create a new article",
	Long:  `Create a new article with optional metadata. Article name can include folders (e.g., folder/article).`,
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
	createCmd.Flags().StringArrayVar(&createMeta, "meta", []string{}, "Metadata in key:value format (can be repeated)")
	createCmd.Flags().StringVar(&createContent, "content", "", "Article content")
	createCmd.Flags().BoolVar(&createStdin, "stdin", false, "Read content from stdin")
}

func runCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	wikiName := wikiNameFrom(cmd)

	store := storage.New(wikiName)

	art := article.New("")

	for _, meta := range createMeta {
		parts := strings.SplitN(meta, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid metadata format: %s (expected key:value)", meta)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if strings.Contains(value, ",") {
			values := strings.Split(value, ",")
			for i := range values {
				values[i] = strings.TrimSpace(values[i])
			}
			if err := art.AddMetadata(key, values...); err != nil {
				return err
			}
		} else {
			art.SetMetadata(key, value)
		}
	}

	var content string
	if createStdin {
		data, err := os.ReadFile("/dev/stdin")
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}
		content = string(data)
	} else if createContent != "" {
		var err error
		content, err = unescapeContent(createContent)
		if err != nil {
			return err
		}
	}
	art.Content = content

	if err := store.Create(name, art); err != nil {
		return err
	}

	return withIndex(wikiName, func(idx *index.Index) error {
		if err := idx.IndexArticle(name, art); err != nil {
			return fmt.Errorf("failed to index article: %w", err)
		}
		fmt.Printf("Created article: %s\n", name)
		return nil
	})
}
