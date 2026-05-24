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

// runWiki executes wiki command with GLOW_DATA set to test directory
func runWiki(args ...string) (string, error) {
	cmd := exec.Command("glow", args...)
	cmd.Env = append(os.Environ(), "GLOW_DATA="+testWikiData)

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
			cmd := exec.Command("sh", "-c", "echo '"+input+"' | glow "+strings.Join(args, " "))
			cmd.Env = append(os.Environ(), "GLOW_DATA="+testWikiData)

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
	runWiki("rebuild")

	tests := []struct {
		name    string
		args    []string
		wantErr bool
		check   string
	}{
		{
			name:    "append with --content flag",
			args:    []string{"append", articleName, "--content", "Appended content."},
			wantErr: false,
			check:   "Appended content.",
		},
		{
			name:    "append to non-existent article",
			args:    []string{"append", "no-such-article", "--content", "content"},
			wantErr: true,
			check:   "",
		},
		{
			name:    "append without content flag",
			args:    []string{"append", articleName},
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

	runWiki("rebuild")

	// Append to specific section using --content
	_, err = runWiki("append", articleName, "--section=Section One", "--content", "New content in section one.")
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

	cmd := exec.Command("sh", "-c", "echo '"+initialContent+"' | glow create "+articleName+" --stdin --meta tags:test")
	cmd.Env = append(os.Environ(), "GLOW_DATA="+testWikiData)
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
	cmd = exec.Command("glow", "update", articleName, "--content", updatedContent)
	cmd.Env = append(os.Environ(), "GLOW_DATA="+testWikiData)
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

// TestWikiUpdateMeta tests updating metadata via update --meta
func TestWikiUpdateMeta(t *testing.T) {
	os.RemoveAll(testWikiData)
	defer os.RemoveAll(testWikiData)

	// Setup: create article
	articleName := "update-meta-test"
	_, err := runWiki("create", articleName, "--content", "# Meta Test\n\nContent.", "--meta", "status:draft")
	if err != nil {
		t.Fatalf("Failed to create article: %v", err)
	}

	// Verify initial metadata
	output, err := runWiki("meta", "get", articleName, "status")
	if err != nil {
		t.Fatalf("meta get failed: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(output, "draft") {
		t.Errorf("Expected status=draft, got: %s", output)
	}

	// Update metadata via update --meta (scalar)
	_, err = runWiki("update", articleName, "--content", "# Updated\n\nUpdated content.", "--meta", "status:published", "--meta", "author:test")
	if err != nil {
		t.Fatalf("update --meta failed: %v", err)
	}

	// Verify scalar metadata changed
	output, err = runWiki("meta", "get", articleName, "status")
	if err != nil {
		t.Fatalf("meta get failed: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(output, "published") {
		t.Errorf("Expected status=published, got: %s", output)
	}

	// Verify new metadata added
	output, err = runWiki("meta", "get", articleName, "author")
	if err != nil {
		t.Fatalf("meta get failed: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(output, "test") {
		t.Errorf("Expected author=test, got: %s", output)
	}

	// Update with array metadata via comma-separated value
	_, err = runWiki("update", articleName, "--meta", "tags:go,cli")
	if err != nil {
		t.Fatalf("update --meta with array failed: %v", err)
	}

	output, err = runWiki("meta", "get", articleName, "tags")
	if err != nil {
		t.Fatalf("meta get tags failed: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(output, "go") || !strings.Contains(output, "cli") {
		t.Errorf("Expected tags=go,cli, got: %s", output)
	}

	// Verify content was also updated
	content := readArticle(t, articleName)
	if !strings.Contains(content, "Updated content") {
		t.Error("Content should also be updated")
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

	runWiki("rebuild")

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

// TestWikiDeleteSection tests delete --section functionality
func TestWikiDeleteSection(t *testing.T) {
	os.RemoveAll(testWikiData)
	defer os.RemoveAll(testWikiData)

	// Setup: create article with multiple sections
	articleName := "delete-section-test"
	content := "---\ntags: [test]\n---\n\n# Article\n\nIntro content.\n\n## Section One\n\nFirst section content.\n\n## Section Two\n\nSecond section content.\n\n## Section Three\n\nThird section content.\n"

	os.MkdirAll(filepath.Join(testWikiData, "default", "articles"), 0755)
	err := os.WriteFile(getArticlePath(articleName), []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to setup test article: %v", err)
	}

	// Delete Section Two
	output, err := runWiki("delete", articleName, "--section", "Section Two")
	if err != nil {
		t.Fatalf("Failed to delete section: %v\nOutput: %s", err, output)
	}

	// Verify output mentions section deletion
	if !strings.Contains(output, "Deleted section") {
		t.Errorf("Expected output to mention section deletion, got: %s", output)
	}

	// Verify article still exists
	remaining := readArticle(t, articleName)

	// Section Two should be gone
	if strings.Contains(remaining, "## Section Two") {
		t.Error("Section Two still exists after deletion")
	}
	if strings.Contains(remaining, "Second section content") {
		t.Error("Section Two content still exists after deletion")
	}

	// Other sections should remain
	if !strings.Contains(remaining, "## Section One") {
		t.Error("Section One missing after deleting Section Two")
	}
	if !strings.Contains(remaining, "## Section Three") {
		t.Error("Section Three missing after deleting Section Two")
	}
	if !strings.Contains(remaining, "Intro content") {
		t.Error("Intro content missing after deleting section")
	}

	// Test deleting non-existent section
	_, err = runWiki("delete", articleName, "--section", "NonExistent")
	if err == nil {
		t.Error("Expected error when deleting non-existent section")
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
	output, err := runWiki("rebuild")
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

	runWiki("rebuild")

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

	runWiki("rebuild")

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

// TestWikiMetaGet tests getting metadata values
func TestWikiMetaGet(t *testing.T) {
	os.RemoveAll(testWikiData)
	defer os.RemoveAll(testWikiData)

	// Setup: create article
	articleName := "meta-get-test"
	content := "---\ntags: [one, two]\nauthor: test\n---\n\n# Meta Get Test\n\nContent.\n"

	os.MkdirAll(filepath.Join(testWikiData, "default", "articles"), 0755)
	err := os.WriteFile(getArticlePath(articleName), []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to setup test article: %v", err)
	}

	runWiki("rebuild")

	tests := []struct {
		name   string
		key    string
		want   string
		wantErr bool
	}{
		{
			name:   "get scalar metadata",
			key:    "author",
			want:   "test",
			wantErr: false,
		},
		{
			name:   "get array metadata",
			key:    "tags",
			want:   "one, two",
			wantErr: false,
		},
		{
			name:   "get non-existent key",
			key:    "missing",
			want:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := runWiki("meta", "get", articleName, tt.key)

			if (err != nil) != tt.wantErr {
				t.Errorf("meta get error = %v, wantErr %v\nOutput: %s", err, tt.wantErr, output)
				return
			}

			if !tt.wantErr && !strings.Contains(output, tt.want) {
				t.Errorf("meta get output = %q, want to contain %q", output, tt.want)
			}
		})
	}
}

// TestWikiAppendStdin tests appending content via stdin
func TestWikiAppendStdin(t *testing.T) {
	os.RemoveAll(testWikiData)
	defer os.RemoveAll(testWikiData)

	// Setup: create initial article
	articleName := "append-stdin-test"
	initialContent := "---\ntags: [test]\n---\n\n# Initial\n\nInitial content.\n"

	os.MkdirAll(filepath.Join(testWikiData, "default", "articles"), 0755)
	err := os.WriteFile(getArticlePath(articleName), []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("Failed to setup test article: %v", err)
	}

	runWiki("rebuild")

	// Append via stdin
	stdinContent := "Appended via stdin."
	cmd := exec.Command("sh", "-c", "echo '"+stdinContent+"' | glow append "+articleName+" --stdin")
	cmd.Env = append(os.Environ(), "GLOW_DATA="+testWikiData)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to append via stdin: %v\nOutput: %s", err, string(output))
	}

	// Verify content
	content := readArticle(t, articleName)
	if !strings.Contains(content, stdinContent) {
		t.Errorf("Article does not contain stdin content. Got:\n%s", content)
	}

	// Append to section via stdin
	sectionContent := "Appended to section via stdin."
	cmd = exec.Command("sh", "-c", "echo '"+sectionContent+"' | glow append "+articleName+" --section \"Initial\" --stdin")
	cmd.Env = append(os.Environ(), "GLOW_DATA="+testWikiData)

	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to append to section via stdin: %v\nOutput: %s", err, string(output))
	}

	content = readArticle(t, articleName)
	if !strings.Contains(content, sectionContent) {
		t.Errorf("Article does not contain section stdin content. Got:\n%s", content)
	}
}

// TestWikiList tests listing articles
func TestWikiList(t *testing.T) {
	os.RemoveAll(testWikiData)
	defer os.RemoveAll(testWikiData)

	// List empty wiki
	output, err := runWiki("list")
	if err != nil {
		t.Fatalf("list on empty wiki failed: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(output, "No articles") {
		t.Errorf("Expected 'No articles' in output, got: %s", output)
	}

	// Create some articles
	for _, name := range []string{"alpha", "beta", "gamma"} {
		_, err := runWiki("create", name, "--content", "# "+name+"\n\nContent.")
		if err != nil {
			t.Fatalf("Failed to create article %s: %v", name, err)
		}
	}

	// List with articles
	output, err = runWiki("list")
	if err != nil {
		t.Fatalf("list failed: %v\nOutput: %s", err, output)
	}

	for _, name := range []string{"alpha", "beta", "gamma"} {
		if !strings.Contains(output, name) {
			t.Errorf("Expected %s in list output, got: %s", name, output)
		}
	}

	// Check article count shown
	if !strings.Contains(output, "(3)") {
		t.Errorf("Expected article count (3) in output, got: %s", output)
	}
}

// TestWikiRead tests reading articles
func TestWikiRead(t *testing.T) {
	os.RemoveAll(testWikiData)
	defer os.RemoveAll(testWikiData)

	// Create article with sections (via --content, frontmatter added automatically)
	content := `# Title

Intro.

## Section A

Content A.

## Section B

Content B.`
	cmd := exec.Command("sh", "-c", "echo '"+content+"' | glow create read-test --stdin --meta tags:test")
	cmd.Env = append(os.Environ(), "GLOW_DATA="+testWikiData)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to create article: %v\nOutput: %s", err, string(output))
	}

	// Read without --raw (should omit frontmatter)
	output, err := runWiki("read", "read-test")
	if err != nil {
		t.Fatalf("read failed: %v\nOutput: %s", err, output)
	}
	if strings.Contains(output, "tags:") {
		t.Error("default read should not include frontmatter")
	}
	if !strings.Contains(output, "Intro.") {
		t.Error("default read should include content")
	}

	// Read with --raw (should include frontmatter)
	output, err = runWiki("read", "read-test", "--raw")
	if err != nil {
		t.Fatalf("read --raw failed: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(output, "tags:") {
		t.Error("raw read should include frontmatter")
	}

	// Read with --sections (list sections)
	output, err = runWiki("read", "read-test", "--sections")
	if err != nil {
		t.Fatalf("read --sections failed: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(output, "Title") {
		t.Error("sections output missing Title")
	}
	if !strings.Contains(output, "Section A") {
		t.Error("sections output missing Section A")
	}
	if !strings.Contains(output, "Section B") {
		t.Error("sections output missing Section B")
	}

	// Read specific section
	output, err = runWiki("read", "read-test", "--section", "Section A")
	if err != nil {
		t.Fatalf("read --section failed: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(output, "Content A.") {
		t.Errorf("section read missing Content A, got: %s", output)
	}
	if strings.Contains(output, "Content B.") {
		t.Error("section read should not include other sections")
	}

	// Read non-existent article
	_, err = runWiki("read", "no-such-article")
	if err == nil {
		t.Error("expected error reading non-existent article")
	}

	// Read non-existent section
	_, err = runWiki("read", "read-test", "--section", "NoSection")
	if err == nil {
		t.Error("expected error reading non-existent section")
	}
}

// TestWikiCreateWiki tests wiki subcommands
func TestWikiCreateWiki(t *testing.T) {
	os.RemoveAll(testWikiData)
	defer os.RemoveAll(testWikiData)

	// Create a new wiki
	output, err := runWiki("wiki-create", "mywiki")
	if err != nil {
		t.Fatalf("wiki-create failed: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(output, "Created wiki: mywiki") {
		t.Errorf("Unexpected output: %s", output)
	}

	// Create duplicate wiki
	_, err = runWiki("wiki-create", "mywiki")
	if err == nil {
		t.Error("expected error on duplicate wiki creation")
	}

	// Articles in new wiki should be separate from default
	_, err = runWiki("create", "test-article", "--content", "# Hello", "--wiki", "mywiki")
	if err != nil {
		t.Fatalf("create in mywiki failed: %v", err)
	}

	// Should not appear in default wiki
	output, err = runWiki("list")
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if strings.Contains(output, "test-article") {
		t.Error("article in mywiki should not appear in default wiki list")
	}

	// Should appear in mywiki
	output, err = runWiki("list", "--wiki", "mywiki")
	if err != nil {
		t.Fatalf("list --wiki mywiki failed: %v", err)
	}
	if !strings.Contains(output, "test-article") {
		t.Errorf("article not found in mywiki: %s", output)
	}
}

// TestWikiListWikis tests wiki-list subcommand
func TestWikiListWikis(t *testing.T) {
	os.RemoveAll(testWikiData)
	defer os.RemoveAll(testWikiData)

	// Should have default wiki
	output, err := runWiki("wiki-list")
	if err != nil {
		t.Fatalf("wiki-list failed: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(output, "default") {
		t.Errorf("Expected default wiki in list, got: %s", output)
	}

	// Create additional wikis
	for _, name := range []string{"docs", "notes"} {
		_, err := runWiki("wiki-create", name)
		if err != nil {
			t.Fatalf("wiki-create %s failed: %v", name, err)
		}
	}

	output, err = runWiki("wiki-list")
	if err != nil {
		t.Fatalf("wiki-list failed: %v\nOutput: %s", err, output)
	}

	for _, name := range []string{"default", "docs", "notes"} {
		if !strings.Contains(output, name) {
			t.Errorf("Expected %s in wiki list, got: %s", name, output)
		}
	}
}

// TestWikiVerify tests verify subcommand
func TestWikiVerify(t *testing.T) {
	os.RemoveAll(testWikiData)
	defer os.RemoveAll(testWikiData)

	// Verify on empty wiki (default created by EnsureWikiExists)
	output, err := runWiki("verify")
	if err != nil {
		t.Fatalf("verify failed: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(output, "Index verification OK") {
		t.Errorf("Expected verification OK, got: %s", output)
	}
	if !strings.Contains(output, "Document count: 0") {
		t.Errorf("Expected 0 documents, got: %s", output)
	}

	// Create an article and verify count updates
	_, err = runWiki("create", "test-art", "--content", "# Test")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	output, err = runWiki("verify")
	if err != nil {
		t.Fatalf("verify after create failed: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(output, "Document count: 1") {
		t.Errorf("Expected 1 document, got: %s", output)
	}

	// Verify non-existent wiki
	_, err = runWiki("verify", "--wiki", "no-such")
	if err == nil {
		t.Error("expected error verifying non-existent wiki")
	}
}
