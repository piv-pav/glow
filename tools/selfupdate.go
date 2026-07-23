package tools

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/mod/semver"

	"github.com/spf13/cobra"
)

const githubTagsURL = "https://api.github.com/repos/piv-pav/glow/tags?per_page=1"

type remoteTag struct {
	Name string `json:"name"`
}

var selfUpdateCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade glow to the latest version",
	Long:  `Check the latest release on GitHub and upgrade the glow binary via go install.`,
	Args:  cobra.NoArgs,
	RunE:  runSelfUpdate,
}

func runSelfUpdate(cmd *cobra.Command, args []string) error {
	fmt.Println("Checking latest version on GitHub...")

	latest, err := fetchLatestTag()
	if err != nil {
		return fmt.Errorf("failed to fetch latest version: %w", err)
	}

	// Normalise: current version may be "v0.11.0" or "v0.11.0-dev"
	current := cmd.Root().Version
	currentBase := current
	if idx := strings.Index(current, "-"); idx >= 0 {
		currentBase = current[:idx]
	}

	fmt.Printf("Current: %s\n", current)
	fmt.Printf("Latest:  %s\n", latest)

	if cmp := semver.Compare(currentBase, latest); cmp >= 0 {
		fmt.Println("Already up to date.")
		return nil
	}

	pkg := "github.com/piv-pav/glow@" + latest
	fmt.Printf("Installing %s ...\n", pkg)

	goCmd := exec.Command("go", "install", pkg)
	goCmd.Stdout = os.Stdout
	goCmd.Stderr = os.Stderr
	goCmd.Env = os.Environ()

	if err := goCmd.Run(); err != nil {
		return fmt.Errorf("go install failed: %w", err)
	}

	fmt.Printf("Upgraded to %s\n", latest)
	return nil
}

func fetchLatestTag() (string, error) {
	resp, err := http.Get(githubTagsURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected HTTP status: %s", resp.Status)
	}

	var tags []remoteTag
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}
	if len(tags) == 0 {
		return "", fmt.Errorf("no tags found on GitHub")
	}

	return tags[0].Name, nil
}
