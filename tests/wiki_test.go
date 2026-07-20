package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
			_, err := e.runGlobal("wiki-create", "otherwiki", "-b", backend)
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

func TestWikiCreateValidation(t *testing.T) {
	e := newEnv(t, "sqlite")

	// 4.3: wiki-create <name> with no flags → error mentioning -i or -b
	t.Run("no_mode_flag_error", func(t *testing.T) {
		out, err := e.runGlobal("wiki-create", "newwiki")
		if err == nil {
			t.Fatal("expected non-zero exit when neither -i nor -b provided")
		}
		if !strings.Contains(out, "-i") && !strings.Contains(out, "-b") &&
			!strings.Contains(out, "--interactive") && !strings.Contains(out, "--backend") {
			t.Errorf("expected error to mention -i or -b, got: %s", out)
		}
	})

	// 4.4: wiki-create <name> -i -b sqlite → mutually exclusive error
	t.Run("interactive_and_backend_error", func(t *testing.T) {
		out, err := e.runGlobal("wiki-create", "newwiki", "-i", "-b", "sqlite")
		if err == nil {
			t.Fatal("expected non-zero exit when -i and -b both provided")
		}
		_ = out
	})

	// 4.5: wiki-create with no name → cobra ExactArgs(1) error
	t.Run("missing_name_error", func(t *testing.T) {
		_, err := e.runGlobal("wiki-create", "-b", "sqlite")
		if err == nil {
			t.Fatal("expected non-zero exit when name omitted")
		}
	})

	// 4.6: wiki-create <name> -b sqlite when wiki already exists → error
	t.Run("duplicate_wiki_error", func(t *testing.T) {
		e2 := newEnv(t, "sqlite")
		// testwiki already created by newEnv
		out, err := e2.runGlobal("wiki-create", e2.wiki, "-b", "sqlite")
		if err == nil {
			t.Fatalf("expected non-zero exit on duplicate wiki, got output: %s", out)
		}
	})

	// 4.8: wiki-create <name> -b rqlite without --url → error mentioning --url
	t.Run("rqlite_missing_url_error", func(t *testing.T) {
		out, err := e.runGlobal("wiki-create", "rqwiki", "-b", "rqlite")
		if err == nil {
			t.Fatal("expected non-zero exit when --url omitted for rqlite")
		}
		assertContains(t, out, "--url")
	})

	// 4.9: --password and --password-stdin together → mutually exclusive error
	t.Run("password_and_stdin_error", func(t *testing.T) {
		cmd := exec.Command("glow", "wiki-create", "rqwiki", "-b", "rqlite",
			"--url", "http://localhost:14001", "--password", "secret", "--password-stdin")
		cmd.Env = append(os.Environ(),
			"GLOW_DATA="+e.data,
			"GLOW_CONFIG="+filepath.Join(e.config, "glow.yaml"),
		)
		cmd.Stdin = strings.NewReader("secret\n")
		out, err := cmd.CombinedOutput()
		if err == nil {
			t.Fatalf("expected non-zero exit when --password and --password-stdin both provided, got: %s", out)
		}
	})
}
