package tools

import (
	"fmt"

	"github.com/piv-pav/glow/internal/storage"
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
		fmt.Printf("Moved article: %s -> %s\n", oldName, newName)
		return nil
	})
}
