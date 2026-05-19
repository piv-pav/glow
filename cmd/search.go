package cmd

import (
	"fmt"
	"strings"

	"github.com/pavelpivovarov/glow/internal/index"
	"github.com/spf13/cobra"
)

var (
	searchFilters []string
	searchLimit   int
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search articles",
	Long: `Search articles by content and metadata. 
	
Query can include embedded filters:
  wiki search "query text tag:go project:glow path:folder/"
  
Or use explicit filter flags:
  wiki search "query text" --filter=tag:go --filter=project:glow`,
	Args: cobra.MinimumNArgs(1),
	RunE: runSearch,
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().StringSliceVar(&searchFilters, "filter", []string{}, "Filter in field:value format (can be repeated)")
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "l", 10, "Maximum number of results")
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := args[0]

	// Create index
	idx, err := index.New(wikiName)
	if err != nil {
		return fmt.Errorf("failed to open index: %w", err)
	}
	defer idx.Close()

	// Parse explicit filters
	filters := make(map[string]string)
	for _, filter := range searchFilters {
		key, value, ok := parseFilter(filter)
		if !ok {
			return fmt.Errorf("invalid filter format: %s (expected field:value)", filter)
		}
		filters[key] = value
	}

	// Search
	results, err := idx.Search(query, filters, searchLimit)
	if err != nil {
		return err
	}

	// Display results
	if len(results) == 0 {
		fmt.Println("No results found")
		return nil
	}

	fmt.Printf("Found %d results:\n\n", len(results))
	for i, result := range results {
		fmt.Printf("%d. %s (score: %.2f)\n", i+1, result.Name, result.Score)
		
		// Show snippet if available
		if result.Snippet != "" {
			fmt.Printf("   %s\n", result.Snippet)
		}
		
		// Collect metadata to show
		var metaParts []string
		
		if tags, ok := result.Metadata["tags"].(string); ok && tags != "" {
			metaParts = append(metaParts, "tags: "+tags)
		}
		
		if project, ok := result.Metadata["project"].(string); ok && project != "" {
			metaParts = append(metaParts, "project: "+project)
		}
		
		// Show metadata on single line if present
		if len(metaParts) > 0 {
			fmt.Printf("   [%s]\n", strings.Join(metaParts, " | "))
		}
		
		fmt.Println()
	}

	return nil
}

func parseFilter(filter string) (string, string, bool) {
	for i, ch := range filter {
		if ch == ':' {
			if i > 0 && i < len(filter)-1 {
				return filter[:i], filter[i+1:], true
			}
		}
	}
	return "", "", false
}
