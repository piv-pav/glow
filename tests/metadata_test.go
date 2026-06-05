package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWikiTags(t *testing.T) {
	os.RemoveAll(testWikiData)
	defer os.RemoveAll(testWikiData)

	// Setup: create article with tags
	articleName := "tag-test"
	content := "---\ntags:\n  - initial\n---\n\n# Tag Test\n\nContent.\n"

	os.MkdirAll(filepath.Join(testWikiData, "default", "articles"), 0755)
	err := os.WriteFile(getArticlePath(articleName), []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to setup test article: %v", err)
	}

	runWiki("rebuild")

	// Test adding tags
	_, err = runWiki("update", articleName, "--tag", "added")
	if err != nil {
		t.Fatalf("Failed to add tag: %v", err)
	}

	updated := readArticle(t, articleName)
	if !strings.Contains(updated, "added") {
		t.Error("Tag add failed")
	}
	if !strings.Contains(updated, "initial") {
		t.Error("Original tag should remain")
	}

	// Test removing tags
	_, err = runWiki("update", articleName, "--untag", "initial")
	if err != nil {
		t.Fatalf("Failed to remove tag: %v", err)
	}

	updated = readArticle(t, articleName)
	if strings.Contains(updated, "initial") {
		t.Error("Tag remove failed")
	}
	if !strings.Contains(updated, "added") {
		t.Error("Other tags should remain")
	}
}

func TestWikiCreateWithTags(t *testing.T) {
	os.RemoveAll(testWikiData)
	defer os.RemoveAll(testWikiData)

	// Create with comma-separated tags
	_, err := runWiki("create", "tag-create-test", "--content", "Hello", "--tag", "go,cli")
	if err != nil {
		t.Fatalf("Failed to create: %v", err)
	}
	t.Cleanup(func() { runWiki("delete", "tag-create-test") })

	content := readArticle(t, "tag-create-test")
	if !strings.Contains(content, "go") || !strings.Contains(content, "cli") {
		t.Errorf("Expected both tags, got: %s", content)
	}
}
