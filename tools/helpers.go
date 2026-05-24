package tools

import (
	"fmt"
	"os"
	"strconv"
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

// unescapeContent interprets escape sequences in content strings using Go stdlib.
// Handles \n, \t, \\, \r, \", \xNN, \uNNNN, \UNNNNNNNN, \NNN (octal).
// Raw newlines in content (already interpreted by shell/Go) are preserved.
// Returns error for invalid escape sequences (e.g. trailing backslash, \').
func unescapeContent(s string) (string, error) {
	// strconv.Unquote rejects raw newlines in quoted strings.
	// Escape them first so both raw newlines and escape sequences get interpreted.
	s = strings.ReplaceAll(s, "\n", "\\n")
	unquoted, err := strconv.Unquote("\"" + s + "\"")
	if err != nil {
		return "", fmt.Errorf("invalid escape sequence in --content: %w\nUse --stdin instead for content with special characters", err)
	}
	return unquoted, nil
}
