package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const testWikiData = "/tmp/glow-test-wiki"

// TestMain sets up and tears down test environment
func TestMain(m *testing.M) {
	// Clean up any existing test data
	os.RemoveAll(testWikiData)
	
	// Run tests
	code := m.Run()
	
	// Clean up after tests
	os.RemoveAll(testWikiData)
	
	os.Exit(code)
}

// runWiki executes wiki command with WIKI_DATA set to test directory
func runWiki(args ...string) (string, error) {
	cmd := exec.Command("wiki", args...)
	cmd.Env = append(os.Environ(), "WIKI_DATA="+testWikiData)
	
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// getArticlePath returns full path to article file
func getArticlePath(name string) string {
	return filepath.Join(testWikiData, "default", "articles", name+".md")
}

// readArticle reads article content from file
func readArticle(t *testing.T, name string) string {
	t.Helper()
	
	path := getArticlePath(name)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read article %s: %v", name, err)
	}
	
	return string(content)
}

// TestWikiCreate tests article creation
func TestWikiCreate(t *testing.T) {
	// Clean test directory
	os.RemoveAll(testWikiData)
	defer os.RemoveAll(testWikiData)
	
	tests := []struct {
		name     string
		article  string
		args     []string
		wantErr  bool
		checkFn  func(t *testing.T, content string)
	}{
		{
			name:    "create simple article",
			article: "test-article",
			args:    []string{"create", "test-article"},
			wantErr: false,
			checkFn: func(t *testing.T, content string) {
				if !strings.Contains(content, "created:") {
					t.Error("Article missing created timestamp")
				}
				if !strings.Contains(content, "modified:") {
					t.Error("Article missing modified timestamp")
				}
			},
		},
		{
			name:    "create article with metadata",
			article: "tagged-article",
			args:    []string{"create", "tagged-article", "--meta", "tags:go", "--meta", "tags:cli", "--meta", "author:test"},
			wantErr: false,
			checkFn: func(t *testing.T, content string) {
				if !strings.Contains(content, "tags:") {
					t.Error("Article missing tags metadata")
				}
				if !strings.Contains(content, "author: test") {
					t.Error("Article missing author metadata")
				}
			},
		},
		{
			name:    "create nested article",
			article: "folder/subfolder/nested",
			args:    []string{"create", "folder/subfolder/nested"},
			wantErr: false,
			checkFn: func(t *testing.T, content string) {
				if !strings.Contains(content, "path: folder/subfolder/nested") {
					t.Error("Article missing path metadata")
				}
			},
		},
	}
	
		for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create article using --stdin flag
			input := "# Test Content\n\nThis is test content."
			args := append(tt.args, "--stdin")
			cmd := exec.Command("sh", "-c", "echo '"+input+"' | wiki "+strings.Join(args, " "))
			cmd.Env = append(os.Environ(), "WIKI_DATA="+testWikiData)
			
			output, err := cmd.CombinedOutput()
			
			if (err != nil) != tt.wantErr {
				t.Errorf("runWiki() error = %v, wantErr %v\nOutput: %s", err, tt.wantErr, string(output))
				return
			}
			
			if !tt.wantErr {
				// Check article was created
				content := readArticle(t, tt.article)
				tt.checkFn(t, content)
			}
		})
	}
}

// TestWikiAppend tests appending to articles
func TestWikiAppend(t *testing.T) {
	os.RemoveAll(testWikiData)
	defer os.RemoveAll(testWikiData)
	
	// Setup: create initial article
	articleName := "append-test"
	initialContent := "---\ntags: [test]\n---\n\n# Initial\n\nInitial content.\n"
	
	os.MkdirAll(filepath.Join(testWikiData, "default", "articles"), 0755)
	err := os.WriteFile(getArticlePath(articleName), []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("Failed to setup test article: %v", err)
	}
	
	// Rebuild index for search to work
	runWiki("wiki-rebuild")
	
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		check   string
	}{
		{
			name:    "append to article",
			args:    []string{"append", articleName, "Appended content."},
			wantErr: false,
			check:   "Appended content.",
		},
		{
			name:    "append to non-existent article",
			args:    []string{"append", "no-such-article", "content"},
			wantErr: true,
			check:   "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := runWiki(tt.args...)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("runWiki() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr && tt.check != "" {
				content := readArticle(t, articleName)
				if !strings.Contains(content, tt.check) {
					t.Errorf("Article does not contain expected content: %s", tt.check)
				}
			}
		})
	}
}

// TestWikiAppendSection tests appending to specific sections
func TestWikiAppendSection(t *testing.T) {
	os.RemoveAll(testWikiData)
	defer os.RemoveAll(testWikiData)
	
	// Setup: create article with sections
	articleName := "section-test"
	initialContent := `---
tags: [test]
---

# Article Title

Introduction text.

## Section One

Section one content.

## Section Two

Section two content.
`
	
	os.MkdirAll(filepath.Join(testWikiData, "default", "articles"), 0755)
	err := os.WriteFile(getArticlePath(articleName), []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("Failed to setup test article: %v", err)
	}
	
	runWiki("wiki-rebuild")
	
	// Append to specific section
	_, err = runWiki("append", articleName, "--section=Section One", "New content in section one.")
	if err != nil {
		t.Fatalf("Failed to append to section: %v", err)
	}
	
	// Verify content
	content := readArticle(t, articleName)
	
	// Check that new content is in Section One
	if !strings.Contains(content, "New content in section one.") {
		t.Error("Section does not contain appended content")
	}
	
	// Check that Section Two is still intact
	if !strings.Contains(content, "Section two content.") {
		t.Error("Other sections were affected")
	}
	
	// Verify ordering: Section One content should come before Section Two
	secOneIdx := strings.Index(content, "## Section One")
	newContentIdx := strings.Index(content, "New content in section one.")
	secTwoIdx := strings.Index(content, "## Section Two")
	
	if secOneIdx == -1 || newContentIdx == -1 || secTwoIdx == -1 {
		t.Fatal("Could not find expected sections in content")
	}
	
	if !(secOneIdx < newContentIdx && newContentIdx < secTwoIdx) {
		t.Error("Content was not appended to correct section")
	}
}

// TestWikiUpdate tests updating articles
func TestWikiUpdate(t *testing.T) {
	os.RemoveAll(testWikiData)
	defer os.RemoveAll(testWikiData)
	
	// Setup: create initial article
	articleName := "update-test"
	initialContent := "# Initial\n\nInitial content."
	
	cmd := exec.Command("sh", "-c", "echo '"+initialContent+"' | wiki create "+articleName+" --stdin --meta tags:test")
	cmd.Env = append(os.Environ(), "WIKI_DATA="+testWikiData)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to setup test article: %v\nOutput: %s", err, string(output))
	}
	
	// Read original to verify created timestamp exists
	original := readArticle(t, articleName)
	if !strings.Contains(original, "created:") {
		t.Error("Original article missing created timestamp")
	}
	
	// Update using --content flag
	updatedContent := "# Updated\n\nUpdated content."
	cmd = exec.Command("wiki", "update", articleName, "--content", updatedContent)
	cmd.Env = append(os.Environ(), "WIKI_DATA="+testWikiData)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to update article: %v\nOutput: %s", err, string(output))
	}
	
	// Verify content updated
	updated := readArticle(t, articleName)
	if !strings.Contains(updated, "Updated content") {
		t.Error("Article content was not updated")
	}
	if !strings.Contains(updated, "created:") {
		t.Error("Created timestamp was not preserved")
	}
	if !strings.Contains(updated, "modified:") {
		t.Error("Modified timestamp was not added")
	}
}

// TestWikiDelete tests deleting articles
func TestWikiDelete(t *testing.T) {
	os.RemoveAll(testWikiData)
	defer os.RemoveAll(testWikiData)
	
	// Setup: create article
	articleName := "delete-test"
	content := "---\ntags: [test]\n---\n\n# Delete Me\n\nContent.\n"
	
	os.MkdirAll(filepath.Join(testWikiData, "default", "articles"), 0755)
	err := os.WriteFile(getArticlePath(articleName), []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to setup test article: %v", err)
	}
	
	runWiki("wiki-rebuild")
	
	// Delete article
	_, err = runWiki("delete", articleName)
	if err != nil {
		t.Fatalf("Failed to delete article: %v", err)
	}
	
	// Verify article is gone
	_, err = os.Stat(getArticlePath(articleName))
	if !os.IsNotExist(err) {
		t.Error("Article still exists after deletion")
	}
}

// TestWikiSearch tests search functionality
func TestWikiSearch(t *testing.T) {
	os.RemoveAll(testWikiData)
	defer os.RemoveAll(testWikiData)
	
	// Setup: create multiple articles
	articles := map[string]string{
		"golang-basics": "---\ntags: [go, programming]\n---\n\n# Go Basics\n\nLearn Golang fundamentals.\n",
		"python-intro":  "---\ntags: [python, programming]\n---\n\n# Python Intro\n\nPython programming language.\n",
		"cli-tools":     "---\ntags: [go, cli]\nproject: tools\n---\n\n# CLI Tools\n\nBuilding CLI with Go.\n",
	}
	
	os.MkdirAll(filepath.Join(testWikiData, "default", "articles"), 0755)
	for name, content := range articles {
		err := os.WriteFile(getArticlePath(name), []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to setup test article %s: %v", name, err)
		}
	}
	
	// Rebuild index
	output, err := runWiki("wiki-rebuild")
	if err != nil {
		t.Fatalf("Failed to rebuild index: %v\nOutput: %s", err, output)
	}
	t.Logf("Index rebuilt: %s", output)
	
	tests := []struct {
		name       string
		args       []string
		wantInside []string
		wantOutside []string
	}{
		{
			name:       "search by content",
			args:       []string{"search", "golang", "-l", "10"},
			wantInside: []string{"golang-basics"},
			wantOutside: []string{"python"},
		},
		{
			name:       "search by tag",
			args:       []string{"search", "tags:go", "-l", "10"},
			wantInside: []string{"golang-basics", "cli-tools"},
			wantOutside: []string{"python-intro"},
		},
		{
			name:       "search by project",
			args:       []string{"search", "project:tools", "-l", "10"},
			wantInside: []string{"cli-tools"},
			wantOutside: []string{"golang-basics", "python-intro"},
		},
		{
			name:       "search multiple tags",
			args:       []string{"search", "tags:cli", "-l", "10"},
			wantInside: []string{"cli-tools"},
			wantOutside: []string{"golang-basics", "python-intro"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := runWiki(tt.args...)
			if err != nil {
				t.Fatalf("Search failed: %v\nOutput: %s", err, output)
			}
			
			// Check expected articles are in results
			for _, want := range tt.wantInside {
				if !strings.Contains(output, want) {
					t.Errorf("Search results missing expected article: %s\nOutput: %s", want, output)
				}
			}
			
			// Check unexpected articles are not in results
			for _, notwant := range tt.wantOutside {
				if strings.Contains(output, notwant) {
					t.Errorf("Search results contain unexpected article: %s\nOutput: %s", notwant, output)
				}
			}
		})
	}
}

// TestWikiMove tests moving/renaming articles
func TestWikiMove(t *testing.T) {
	os.RemoveAll(testWikiData)
	defer os.RemoveAll(testWikiData)
	
	// Setup: create article
	oldName := "old-article"
	newName := "folder/new-article"
	content := "---\ntags: [test]\npath: old-article\n---\n\n# Article\n\nContent.\n"
	
	os.MkdirAll(filepath.Join(testWikiData, "default", "articles"), 0755)
	err := os.WriteFile(getArticlePath(oldName), []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to setup test article: %v", err)
	}
	
	runWiki("wiki-rebuild")
	
	// Move article
	_, err = runWiki("move", oldName, newName)
	if err != nil {
		t.Fatalf("Failed to move article: %v", err)
	}
	
	// Verify old location is gone
	_, err = os.Stat(getArticlePath(oldName))
	if !os.IsNotExist(err) {
		t.Error("Old article still exists after move")
	}
	
	// Verify new location exists
	newContent := readArticle(t, newName)
	
	// Verify path metadata was updated
	if !strings.Contains(newContent, "path: folder/new-article") {
		t.Error("Path metadata was not updated after move")
	}
	
	// Verify content preserved
	if !strings.Contains(newContent, "# Article") {
		t.Error("Content was not preserved after move")
	}
}

// TestWikiMetadata tests metadata operations
func TestWikiMetadata(t *testing.T) {
	os.RemoveAll(testWikiData)
	defer os.RemoveAll(testWikiData)
	
	// Setup: create article
	articleName := "meta-test"
	content := "---\ntags: [initial]\nauthor: test\n---\n\n# Meta Test\n\nContent.\n"
	
	os.MkdirAll(filepath.Join(testWikiData, "default", "articles"), 0755)
	err := os.WriteFile(getArticlePath(articleName), []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to setup test article: %v", err)
	}
	
	runWiki("wiki-rebuild")
	
	// Test meta set
	_, err = runWiki("meta", "set", articleName, "status", "draft")
	if err != nil {
		t.Fatalf("Failed to set metadata: %v", err)
	}
	
	updated := readArticle(t, articleName)
	if !strings.Contains(updated, "status: draft") {
		t.Error("Metadata set failed")
	}
	
	// Test meta add (array)
	_, err = runWiki("meta", "add", articleName, "tags", "added")
	if err != nil {
		t.Fatalf("Failed to add metadata: %v", err)
	}
	
	updated = readArticle(t, articleName)
	if !strings.Contains(updated, "added") {
		t.Error("Metadata add failed")
	}
	
	// Test meta delete
	_, err = runWiki("meta", "delete", articleName, "status")
	if err != nil {
		t.Fatalf("Failed to delete metadata: %v", err)
	}
	
	updated = readArticle(t, articleName)
	if strings.Contains(updated, "status: draft") {
		t.Error("Metadata delete failed")
	}
}
