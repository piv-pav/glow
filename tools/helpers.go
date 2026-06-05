package tools

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"codeberg.org/pivpav/glow/internal/article"
	"codeberg.org/pivpav/glow/internal/index"
	"github.com/spf13/cobra"
)

// wikiNameFrom returns the wiki name from the persistent --wiki flag.
func wikiNameFrom(cmd *cobra.Command) string {
	return cmd.Flag("wiki").Value.String()
}

// withIndex opens index, executes function, guarantees cleanup.
func withIndex(wikiName string, fn func(*index.Index) error) error {
	idx, err := index.New(wikiName)
	if err != nil {
		return fmt.Errorf("failed to open index: %w", err)
	}
	defer idx.Close()
	return fn(idx)
}

// readContent returns content from stdin or --content flag (with escape interpretation).
func readContent(stdin bool, content string) (string, error) {
	if stdin {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("failed to read stdin: %w", err)
		}
		return string(data), nil
	}
	return unescapeContent(content)
}

// parseMeta parses "key:value" metadata strings into an article's metadata.
func parseMeta(art *article.Article, meta []string) error {
	for _, m := range meta {
		parts := strings.SplitN(m, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid metadata format: %s (expected key:value)", m)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if strings.Contains(value, ",") {
			values := strings.Split(value, ",")
			for i := range values {
				values[i] = strings.TrimSpace(values[i])
			}
			if err := art.AddMetadata(key, values...); err != nil {
				return err
			}
		} else {
			art.SetMetadata(key, value)
		}
	}
	return nil
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
	// Fast path: no escapes
	if !strings.Contains(s, "\\") {
		return s, nil
	}

	// strconv.Unquote rejects raw newlines in quoted strings.
	// Escape them first so both raw newlines and escape sequences get interpreted.
	s = strings.ReplaceAll(s, "\n", "\\n")
	unquoted, err := strconv.Unquote(`"` + s + `"`)
	if err != nil {
		return "", fmt.Errorf("invalid escape sequence in --content: %w", err)
	}
	return unquoted, nil
}
