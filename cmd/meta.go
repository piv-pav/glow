package cmd

import (
	"fmt"
	"strings"

	"github.com/pavelpivovarov/glow/internal/index"
	"github.com/pavelpivovarov/glow/internal/storage"
	"github.com/spf13/cobra"
)

var metaCmd = &cobra.Command{
	Use:   "meta",
	Short: "Manage article metadata",
	Long:  `Set, add, or delete metadata fields in articles.`,
}

var metaSetCmd = &cobra.Command{
	Use:   "set [article-name] [key] [value]",
	Short: "Set metadata field (scalar, overwrites)",
	Args:  cobra.ExactArgs(3),
	RunE:  runMetaSet,
}

var metaAddCmd = &cobra.Command{
	Use:   "add [article-name] [key] [value...]",
	Short: "Add to metadata array field",
	Args:  cobra.MinimumNArgs(3),
	RunE:  runMetaAdd,
}

var metaDeleteCmd = &cobra.Command{
	Use:   "delete [article-name] [key] [value?]",
	Short: "Delete metadata field or remove value from array",
	Long:  `Delete entire field if no value specified, or remove value from array field.`,
	Args:  cobra.RangeArgs(2, 3),
	RunE:  runMetaDelete,
}

var metaGetCmd = &cobra.Command{
	Use:   "get [article-name] [key]",
	Short: "Get metadata field value",
	Args:  cobra.ExactArgs(2),
	RunE:  runMetaGet,
}

func init() {
	rootCmd.AddCommand(metaCmd)
	metaCmd.AddCommand(metaSetCmd)
	metaCmd.AddCommand(metaAddCmd)
	metaCmd.AddCommand(metaDeleteCmd)
	metaCmd.AddCommand(metaGetCmd)
}

func runMetaSet(cmd *cobra.Command, args []string) error {
	name := args[0]
	key := args[1]
	value := args[2]

	store := storage.New(wikiName)
	idx, err := index.New(wikiName)
	if err != nil {
		return fmt.Errorf("failed to open index: %w", err)
	}
	defer idx.Close()

	// Read article
	art, err := store.Read(name)
	if err != nil {
		return err
	}

	// Set metadata
	art.SetMetadata(key, value)

	// Save
	if err := store.Update(name, art); err != nil {
		return err
	}

	// Update index
	if err := idx.UpdateArticle(name, art); err != nil {
		return fmt.Errorf("failed to update index: %w", err)
	}

	fmt.Printf("Set %s.%s = %s\n", name, key, value)
	return nil
}

func runMetaAdd(cmd *cobra.Command, args []string) error {
	name := args[0]
	key := args[1]
	values := args[2:]

	store := storage.New(wikiName)
	idx, err := index.New(wikiName)
	if err != nil {
		return fmt.Errorf("failed to open index: %w", err)
	}
	defer idx.Close()

	// Read article
	art, err := store.Read(name)
	if err != nil {
		return err
	}

	// Add metadata
	if err := art.AddMetadata(key, values...); err != nil {
		return err
	}

	// Save
	if err := store.Update(name, art); err != nil {
		return err
	}

	// Update index
	if err := idx.UpdateArticle(name, art); err != nil {
		return fmt.Errorf("failed to update index: %w", err)
	}

	fmt.Printf("Added to %s.%s: %s\n", name, key, strings.Join(values, ", "))
	return nil
}

func runMetaDelete(cmd *cobra.Command, args []string) error {
	name := args[0]
	key := args[1]
	value := ""
	if len(args) > 2 {
		value = args[2]
	}

	store := storage.New(wikiName)
	idx, err := index.New(wikiName)
	if err != nil {
		return fmt.Errorf("failed to open index: %w", err)
	}
	defer idx.Close()

	// Read article
	art, err := store.Read(name)
	if err != nil {
		return err
	}

	// Delete metadata
	if err := art.DeleteMetadata(key, value); err != nil {
		return err
	}

	// Save
	if err := store.Update(name, art); err != nil {
		return err
	}

	// Update index
	if err := idx.UpdateArticle(name, art); err != nil {
		return fmt.Errorf("failed to update index: %w", err)
	}

	if value == "" {
		fmt.Printf("Deleted %s.%s\n", name, key)
	} else {
		fmt.Printf("Removed from %s.%s: %s\n", name, key, value)
	}

	return nil
}

func runMetaGet(cmd *cobra.Command, args []string) error {
	name := args[0]
	key := args[1]

	store := storage.New(wikiName)

	// Read article
	art, err := store.Read(name)
	if err != nil {
		return err
	}

	// Try string first, then array
	if val, ok := art.GetMetadataString(key); ok {
		fmt.Println(val)
		return nil
	}
	if vals, ok := art.GetMetadataArray(key); ok {
		fmt.Println(strings.Join(vals, ", "))
		return nil
	}

	return fmt.Errorf("metadata key not found: %s", key)
}
