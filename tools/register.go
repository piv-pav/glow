package tools

import "github.com/spf13/cobra"

func RegisterCommands(root *cobra.Command) {
	root.AddCommand(appendCmd)
	root.AddCommand(createCmd)
	root.AddCommand(deleteCmd)
	root.AddCommand(listCmd)
	root.AddCommand(metaCmd)
	root.AddCommand(moveCmd)
	root.AddCommand(readCmd)
	root.AddCommand(searchCmd)
	root.AddCommand(updateCmd)
	root.AddCommand(wikiCreateCmd)
	root.AddCommand(wikiListCmd)
	root.AddCommand(wikiRebuildCmd)
}
