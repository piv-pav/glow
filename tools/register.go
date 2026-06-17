package tools

import "github.com/spf13/cobra"

func RegisterCommands(root *cobra.Command) {
	root.AddCommand(exportCmd)
	root.AddCommand(importCmd)
	root.AddCommand(appendCmd)
	root.AddCommand(createCmd)
	root.AddCommand(deleteCmd)
	root.AddCommand(listCmd)
	root.AddCommand(moveCmd)
	root.AddCommand(readCmd)
	root.AddCommand(searchCmd)
	root.AddCommand(updateCmd)
	root.AddCommand(wikiInitCmd)
	root.AddCommand(wikiCreateCmd)
	root.AddCommand(wikiDeleteCmd)
	root.AddCommand(wikiListCmd)
	root.AddCommand(wikiRebuildCmd)
}
