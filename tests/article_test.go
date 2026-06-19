package tests

import (
	"testing"
)

func TestWikiCreate(t *testing.T) {
	for _, backend := range backends {
		t.Run(backend, func(t *testing.T) {
			e := newEnv(t, backend)

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
						assertContains(t, output, "Created article: test-create")
						content := e.readArticle(t, "test-create")
						assertContains(t, content, "Hello World")
					},
				},
				{
					name: "create with tags",
					args: []string{"create", "test-tags", "--content", "Test", "--tag", "go", "--tag", "cli"},
					check: func(t *testing.T, output string) {
						out, _ := e.run("read", "test-tags", "--tags")
						assertContains(t, out, "go")
						assertContains(t, out, "cli")
					},
				},
				{
					name: "create with comma-separated tags",
					args: []string{"create", "test-comma-tags", "--content", "Test", "--tag", "go,cli"},
					check: func(t *testing.T, output string) {
						out, _ := e.run("read", "test-comma-tags", "--tags")
						assertContains(t, out, "go")
						assertContains(t, out, "cli")
					},
				},
				{
					name:    "create without content fails",
					args:    []string{"create", "test-fail"},
					wantErr: true,
				},
				{
					name: "create nested path",
					args: []string{"create", "folder/nested", "--content", "Nested article"},
					check: func(t *testing.T, output string) {
						assertContains(t, output, "Created article: folder/nested")
						content := e.readArticle(t, "folder/nested")
						assertContains(t, content, "Nested article")
					},
				},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					output, err := e.run(tt.args...)
					if tt.wantErr {
						if err == nil {
							t.Errorf("expected error, got none; output: %s", output)
						}
						return
					}
					if err != nil {
						t.Fatalf("unexpected error: %v\noutput: %s", err, output)
					}
					if tt.check != nil {
						tt.check(t, output)
					}
				})
			}
		})
	}
}

func TestWikiReadUpdateDelete(t *testing.T) {
	for _, backend := range backends {
		t.Run(backend, func(t *testing.T) {
			e := newEnv(t, backend)
			e.mustRun(t, "create", "myarticle", "--content", "# Hello\n\nOriginal content")

			out := e.readArticle(t, "myarticle")
			assertContains(t, out, "Original content")

			e.mustRun(t, "update", "myarticle", "--content", "# Hello\n\nUpdated content")
			out = e.readArticle(t, "myarticle")
			assertContains(t, out, "Updated content")
			assertNotContains(t, out, "Original content")

			e.mustRun(t, "delete", "myarticle")
			_, err := e.run("read", "myarticle")
			if err == nil {
				t.Error("expected error reading deleted article")
			}
		})
	}
}

func TestWikiMove(t *testing.T) {
	for _, backend := range backends {
		t.Run(backend, func(t *testing.T) {
			e := newEnv(t, backend)
			e.mustRun(t, "create", "original", "--content", "Move me")

			e.mustRun(t, "move", "original", "moved")

			out := e.readArticle(t, "moved")
			assertContains(t, out, "Move me")

			_, err := e.run("read", "original")
			if err == nil {
				t.Error("expected error reading moved article at old path")
			}
		})
	}
}

func TestWikiListArticles(t *testing.T) {
	for _, backend := range backends {
		t.Run(backend, func(t *testing.T) {
			e := newEnv(t, backend)
			e.mustRun(t, "create", "alpha", "--content", "A")
			e.mustRun(t, "create", "beta", "--content", "B")
			e.mustRun(t, "create", "gamma", "--content", "C")

			out := e.mustRun(t, "list")
			assertContains(t, out, "alpha")
			assertContains(t, out, "beta")
			assertContains(t, out, "gamma")
		})
	}
}

func TestWikiAppend(t *testing.T) {
	for _, backend := range backends {
		t.Run(backend, func(t *testing.T) {
			e := newEnv(t, backend)
			e.mustRun(t, "create", "appendme", "--content", "# Hello\n\nFirst.")

			e.mustRun(t, "append", "appendme", "--content", "\n\nSecond.")

			out := e.readArticle(t, "appendme")
			assertContains(t, out, "First.")
			assertContains(t, out, "Second.")
		})
	}
}

func TestWikiSections(t *testing.T) {
	for _, backend := range backends {
		t.Run(backend, func(t *testing.T) {
			e := newEnv(t, backend)
			e.mustRun(t, "create", "sectioned", "--content", "# Title\n\nIntro.\n\n## Section A\n\nContent A.\n\n## Section B\n\nContent B.")

			out := e.mustRun(t, "read", "sectioned", "--section", "Section A")
			assertContains(t, out, "Content A.")
			assertNotContains(t, out, "Content B.")

			out = e.mustRun(t, "read", "sectioned", "--sections")
			assertContains(t, out, "Section A")
			assertContains(t, out, "Section B")

			e.mustRun(t, "delete", "sectioned", "--section", "Section B")
			out = e.readArticle(t, "sectioned")
			assertContains(t, out, "Content A.")
			assertNotContains(t, out, "Content B.")
		})
	}
}
