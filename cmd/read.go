package cmd

import (
	"fmt"
	"strings"

	"github.com/pavelpivovarov/glow/internal/storage"
	"github.com/spf13/cobra"
)

var (
	readRaw      bool
	readSection  string
	readSections bool
)

var readCmd = &cobra.Command{
	Use:     "read [article-name]",
	Aliases: []string{"show", "cat"},
	Short:   "Read an article",
	Long:    `Display the full content of an article. Use --raw to include frontmatter.`,
	Args:    cobra.ExactArgs(1),
	RunE:    runRead,
}

func init() {
	rootCmd.AddCommand(readCmd)
	readCmd.Flags().BoolVarP(&readRaw, "raw", "r", false, "Show raw content including frontmatter")
	readCmd.Flags().StringVarP(&readSection, "section", "s", "", "Read only specific section by heading")
	readCmd.Flags().BoolVar(&readSections, "sections", false, "List all sections in the article")
}

func runRead(cmd *cobra.Command, args []string) error {
	name := args[0]

	store := storage.New(wikiName)

	// Read article
	art, err := store.Read(name)
	if err != nil {
		return err
	}

	// If listing sections
	if readSections {
		sections := art.ParseSections()
		fmt.Printf("Sections in %s:\n\n", name)
		for _, section := range sections {
			if section.Heading == "" {
				fmt.Printf("  (preamble)\n")
			} else {
				fmt.Printf("  %s %s\n", strings.Repeat("#", section.Level), section.Heading)
			}
		}
		return nil
	}

	// If section specified, find and read only that section
	if readSection != "" {
		section := art.FindSection(readSection)
		if section == nil {
			return fmt.Errorf("section not found: %s", readSection)
		}
		
		if readRaw {
			// Show section with heading
			fmt.Print(section.Content)
		} else {
			// Show section content without heading line
			lines := splitLines(section.Content)
			if len(lines) > 1 {
				fmt.Print(joinLines(lines[1:]))
			}
		}
		return nil
	}

	if readRaw {
		// Show full raw content with frontmatter
		data, err := art.Serialize()
		if err != nil {
			return fmt.Errorf("failed to serialize article: %w", err)
		}
		fmt.Print(string(data))
	} else {
		// Show just the content
		fmt.Print(art.Content)
	}

	return nil
}
