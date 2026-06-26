package tools

import (
	"fmt"
	"strings"

	"codeberg.org/pivpav/glow/internal/storage"
	"github.com/spf13/cobra"
)

var searchLimit int

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search articles",
	Long: `Search articles by content and tags.

Embed tag: and path: filters directly in the query:
  glow search "query text tag:go path:folder/"`,
	Args: cobra.MinimumNArgs(1),
	RunE: runSearch,
}

func init() {
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "l", 10, "Maximum number of results")
}

func runSearch(cmd *cobra.Command, args []string) error {
	queryStr := args[0]
	wikiName := wikiNameFrom(cmd)

	queryStr, filters := parseEmbeddedFilters(queryStr)

	store, err := storage.New(wikiName)
	if err != nil {
		return err
	}
	defer store.Close()

	searcher, ok := store.(storage.Searcher)
	if !ok {
		return fmt.Errorf("backend does not support search")
	}
	out, err := searcher.Search(queryStr, filters, searchLimit)
	if err != nil {
		return err
	}
	printStorageResults(out)
	return nil
}

func printStorageResults(out *storage.SearchOutput) {
	if len(out.Results) == 0 {
		fmt.Println("No results found")
		return
	}
	if out.Total > len(out.Results) {
		fmt.Printf("Found %d results (showing top %d):\n\n", out.Total, len(out.Results))
	} else {
		fmt.Printf("Found %d results:\n\n", out.Total)
	}
	for i, r := range out.Results {
		fmt.Printf("%d. %s\n", i+1, r.Name)
		if r.Snippet != "" {
			fmt.Printf("   %s\n", r.Snippet)
		}
		if len(r.Tags) > 0 {
			fmt.Printf("   [tags: %s]\n", strings.Join(r.Tags, ", "))
		}
		fmt.Println()
	}
}

func parseFilter(filter string) (string, string, bool) {
	k, v, ok := strings.Cut(filter, ":")
	if ok && k != "" && v != "" {
		return k, v, true
	}
	return "", "", false
}

// parseEmbeddedFilters splits "word tag:go path:foo" into text + filters map.
func parseEmbeddedFilters(q string) (string, map[string]string) {
	filters := make(map[string]string)
	var text []string
	for _, part := range strings.Fields(q) {
		k, v, ok := parseFilter(part)
		if ok {
			filters[k] = v
		} else {
			text = append(text, part)
		}
	}
	return strings.Join(text, " "), filters
}
