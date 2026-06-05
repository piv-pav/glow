package tests

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestWikiCreate(t *testing.T) {
	t.Cleanup(func() { runWiki("delete", "test-create") })

	tests := []struct {
		name    string
		args    []string
		wantErr bool
		check   func(t *testing.T, output string)
	}{
		{
			name: "create with content",
			args: []string{"create", "test-create", "--content", "Hello World"},
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "Created article: test-create") {
					t.Errorf("Expected success message, got: %s", output)
				}

				content := readArticle(t, "test-create")
				if !strings.Contains(content, "Hello World") {
					t.Errorf("Expected content 'Hello World', got: %s", content)
				}
			},
		},
		{
			name: "create with metadata",
			args: []string{"create", "test-meta", "--content", "Test", "--meta", "project:glow", "--meta", "tags:go,cli"},
			check: func(t *testing.T, output string) {
				content := readArticle(t, "test-meta")
				if !strings.Contains(content, "project: glow") {
					t.Errorf("Expected project metadata")
				}
				if !strings.Contains(content, "- go") && !strings.Contains(content, "tags: go") {
					t.Errorf("Expected tags metadata")
				}
			},
		},
		{
			name:    "create without content fails",
			args:    []string{"create", "test-fail"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := runWiki(tt.args...)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v\nOutput: %s", err, output)
			}

			if tt.check != nil {
				tt.check(t, output)
			}
		})
	}
}

func TestWikiAppend(t *testing.T) {
	// Setup: create article
	_, err := runWiki("create", "test-append", "--content", "Initial content")
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	t.Cleanup(func() { runWiki("delete", "test-append") })

	tests := []struct {
		name  string
		args  []string
		check func(t *testing.T, content string)
	}{
		{
			name: "append content",
			args: []string{"append", "test-append", "--content", "Appended text"},
			check: func(t *testing.T, content string) {
				if !strings.Contains(content, "Initial content") {
					t.Error("Original content missing")
				}
				if !strings.Contains(content, "Appended text") {
					t.Error("Appended content missing")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := runWiki(tt.args...)
			if err != nil {
				t.Fatalf("Command failed: %v\nOutput: %s", err, output)
			}

			content := readArticle(t, "test-append")
			tt.check(t, content)
		})
	}
}

func TestWikiAppendSection(t *testing.T) {
	content := `# Header
Some content

## Section 1
Section 1 content

## Section 2
Section 2 content`

	_, err := runWiki("create", "test-section-append", "--content", content)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	t.Cleanup(func() { runWiki("delete", "test-section-append") })

	_, err = runWiki("append", "test-section-append", "--section", "Section 1", "--content", "Appended to section 1")
	if err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	result := readArticle(t, "test-section-append")
	
	if !strings.Contains(result, "Section 1 content") {
		t.Error("Original section content missing")
	}
	if !strings.Contains(result, "Appended to section 1") {
		t.Error("Appended content missing from section")
	}
	if !strings.Contains(result, "Section 2 content") {
		t.Error("Other sections should remain intact")
	}

	// Appending to non-existent section must error
	out, err := runWiki("append", "test-section-append", "--section", "Nonexistent", "--content", "x")
	if err == nil {
		t.Errorf("Expected error appending to missing section, got: %s", out)
	}
}

func TestWikiUpdate(t *testing.T) {
	_, err := runWiki("create", "test-update", "--content", "Original content")
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	t.Cleanup(func() { runWiki("delete", "test-update") })

	_, err = runWiki("update", "test-update", "--content", "Updated content")
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	content := readArticle(t, "test-update")
	if strings.Contains(content, "Original content") {
		t.Error("Old content should be replaced")
	}
	if !strings.Contains(content, "Updated content") {
		t.Error("New content missing")
	}
}

func TestWikiUpdateMeta(t *testing.T) {
	_, err := runWiki("create", "test-update-meta", "--content", "Test", "--meta", "version:1")
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	t.Cleanup(func() { runWiki("delete", "test-update-meta") })

	_, err = runWiki("meta", "set", "test-update-meta", "version", "2")
	if err != nil {
		t.Fatalf("Meta set version failed: %v", err)
	}
	_, err = runWiki("meta", "set", "test-update-meta", "status", "ready")
	if err != nil {
		t.Fatalf("Meta set status failed: %v", err)
	}

	content := readArticle(t, "test-update-meta")
	if !strings.Contains(content, "version: 2") && !strings.Contains(content, "version: \"2\"") {
		t.Error("Version metadata not updated")
	}
	if !strings.Contains(content, "status: ready") {
		t.Error("Status metadata not added")
	}
}

func TestWikiDelete(t *testing.T) {
	_, err := runWiki("create", "test-delete", "--content", "To be deleted")
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	output, err := runWiki("delete", "test-delete")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if !strings.Contains(output, "Deleted article: test-delete") {
		t.Errorf("Expected success message, got: %s", output)
	}

	// Verify file is gone
	_, err = runWiki("read", "test-delete")
	if err == nil {
		t.Error("Article should not exist after delete")
	}
}

func TestWikiDeleteSection(t *testing.T) {
	content := `## Section 1
Content 1

## Section 2
Content 2

## Section 3
Content 3`

	_, err := runWiki("create", "test-delete-section", "--content", content)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	t.Cleanup(func() { runWiki("delete", "test-delete-section") })

	_, err = runWiki("delete", "test-delete-section", "--section", "Section 2")
	if err != nil {
		t.Fatalf("Delete section failed: %v", err)
	}

	result := readArticle(t, "test-delete-section")
	
	if !strings.Contains(result, "Section 1") || !strings.Contains(result, "Content 1") {
		t.Error("Section 1 should remain")
	}
	if strings.Contains(result, "Section 2") || strings.Contains(result, "Content 2") {
		t.Error("Section 2 should be deleted")
	}
	if !strings.Contains(result, "Section 3") || !strings.Contains(result, "Content 3") {
		t.Error("Section 3 should remain")
	}
}

func TestWikiMove(t *testing.T) {
	_, err := runWiki("create", "test-move-old", "--content", "Move me")
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	t.Cleanup(func() {
		runWiki("delete", "test-move-old")
		runWiki("delete", "test-move-new")
	})

	output, err := runWiki("move", "test-move-old", "test-move-new")
	if err != nil {
		t.Fatalf("Move failed: %v", err)
	}

	if !strings.Contains(output, "Moved article: test-move-old -> test-move-new") {
		t.Errorf("Expected success message, got: %s", output)
	}

	// Old should not exist
	_, err = runWiki("read", "test-move-old")
	if err == nil {
		t.Error("Old article should not exist")
	}

	// New should exist
	content := readArticle(t, "test-move-new")
	if !strings.Contains(content, "Move me") {
		t.Error("Content not preserved in move")
	}
}

func TestWikiAppendStdin(t *testing.T) {
	_, err := runWiki("create", "test-stdin", "--content", "Initial")
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	t.Cleanup(func() { runWiki("delete", "test-stdin") })

	// Test append via stdin
	cmd := exec.Command("glow", "append", "test-stdin", "--stdin")
	cmd.Env = append(os.Environ(), "GLOW_DATA="+testWikiData)
	cmd.Stdin = strings.NewReader("Stdin content")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Append stdin failed: %v\nOutput: %s", err, output)
	}

	content := readArticle(t, "test-stdin")
	if !strings.Contains(content, "Initial") {
		t.Error("Original content missing")
	}
	if !strings.Contains(content, "Stdin content") {
		t.Error("Stdin content not appended")
	}
}

func TestContentMultiline(t *testing.T) {
	tests := []struct {
		name    string
		content string // as passed to --content
		want    string // expected in stored article
	}{
		{
			name:    "literal newlines in --content",
			content: "line1\nline2\nline3",
			want:    "line1\nline2\nline3",
		},
		{
			name:    "escaped \\n in --content",
			content: `line1\nline2\nline3`,
			want:    "line1\nline2\nline3",
		},
		{
			name:    "escaped \\t in --content",
			content: `col1\tcol2`,
			want:    "col1\tcol2",
		},
		{
			name:    "markdown with literal newlines",
			content: "# Title\n\nParagraph one.\n\nParagraph two.",
			want:    "# Title\n\nParagraph one.\n\nParagraph two.",
		},
		{
			name:    "backslash preserved",
			content: `path\\to\\file`,
			want:    `path\to\file`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			article := "test-multiline-" + strings.ReplaceAll(tt.name, " ", "-")
			t.Cleanup(func() { runWiki("delete", article) })

			out, err := runWiki("create", article, "--content", tt.content)
			if err != nil {
				t.Fatalf("create failed: %v\nOutput: %s", err, out)
			}

			stored := readArticle(t, article)
			if !strings.Contains(stored, tt.want) {
				t.Errorf("expected content to contain:\n%q\ngot:\n%s", tt.want, stored)
			}
		})
	}
}
