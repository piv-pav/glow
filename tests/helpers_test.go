package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// testEnv holds isolated data and config directories for one test run.
type testEnv struct {
	data    string
	config  string
	wiki    string // default wiki name for this env
	backend string
}

func newEnv(t *testing.T, backend string) *testEnv {
	t.Helper()
	dir := t.TempDir()
	e := &testEnv{
		data:    filepath.Join(dir, "data"),
		config:  filepath.Join(dir, "config"),
		wiki:    "testwiki",
		backend: backend,
	}
	e.initWiki(t, e.wiki, backend)
	return e
}

// run executes a glow command in this environment, defaulting --wiki to e.wiki.
func (e *testEnv) run(args ...string) (string, error) {
	// Inject --wiki if not already present
	hasWiki := false
	for _, a := range args {
		if a == "--wiki" || a == "-w" {
			hasWiki = true
			break
		}
	}
	if !hasWiki {
		args = append([]string{"--wiki", e.wiki}, args...)
	}
	cmd := exec.Command("glow", args...)
	cmd.Env = append(os.Environ(),
		"GLOW_DATA="+e.data,
		"GLOW_CONFIG="+filepath.Join(e.config, "glow.yaml"),
	)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// runStdin executes a glow command feeding stdin, defaulting --wiki to e.wiki.
func (e *testEnv) runStdin(stdin string, args ...string) (string, error) {
	hasWiki := false
	for _, a := range args {
		if a == "--wiki" || a == "-w" {
			hasWiki = true
			break
		}
	}
	if !hasWiki {
		args = append([]string{"--wiki", e.wiki}, args...)
	}
	cmd := exec.Command("glow", args...)
	cmd.Env = append(os.Environ(),
		"GLOW_DATA="+e.data,
		"GLOW_CONFIG="+filepath.Join(e.config, "glow.yaml"),
	)
	cmd.Stdin = strings.NewReader(stdin)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// mustRunStdin fails the test if the command errors.
func (e *testEnv) mustRunStdin(t *testing.T, stdin string, args ...string) string {
	t.Helper()
	out, err := e.runStdin(stdin, args...)
	if err != nil {
		t.Fatalf("glow %s failed: %v\nOutput: %s", strings.Join(args, " "), err, out)
	}
	return out
}

// runGlobal runs a command without injecting --wiki (for wiki-list, export, import, init).
func (e *testEnv) runGlobal(args ...string) (string, error) {
	cmd := exec.Command("glow", args...)
	cmd.Env = append(os.Environ(),
		"GLOW_DATA="+e.data,
		"GLOW_CONFIG="+filepath.Join(e.config, "glow.yaml"),
	)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// mustRun fails the test if the command errors.
func (e *testEnv) mustRun(t *testing.T, args ...string) string {
	t.Helper()
	out, err := e.run(args...)
	if err != nil {
		t.Fatalf("glow %s failed: %v\nOutput: %s", strings.Join(args, " "), err, out)
	}
	return out
}

// mustRunGlobal fails the test if the global command errors.
func (e *testEnv) mustRunGlobal(t *testing.T, args ...string) string {
	t.Helper()
	out, err := e.runGlobal(args...)
	if err != nil {
		t.Fatalf("glow %s failed: %v\nOutput: %s", strings.Join(args, " "), err, out)
	}
	return out
}

// initWiki creates a wiki with the given backend non-interactively.
func (e *testEnv) initWiki(t *testing.T, name, backend string) {
	t.Helper()
	cmd := exec.Command("glow", "init", name)
	cmd.Env = append(os.Environ(),
		"GLOW_DATA="+e.data,
		"GLOW_CONFIG="+filepath.Join(e.config, "glow.yaml"),
	)
	cmd.Stdin = strings.NewReader("\n") // accept sqlite default
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("init wiki %s (%s) failed: %v\nOutput: %s", name, backend, err, string(out))
	}
}

// readArticle reads article content via glow read (backend-agnostic).
func (e *testEnv) readArticle(t *testing.T, name string) string {
	t.Helper()
	return e.mustRun(t, "read", name)
}

// assertContains fails if haystack doesn't contain needle.
func assertContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Errorf("expected output to contain %q\ngot:\n%s", needle, haystack)
	}
}

// assertNotContains fails if haystack contains needle.
func assertNotContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if strings.Contains(haystack, needle) {
		t.Errorf("expected output NOT to contain %q\ngot:\n%s", needle, haystack)
	}
}

// backends lists the backends covered by integration tests.
var backends = []string{"sqlite"}

// TestMain just ensures the binary is available.
func TestMain(m *testing.M) {
	if _, err := exec.LookPath("glow"); err != nil {
		// Built by `just test` which runs `just build` first
		os.Exit(1)
	}
	os.Exit(m.Run())
}
