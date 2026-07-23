package tools

import (
	"fmt"

	"github.com/piv-pav/glow/internal/article"
	"github.com/piv-pav/glow/internal/storage"
	"github.com/spf13/cobra"
)

var deleteSection string

var deleteCmd = &cobra.Command{
	Use:   "delete [article-name]",
	Short: "Delete an article or section",
	Long:  `Delete an article and remove it from the index, or delete a specific section from an article.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runDelete,
}

func init() {
	deleteCmd.Flags().StringVar(&deleteSection, "section", "", "Delete only specific section by heading")
}

func runDelete(cmd *cobra.Command, args []string) error {
	name := args[0]
	wikiName := wikiNameFrom(cmd)

	if deleteSection != "" {
		msg := fmt.Sprintf("Deleted section: %s from article: %s", deleteSection, name)
		err := modifyArticle(wikiName, name, func(art *article.Article) error {
			return art.DeleteSection(deleteSection)
		})
		if err != nil {
			return err
		}
		fmt.Println(msg)
		return nil
	}

	return withStore(wikiName, func(store storage.Store) error {
		if err := store.Delete(name); err != nil {
			return err
		}
		fmt.Printf("Deleted article: %s\n", name)
		return nil
	})
}
