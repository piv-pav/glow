package tools

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"codeberg.org/pivpav/glow/internal/article"
	"codeberg.org/pivpav/glow/internal/index"
	"codeberg.org/pivpav/glow/internal/storage"
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

// modifyArticle reads an article, applies a modification, saves, updates index, and prints successMsg.
func modifyArticle(wikiName, name string, modify func(*article.Article) error, successMsg string) error {
	store := storage.New(wikiName)
	art, err := store.Read(name)
	if err != nil {
		return err
	}

	if err := modify(art); err != nil {
		return err
	}

	if err := store.Update(name, art); err != nil {
		return err
	}

	return withIndex(wikiName, func(idx *index.Index) error {
		if err := idx.UpdateArticle(name, art); err != nil {
			return fmt.Errorf("failed to update index: %w", err)
		}
		fmt.Println(successMsg)
		return nil
	})
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
