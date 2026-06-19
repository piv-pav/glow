package article

import (
	"fmt"
	"strings"
)

// SEARCH/REPLACE diff format (Aider-style), the way most AI tools emit edits:
//
//	<<<<<<< SEARCH
//	exact text to find
//	=======
//	replacement text
//	>>>>>>> REPLACE
//
// Multiple blocks are applied in order. Each SEARCH must match exactly once
// in the current content (after previous blocks applied). An empty SEARCH
// block means "prepend to the (currently empty) content" — used for creating
// content from scratch.

const (
	diffMarkerSearch    = "<<<<<<< SEARCH"
	diffMarkerSeparator = "======="
	diffMarkerReplace   = ">>>>>>> REPLACE"
)

// DiffBlock is a single search/replace hunk.
type DiffBlock struct {
	Search  string
	Replace string
}

// ParseDiff parses one or more SEARCH/REPLACE blocks from raw diff text.
func ParseDiff(diff string) ([]DiffBlock, error) {
	lines := strings.Split(diff, "\n")
	var blocks []DiffBlock

	i := 0
	for i < len(lines) {
		line := strings.TrimRight(lines[i], "\r")
		if strings.TrimSpace(line) != diffMarkerSearch {
			i++
			continue
		}

		// Collect SEARCH lines until separator.
		i++
		var search []string
		foundSep := false
		for i < len(lines) {
			l := strings.TrimRight(lines[i], "\r")
			if strings.TrimSpace(l) == diffMarkerSeparator {
				foundSep = true
				i++
				break
			}
			search = append(search, l)
			i++
		}
		if !foundSep {
			return nil, fmt.Errorf("malformed diff: SEARCH block missing %q separator", diffMarkerSeparator)
		}

		// Collect REPLACE lines until end marker.
		var replace []string
		foundEnd := false
		for i < len(lines) {
			l := strings.TrimRight(lines[i], "\r")
			if strings.TrimSpace(l) == diffMarkerReplace {
				foundEnd = true
				i++
				break
			}
			replace = append(replace, l)
			i++
		}
		if !foundEnd {
			return nil, fmt.Errorf("malformed diff: REPLACE block missing %q marker", diffMarkerReplace)
		}

		blocks = append(blocks, DiffBlock{
			Search:  strings.Join(search, "\n"),
			Replace: strings.Join(replace, "\n"),
		})
	}

	if len(blocks) == 0 {
		return nil, fmt.Errorf("no SEARCH/REPLACE blocks found in diff")
	}
	return blocks, nil
}

// ApplyDiff applies SEARCH/REPLACE blocks to content. Each block's SEARCH must
// match exactly once. Returns the new content and the number of blocks applied.
func ApplyDiff(content, diff string) (string, int, error) {
	blocks, err := ParseDiff(diff)
	if err != nil {
		return "", 0, err
	}

	result := content
	for idx, b := range blocks {
		// Empty search → prepend (creation/insert-at-top).
		if b.Search == "" {
			result = b.Replace + result
			continue
		}

		count := strings.Count(result, b.Search)
		if count == 0 {
			return "", 0, fmt.Errorf("block %d: SEARCH text not found:\n%s", idx+1, indent(b.Search))
		}
		if count > 1 {
			return "", 0, fmt.Errorf("block %d: SEARCH text matches %d times (must be unique); add more context:\n%s", idx+1, count, indent(b.Search))
		}
		result = strings.Replace(result, b.Search, b.Replace, 1)
	}

	return result, len(blocks), nil
}

func indent(s string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = "  | " + l
	}
	return strings.Join(lines, "\n")
}
