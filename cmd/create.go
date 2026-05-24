package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/pavelpivovarov/glow/internal/article"
	"github.com/pavelpivovarov/glow/internal/index"
	"github.com/pavelpivovarov/glow/internal/storage"
	"github.com/spf13/cobra"
)

var (
	createMeta    []string
	createContent string
	createStdin   bool
	createEditor  bool
)

var createCmd = &cobra.Command{
	Use:   "create [article-name]",
	Short: "Create a new article",
	Long:  `Create a new article with optional metadata. Article name can include folders (e.g., folder/article).`,
	Args:  cobra.ExactArgs(1),
	RunE:  runCreate,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if !createStdin && createContent == "" && !createEditor {
			return fmt.Errorf("must specify one of: --content, --stdin, or --editor")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringSliceVar(&createMeta, "meta", []string{}, "Metadata in key:value format (can be repeated)")
	createCmd.Flags().StringVar(&createContent, "content", "", "Article content")
	createCmd.Flags().BoolVar(&createStdin, "stdin", false, "Read content from stdin")
	createCmd.Flags().BoolVar(&createEditor, "editor", false, "Open editor for content")
}

func runCreate(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Create storage and index
	store := storage.New(wikiName)
	idx, err := index.New(wikiName)
	if err != nil {
		return fmt.Errorf("failed to open index: %w", err)
	}
	defer idx.Close()

	// Create new article
	art := article.New("")

	// Parse and add metadata from flags
	for _, meta := range createMeta {
		parts := strings.SplitN(meta, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid metadata format: %s (expected key:value)", meta)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Check if value is comma-separated (array)
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

	// Get content from flag, stdin, or editor
	var content string
	if createStdin {
		// Read from stdin
		data, err := os.ReadFile("/dev/stdin")
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}
		content = string(data)
	} else if createContent != "" {
		// Use content from flag
		content = createContent
	} else if createEditor {
		// Open editor
		var err error
		content, err = openEditor("")
		if err != nil {
			return fmt.Errorf("failed to open editor: %w", err)
		}
	}
	art.Content = content

	// Save article
	if err := store.Create(name, art); err != nil {
		return err
	}

	// Index article
	if err := idx.IndexArticle(name, art); err != nil {
		return fmt.Errorf("failed to index article: %w", err)
	}

	fmt.Printf("Created article: %s\n", name)
	return nil
}

// openEditor opens system editor for content input
func openEditor(initialContent string) (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	// Create temp file
	tmpFile, err := os.CreateTemp("", "wiki-*.md")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpFile.Name())

	// Write initial content
	if initialContent != "" {
		if _, err := tmpFile.WriteString(initialContent); err != nil {
			return "", err
		}
	}
	tmpFile.Close()

	// Open editor
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

	// Read content
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return "", err
	}

	return string(content), nil
}
