package tools

import (
	"fmt"
	"strings"

	"codeberg.org/pivpav/glow/internal/index"
	"codeberg.org/pivpav/glow/internal/storage"
	"github.com/spf13/cobra"
)

var (
	searchFilters []string
	searchLimit   int
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search articles",
	Long: `Search articles by content and tags.

Query can include embedded filters:
  glow search "query text tag:go path:folder/"

Or use explicit filter flags:
  glow search "query text" --filter=tag:go`,
	Args: cobra.MinimumNArgs(1),
	RunE: runSearch,
}

func init() {
	searchCmd.Flags().StringSliceVar(&searchFilters, "filter", []string{}, "Filter in field:value format (can be repeated)")
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "l", 10, "Maximum number of results")
}

func runSearch(cmd *cobra.Command, args []string) error {
	queryStr := args[0]
	wikiName := wikiNameFrom(cmd)

	filters := make(map[string]string)
	for _, f := range searchFilters {
		k, v, ok := parseFilter(f)
		if !ok {
			return fmt.Errorf("invalid filter format: %s (expected field:value)", f)
		}
		filters[k] = v
	}

	// Parse embedded field:value tokens out of the query string
	queryStr, embedded := parseEmbeddedFilters(queryStr)
	for k, v := range embedded {
		if _, exists := filters[k]; !exists {
			filters[k] = v
		}
	}

	// Try native DB search first; fall back to Bleve for files backend
	store, err := storage.New(wikiName)
	if err != nil {
		return err
	}
	defer store.Close()

	if searcher, ok := store.(storage.Searcher); ok {
		results, err := searcher.Search(queryStr, filters, searchLimit)
		if err != nil {
			return err
		}
		printStorageResults(results)
		return nil
	}

	// Files backend — use Bleve
	return withIndex(wikiName, func(idx *index.Index) error {
		results, err := idx.Search(queryStr, filters, searchLimit)
		if err != nil {
			return err
		}
		printBleveResults(results)
		return nil
	})
}

func printStorageResults(results []storage.SearchResult) {
	if len(results) == 0 {
		fmt.Println("No results found")
		return
	}
	fmt.Printf("Found %d results:\n\n", len(results))
	for i, r := range results {
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

func printBleveResults(results []index.SearchResult) {
	if len(results) == 0 {
		fmt.Println("No results found")
		return
	}
	fmt.Printf("Found %d results:\n\n", len(results))
	for i, r := range results {
		fmt.Printf("%d. %s\n", i+1, r.Name)
		if r.Snippet != "" {
			fmt.Printf("   %s\n", r.Snippet)
		}
		if tags, ok := r.Fields["tags"].(string); ok && tags != "" {
			fmt.Printf("   [tags: %s]\n", tags)
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
