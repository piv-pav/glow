package tools

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// wikiNameFrom returns the wiki name from the persistent --wiki flag.
func wikiNameFrom(cmd *cobra.Command) string {
	return cmd.Flag("wiki").Value.String()
}

// openEditor opens system editor for content input.
func openEditor(initialContent string) (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	tmpFile, err := os.CreateTemp("", "wiki-*.md")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpFile.Name())

	if initialContent != "" {
		if _, err := tmpFile.WriteString(initialContent); err != nil {
			return "", err
		}
	}
	tmpFile.Close()

	cmd := os.Getenv("SHELL")
	if cmd == "" {
		cmd = "/bin/sh"
	}

	editorCmd := fmt.Sprintf("%s %s", editor, tmpFile.Name())
	proc, err := os.StartProcess(cmd, []string{cmd, "-c", editorCmd}, &os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
	})
	if err != nil {
		return "", err
	}

	if _, err := proc.Wait(); err != nil {
		return "", err
	}

	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func splitLines(s string) []string {
	return strings.Split(s, "\n")
}

func joinLines(lines []string) string {
	return strings.Join(lines, "\n")
}
