package tests

import (
	"strings"
	"testing"
)

func TestWikiTags(t *testing.T) {
	for _, backend := range backends {
		t.Run(backend, func(t *testing.T) {
			e := newEnv(t, backend)
			e.mustRun(t, "create", "tag-test", "--content", "# Tag Test\n\nContent.", "--tag", "initial")

			// Add tag
			e.mustRun(t, "update", "tag-test", "--tag", "added")
			out, _ := e.run("read", "tag-test", "--tags")
			assertContains(t, out, "initial")
			assertContains(t, out, "added")

			// Remove tag
			e.mustRun(t, "update", "tag-test", "--untag", "initial")
			out, _ = e.run("read", "tag-test", "--tags")
			assertNotContains(t, out, "initial")
			assertContains(t, out, "added")

			// Tag deduplication
			e.mustRun(t, "update", "tag-test", "--tag", "added")
			out, _ = e.run("read", "tag-test", "--tags")
			if strings.Count(out, "added") > 1 {
				t.Errorf("tag 'added' duplicated in output:\n%s", out)
			}
		})
	}
}
