package article

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// headingRegex matches markdown headings (1-6 # characters followed by a space and the heading text).
var headingRegex = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)

// Article represents a wiki article with frontmatter and content
type Article struct {
	Frontmatter map[string]interface{} // tags, created, modified, path
	Content     string
	FilePath    string
}

// New creates a new article with default timestamps
func New(content string) *Article {
	return &Article{
		Frontmatter: map[string]interface{}{
			"created":  time.Now().Format(time.RFC3339),
			"modified": time.Now().Format(time.RFC3339),
		},
		Content: content,
	}
}

// Parse parses article from markdown with YAML frontmatter
func Parse(data []byte) (*Article, error) {
	article := &Article{
		Frontmatter: make(map[string]interface{}),
	}

	// Check for frontmatter
	if !bytes.HasPrefix(data, []byte("---\n")) && !bytes.HasPrefix(data, []byte("---\r\n")) {
		// No frontmatter, treat all as content
		article.Content = string(data)
		return article, nil
	}

	// Find end of frontmatter
	endDelim := []byte("\n---\n")
	endIdx := bytes.Index(data[4:], endDelim)
	if endIdx == -1 {
		endDelim = []byte("\r\n---\r\n")
		endIdx = bytes.Index(data[4:], endDelim)
	}

	if endIdx == -1 {
		return nil, fmt.Errorf("malformed frontmatter: no closing ---")
	}

	// Parse YAML frontmatter
	frontmatter := data[4 : 4+endIdx]
	if err := yaml.Unmarshal(frontmatter, &article.Frontmatter); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Extract content after frontmatter
	contentStart := 4 + endIdx + len(endDelim)
	if contentStart < len(data) {
		article.Content = string(data[contentStart:])
	}

	return article, nil
}

// Serialize converts article back to markdown with frontmatter
func (a *Article) Serialize() ([]byte, error) {
	var buf bytes.Buffer

	if len(a.Frontmatter) > 0 {
		buf.WriteString("---\n")
		yamlData, err := yaml.Marshal(a.Frontmatter)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal frontmatter: %w", err)
		}
		buf.Write(yamlData)
		buf.WriteString("---\n")
	}

	buf.WriteString(a.Content)

	return buf.Bytes(), nil
}

// Section represents a markdown section
type Section struct {
	Heading string
	Level   int
	Content string
	Start   int
	End     int
}

// ParseSections parses markdown content into sections by headings
func (a *Article) ParseSections() []Section {
	lines := strings.Split(a.Content, "\n")
	var sections []Section

	currentSection := Section{
		Heading: "",
		Level:   0,
		Start:   0,
	}

	for i, line := range lines {
		if matches := headingRegex.FindStringSubmatch(line); matches != nil {
			level := len(matches[1])
			heading := strings.TrimSpace(matches[2])

			if currentSection.Heading != "" || i > 0 {
				currentSection.End = i
				currentSection.Content = strings.Join(lines[currentSection.Start:currentSection.End], "\n")
				sections = append(sections, currentSection)
			}

			currentSection = Section{
				Heading: heading,
				Level:   level,
				Start:   i,
			}
		}
	}

	// Add final section
	currentSection.End = len(lines)
	currentSection.Content = strings.Join(lines[currentSection.Start:currentSection.End], "\n")
	sections = append(sections, currentSection)

	return sections
}

// FindSection finds section by heading (case-insensitive, ignores # markers)
func (a *Article) FindSection(heading string) *Section {
	sections := a.ParseSections()
	searchHeading := strings.ToLower(strings.TrimSpace(heading))

	// Remove leading # if present
	searchHeading = strings.TrimLeft(searchHeading, "# ")

	for i := range sections {
		sectionHeading := strings.ToLower(strings.TrimSpace(sections[i].Heading))
		if sectionHeading == searchHeading {
			return &sections[i]
		}
	}

	return nil
}

// UpdateSection replaces content of specific section
func (a *Article) UpdateSection(heading, newContent string) error {
	section := a.FindSection(heading)
	if section == nil {
		return fmt.Errorf("section not found: %s", heading)
	}

	lines := strings.Split(a.Content, "\n")

	// Check if newContent starts with the same heading
	newContentLines := strings.Split(strings.TrimSpace(newContent), "\n")
	contentStartsWithHeading := false
	if len(newContentLines) > 0 {
		if matches := headingRegex.FindStringSubmatch(newContentLines[0]); matches != nil {
			newContentHeading := strings.TrimSpace(matches[2])
			if newContentHeading == heading {
				contentStartsWithHeading = true
			}
		}
	}

	var newLines []string
	newLines = append(newLines, lines[:section.Start]...)

	if !contentStartsWithHeading {
		newLines = append(newLines, lines[section.Start])
		newLines = append(newLines, "")
	}

	newLines = append(newLines, newContent)

	if section.End < len(lines) && !strings.HasSuffix(newContent, "\n") {
		newLines = append(newLines, "")
	}

	if section.End < len(lines) {
		newLines = append(newLines, lines[section.End:]...)
	}

	a.Content = strings.Join(newLines, "\n")
	return nil
}

// DeleteSection removes a section by heading
func (a *Article) DeleteSection(heading string) error {
	section := a.FindSection(heading)
	if section == nil {
		return fmt.Errorf("section not found: %s", heading)
	}

	lines := strings.Split(a.Content, "\n")

	var newLines []string
	newLines = append(newLines, lines[:section.Start]...)
	if section.End < len(lines) {
		newLines = append(newLines, lines[section.End:]...)
	}

	for len(newLines) > 0 && strings.TrimSpace(newLines[len(newLines)-1]) == "" {
		newLines = newLines[:len(newLines)-1]
	}
	if len(newLines) > 0 {
		newLines = append(newLines, "")
	}

	a.Content = strings.Join(newLines, "\n")
	return nil
}

func (a *Article) AppendToSection(heading, content string) error {
	section := a.FindSection(heading)
	if section == nil {
		return fmt.Errorf("section %q not found", heading)
	}

	lines := strings.Split(a.Content, "\n")

	insertIdx := section.End

	var newLines []string
	newLines = append(newLines, lines[:insertIdx]...)
	newLines = append(newLines, "", content)

	if insertIdx < len(lines) {
		newLines = append(newLines, lines[insertIdx:]...)
	}

	a.Content = strings.Join(newLines, "\n")
	return nil
}
