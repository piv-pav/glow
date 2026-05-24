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

// assertContains checks if haystack contains needle
func assertContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Errorf("Expected output to contain %q, got:\n%s", needle, haystack)
	}
}

// assertNotContains checks if haystack doesn't contain needle
func assertNotContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if strings.Contains(haystack, needle) {
		t.Errorf("Expected output to NOT contain %q, got:\n%s", needle, haystack)
	}
}
