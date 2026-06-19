package tools

import (
	"fmt"
	"strings"

	"codeberg.org/pivpav/glow/internal/storage"
	"github.com/spf13/cobra"
)

var (
	readTags     bool
	readSection  string
	readSections bool
)

var readCmd = &cobra.Command{
	Use:     "read [article-name]",
	Aliases: []string{"show", "cat"},
	Short:   "Read an article",
	Long:    `Display the full content of an article. Use --tags to list its tags.`,
	Args:    cobra.ExactArgs(1),
	RunE:    runRead,
}

func init() {
	readCmd.Flags().BoolVarP(&readTags, "tags", "t", false, "List only the article's tags")
	readCmd.Flags().StringVarP(&readSection, "section", "s", "", "Read only specific section by heading")
	readCmd.Flags().BoolVar(&readSections, "sections", false, "List all sections in the article")
}

func runRead(cmd *cobra.Command, args []string) error {
	name := args[0]
	wikiName := wikiNameFrom(cmd)

	return withStore(wikiName, func(store storage.Store) error {
		art, err := store.Read(name)
		if err != nil {
			return err
		}

		if readTags {
			for _, tag := range art.GetTags() {
				fmt.Println(tag)
			}
			return nil
		}

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

		if readSection != "" {
			section := art.FindSection(readSection)
			if section == nil {
				return fmt.Errorf("section not found: %s", readSection)
			}
			lines := strings.Split(section.Content, "\n")
			if len(lines) > 1 {
				fmt.Print(strings.Join(lines[1:], "\n"))
			}
			return nil
		}

		fmt.Print(art.Content)
		return nil
	})
}
