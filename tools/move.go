package tools

import (
	"fmt"

	"git.netra.pivpav.com/public/glow/internal/index"
	"git.netra.pivpav.com/public/glow/internal/storage"
	"github.com/spf13/cobra"
)

var moveCmd = &cobra.Command{
	Use:   "move [old-name] [new-name]",
	Short: "Move or rename an article",
	Long:  `Move or rename an article. Can move to different folder by specifying path in new name.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runMove,
}

func runMove(cmd *cobra.Command, args []string) error {
	oldName := args[0]
	newName := args[1]
	wikiName := wikiNameFrom(cmd)

	store := storage.New(wikiName)

	art, err := store.Read(oldName)
	if err != nil {
		return err
	}

	if err := store.Move(oldName, newName); err != nil {
		return err
	}

	return withIndex(wikiName, func(idx *index.Index) error {
		if err := idx.DeleteArticle(oldName); err != nil {
			return fmt.Errorf("failed to remove old entry from index: %w", err)
		}

		art, err = store.Read(newName)
		if err != nil {
			return err
		}

		if err := idx.IndexArticle(newName, art); err != nil {
			return fmt.Errorf("failed to index article with new name: %w", err)
		}

		fmt.Printf("Moved article: %s -> %s\n", oldName, newName)
		return nil
	})
}
