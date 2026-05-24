package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
		name    string
		key     string
		want    string
		wantErr bool
	}{
		{
			name:    "get scalar metadata",
			key:     "author",
			want:    "test",
			wantErr: false,
		},
		{
			name:    "get array metadata",
			key:     "tags",
			want:    "one, two",
			wantErr: false,
		},
		{
			name:    "get non-existent key",
			key:     "missing",
			want:    "",
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
