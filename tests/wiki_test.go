package tests

import (
	"os"
	"strings"
	"testing"
)

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
