package cmd

import (
	"fmt"

	"github.com/pavelpivovarov/glow/internal/article"
	"github.com/pavelpivovarov/glow/internal/config"
	"github.com/pavelpivovarov/glow/internal/index"
	"github.com/pavelpivovarov/glow/internal/storage"
	"github.com/spf13/cobra"
)

var wikiCreateCmd = &cobra.Command{
	Use:   "wiki-create [name]",
	Short: "Create a new wiki",
	Long:  `Create a new wiki with the specified name. Creates directory structure and initializes index.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runWikiCreate,
}

var wikiListCmd = &cobra.Command{
	Use:   "wiki-list",
	Short: "List all wikis",
	Long:  `List all available wikis.`,
	Args:  cobra.NoArgs,
	RunE:  runWikiList,
}

var wikiVerifyCmd = &cobra.Command{
	Use:   "wiki-verify",
	Short: "Verify wiki index health",
	Long:  `Verify the health of the wiki index and display statistics.`,
	Args:  cobra.NoArgs,
	RunE:  runWikiVerify,
}

var wikiRebuildCmd = &cobra.Command{
	Use:   "wiki-rebuild",
	Short: "Rebuild wiki index",
	Long:  `Completely rebuild the wiki index from all articles. Use when index is corrupted.`,
	Args:  cobra.NoArgs,
	RunE:  runWikiRebuild,
}

func init() {
	rootCmd.AddCommand(wikiCreateCmd)
	rootCmd.AddCommand(wikiListCmd)
	rootCmd.AddCommand(wikiVerifyCmd)
	rootCmd.AddCommand(wikiRebuildCmd)
}

func runWikiCreate(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Create wiki structure
	if err := config.CreateWiki(name); err != nil {
		return err
	}

	// Initialize index
	idx, err := index.New(name)
	if err != nil {
		return fmt.Errorf("failed to initialize index: %w", err)
	}
	defer idx.Close()

	fmt.Printf("Created wiki: %s\n", name)
	
	wikiPath, _ := config.GetWikiPath(name)
	fmt.Printf("Location: %s\n", wikiPath)

	return nil
}

func runWikiList(cmd *cobra.Command, args []string) error {
	wikis, err := config.ListWikis()
	if err != nil {
		return err
	}

	if len(wikis) == 0 {
		fmt.Println("No wikis found")
		return nil
	}

	fmt.Printf("Available wikis (%d):\n\n", len(wikis))
	for _, wiki := range wikis {
		wikiPath, _ := config.GetWikiPath(wiki)
		fmt.Printf("  %s\n", wiki)
		fmt.Printf("    %s\n", wikiPath)
	}

	return nil
}

func runWikiVerify(cmd *cobra.Command, args []string) error {
	// Ensure wiki exists
	exists, err := config.WikiExists(wikiName)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("wiki does not exist: %s", wikiName)
	}

	idx, err := index.New(wikiName)
	if err != nil {
		return fmt.Errorf("failed to open index: %w", err)
	}
	defer idx.Close()

	stats, err := idx.Verify()
	if err != nil {
		fmt.Printf("Index verification FAILED for wiki '%s':\n", wikiName)
		fmt.Printf("  Error: %v\n", err)
		if statsData, ok := stats["error"]; ok {
			fmt.Printf("  Details: %v\n", statsData)
		}
		return err
	}

	fmt.Printf("Index verification OK for wiki '%s':\n", wikiName)
	fmt.Printf("  Document count: %v\n", stats["document_count"])
	fmt.Printf("  Searchable: %v\n", stats["searchable"])
	fmt.Printf("  Total hits (test): %v\n", stats["total_hits"])

	return nil
}

func runWikiRebuild(cmd *cobra.Command, args []string) error {
	// Ensure wiki exists
	exists, err := config.WikiExists(wikiName)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("wiki does not exist: %s", wikiName)
	}

	fmt.Printf("Rebuilding index for wiki '%s'...\n", wikiName)

	// Load all articles
	store := storage.New(wikiName)
	articleNames, err := store.List()
	if err != nil {
		return fmt.Errorf("failed to list articles: %w", err)
	}

	articles := make(map[string]*article.Article)
	for _, name := range articleNames {
		art, err := store.Read(name)
		if err != nil {
			fmt.Printf("Warning: failed to read article %s: %v\n", name, err)
			continue
		}
		articles[name] = art
	}

	// Rebuild index
	idx, err := index.New(wikiName)
	if err != nil {
		return fmt.Errorf("failed to open index: %w", err)
	}
	defer idx.Close()

	if err := idx.Rebuild(articles); err != nil {
		return fmt.Errorf("failed to rebuild index: %w", err)
	}

	fmt.Printf("Successfully rebuilt index for wiki '%s'\n", wikiName)
	fmt.Printf("Indexed %d articles\n", len(articles))

	return nil
}
