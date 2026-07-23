package tools

import (
	"fmt"

	"github.com/spf13/cobra"
)

var selfUpdateCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade glow to the latest version",
	Long:  `Glow has migrated to GitHub. Follow the manual upgrade instructions below.`,
	Args:  cobra.NoArgs,
	RunE:  runSelfUpdate,
}

func runSelfUpdate(cmd *cobra.Command, args []string) error {
	fmt.Println(`Glow has moved to GitHub.

This binary was installed from Codeberg and cannot auto-upgrade.
To switch to the GitHub version:

    go install github.com/piv-pav/glow@latest

This is a one-time manual step. Future upgrades will work with 'glow upgrade'.`)
	return nil
}
