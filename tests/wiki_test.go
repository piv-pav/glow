package tests

import (
	"testing"
)

func TestWikiCreateWiki(t *testing.T) {
	for _, backend := range backends {
		t.Run(backend, func(t *testing.T) {
			e := newEnv(t, backend)

			// Create a second wiki with same backend
			e.initWiki(t, "otherwiki", backend)
			out := e.mustRunGlobal(t, "wiki-list")
			assertContains(t, out, "otherwiki")

			// Duplicate wiki should fail
			_, err := e.runGlobal("init", "otherwiki")
			if err == nil {
				t.Error("expected error on duplicate wiki creation")
			}

			// Articles in named wiki are isolated from testwiki
			e.mustRunGlobal(t, "--wiki", "otherwiki", "create", "isolated", "--content", "only in otherwiki")

			out = e.mustRun(t, "list")
			assertNotContains(t, out, "isolated")

			out = e.mustRunGlobal(t, "--wiki", "otherwiki", "list")
			assertContains(t, out, "isolated")
		})
	}
}

func TestWikiListWikis(t *testing.T) {
	for _, backend := range backends {
		t.Run(backend, func(t *testing.T) {
			e := newEnv(t, backend)

			for _, name := range []string{"docs", "notes"} {
				e.initWiki(t, name, backend)
			}

			out := e.mustRunGlobal(t, "wiki-list")
			for _, name := range []string{"testwiki", "docs", "notes"} {
				assertContains(t, out, name)
			}
			assertContains(t, out, backend)
		})
	}
}

func TestWikiExportImport(t *testing.T) {
	for _, srcBackend := range backends {
		for _, dstBackend := range backends {
			t.Run(srcBackend+"_to_"+dstBackend, func(t *testing.T) {
				src := newEnv(t, srcBackend)
				src.mustRun(t, "create", "foo", "--content", "# Foo\n\nHello", "--tag", "test")
				src.mustRun(t, "create", "bar", "--content", "# Bar\n\nWorld")

				archive := t.TempDir() + "/export.tar.gz"
				src.mustRunGlobal(t, "export", src.wiki, archive)

				dst := newEnv(t, dstBackend)
				dst.mustRunGlobal(t, "import", dst.wiki, archive)

				out := dst.mustRun(t, "list")
				assertContains(t, out, "foo")
				assertContains(t, out, "bar")

				out = dst.readArticle(t, "foo")
				assertContains(t, out, "Hello")

				// Re-import — duplicates skipped, not error
				out, err := dst.runGlobal("import", dst.wiki, archive)
				if err != nil {
					t.Fatalf("re-import failed: %v\noutput: %s", err, out)
				}
				assertContains(t, out, "skipped")
			})
		}
	}
}
