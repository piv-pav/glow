package tools

import (
	"fmt"
	"os"
	"strings"

	"git.netra.pivpav.com/public/glow/internal/index"
	"git.netra.pivpav.com/public/glow/internal/storage"
	"github.com/spf13/cobra"
)

var (
	updateSection string
	updateContent string
	updateStdin   bool
	updateMeta    []string
)

var updateCmd = &cobra.Command{
	Use:   "update [article-name]",
	Short: "Update an existing article",
	Long:  `Update article content or specific section. Opens editor for content modification. Supports --meta for updating metadata.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runUpdate,
}

func init() {
	updateCmd.Flags().StringVar(&updateSection, "section", "", "Update only specific section by heading")
	updateCmd.Flags().StringVar(&updateContent, "content", "", "New content (skips editor)")
	updateCmd.Flags().BoolVar(&updateStdin, "stdin", false, "Read content from stdin (skips editor)")
	updateCmd.Flags().StringArrayVar(&updateMeta, "meta", []string{}, "Metadata in key:value format (can be repeated)")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	name := args[0]
	wikiName := wikiNameFrom(cmd)

	store := storage.New(wikiName)
	idx, err := index.New(wikiName)
	if err != nil {
		return fmt.Errorf("failed to open index: %w", err)
	}
	defer idx.Close()

	art, err := store.Read(name)
	if err != nil {
		return err
	}

	var initialContent string
	if updateSection != "" {
		section := art.FindSection(updateSection)
		if section == nil {
			return fmt.Errorf("section not found: %s", updateSection)
		}
		lines := splitLines(section.Content)
		if len(lines) > 1 {
			initialContent = joinLines(lines[1:])
		}
	} else {
		initialContent = art.Content
	}

	var newContent string
	hasContentFlags := updateContent != "" || updateStdin || updateSection != ""
	hasMetaFlags := len(updateMeta) > 0

	if !hasContentFlags && !hasMetaFlags {
		return fmt.Errorf("nothing to update: specify --content, --stdin, --section, or --meta")
	}

	if hasContentFlags {
		if updateStdin {
			data, err := os.ReadFile("/dev/stdin")
			if err != nil {
				return fmt.Errorf("failed to read stdin: %w", err)
			}
			newContent = string(data)
		} else if updateContent != "" {
			var err error
			newContent, err = unescapeContent(updateContent)
			if err != nil {
				return err
			}
		} else {
			var err error
			newContent, err = openEditor(initialContent)
			if err != nil {
				return fmt.Errorf("failed to open editor: %w", err)
			}
		}

		if updateSection != "" {
			if err := art.UpdateSection(updateSection, newContent); err != nil {
				return err
			}
		} else {
			art.Content = newContent
		}
	}

	// Apply metadata changes
	for _, meta := range updateMeta {
		parts := strings.SplitN(meta, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid metadata format: %s (expected key:value)", meta)
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

	if err := store.Update(name, art); err != nil {
		return err
	}

	if err := idx.UpdateArticle(name, art); err != nil {
		return fmt.Errorf("failed to update index: %w", err)
	}

	fmt.Printf("Updated article: %s\n", name)
	if updateSection != "" {
		fmt.Printf("Section: %s\n", updateSection)
	}

	return nil
}
