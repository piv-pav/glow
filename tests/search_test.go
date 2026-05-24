package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

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
		name        string
		args        []string
		wantInside  []string
		wantOutside []string
	}{
		{
			name:        "search by content",
			args:        []string{"search", "golang", "-l", "10"},
			wantInside:  []string{"golang-basics"},
			wantOutside: []string{"python"},
		},
		{
			name:        "search by tag",
			args:        []string{"search", "tags:go", "-l", "10"},
			wantInside:  []string{"golang-basics", "cli-tools"},
			wantOutside: []string{"python-intro"},
		},
		{
			name:        "search by project",
			args:        []string{"search", "project:tools", "-l", "10"},
			wantInside:  []string{"cli-tools"},
			wantOutside: []string{"golang-basics", "python-intro"},
		},
		{
			name:        "search multiple tags",
			args:        []string{"search", "tags:cli", "-l", "10"},
			wantInside:  []string{"cli-tools"},
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
		_, err := runWiki("create", name, "--content", "# "+name+"\\n\\nContent.")
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
