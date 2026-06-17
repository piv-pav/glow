package tools

import (
	"fmt"

	"codeberg.org/pivpav/glow/internal/index"
	"codeberg.org/pivpav/glow/internal/storage"
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

	return withStore(wikiName, func(store storage.Store) error {
		if err := store.Move(oldName, newName); err != nil {
			return err
		}
		art, err := store.Read(newName)
		if err != nil {
			return err
		}
		if err := withIndex(wikiName, func(idx *index.Index) error {
			if err := idx.DeleteArticle(oldName); err != nil {
				return err
			}
			return idx.IndexArticle(newName, art)
		}); err != nil {
			return fmt.Errorf("failed to update index: %w", err)
		}
		fmt.Printf("Moved article: %s -> %s\n", oldName, newName)
		return nil
	})
}
