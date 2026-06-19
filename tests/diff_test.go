package tests

import "testing"

const diffStart = "<<<<<<< SEARCH\n"
const diffMid = "\n=======\n"
const diffEnd = "\n>>>>>>> REPLACE\n"

// block builds a single SEARCH/REPLACE block.
func block(search, replace string) string {
	return diffStart + search + diffMid + replace + diffEnd
}

func TestWikiUpdateDiff(t *testing.T) {
	for _, backend := range backends {
		t.Run(backend, func(t *testing.T) {
			e := newEnv(t, backend)

			// --- whole-article single block ---
			e.mustRun(t, "create", "diff-single", "--content", "alpha\nbeta\ngamma", "--tag", "x")
			out := e.mustRunStdin(t, block("beta", "BETA"), "update", "diff-single", "--diff")
			assertContains(t, out, "Applied 1 diff block(s) to article: diff-single")
			got := e.readArticle(t, "diff-single")
			assertContains(t, got, "alpha\nBETA\ngamma")

			// --- whole-article multi block, applied in order ---
			e.mustRun(t, "create", "diff-multi", "--content", "# Title\n\none\ntwo\nthree", "--tag", "x")
			diff := block("# Title", "# New") + "\n" + block("two", "TWO\nextra")
			out = e.mustRunStdin(t, diff, "update", "diff-multi", "--diff")
			assertContains(t, out, "Applied 2 diff block(s)")
			got = e.readArticle(t, "diff-multi")
			assertContains(t, got, "# New\n\none\nTWO\nextra\nthree")

			// --- diff combined with tags ---
			e.mustRunStdin(t, block("alpha", "ALPHA"), "update", "diff-single", "--diff", "--tag", "added")
			tags := e.mustRun(t, "read", "diff-single", "--tags")
			assertContains(t, tags, "added")
		})
	}
}

func TestWikiUpdateDiffSection(t *testing.T) {
	for _, backend := range backends {
		t.Run(backend, func(t *testing.T) {
			e := newEnv(t, backend)

			// "value" appears in BOTH sections: globally ambiguous, unique per-section.
			content := "# Doc\n\n## Alpha\nvalue\nkeep\n\n## Beta\nvalue\nkeep"
			e.mustRun(t, "create", "diff-sec", "--content", content, "--tag", "x")

			out := e.mustRunStdin(t, block("value", "BETA-VALUE"), "update", "diff-sec", "--diff", "--section", "Beta")
			assertContains(t, out, `Applied 1 diff block(s) to section "Beta" in article: diff-sec`)

			got := e.readArticle(t, "diff-sec")
			// Beta edited, Alpha's "value" untouched.
			assertContains(t, got, "## Alpha\nvalue\nkeep")
			assertContains(t, got, "## Beta\nBETA-VALUE\nkeep")
		})
	}
}

func TestWikiUpdateDiffErrors(t *testing.T) {
	for _, backend := range backends {
		t.Run(backend, func(t *testing.T) {
			e := newEnv(t, backend)
			// "x" appears twice -> ambiguous whole-article.
			e.mustRun(t, "create", "diff-err", "--content", "x\ny\nx", "--tag", "x")
			orig := e.readArticle(t, "diff-err")

			cases := []struct {
				name    string
				stdin   string
				args    []string
				wantOut string
			}{
				{
					name:    "search not found",
					stdin:   block("nope", "z"),
					args:    []string{"update", "diff-err", "--diff"},
					wantOut: "SEARCH text not found",
				},
				{
					name:    "ambiguous match",
					stdin:   block("x", "Z"),
					args:    []string{"update", "diff-err", "--diff"},
					wantOut: "matches 2 times",
				},
				{
					name:    "malformed missing separator",
					stdin:   "<<<<<<< SEARCH\nx\n>>>>>>> REPLACE\n",
					args:    []string{"update", "diff-err", "--diff"},
					wantOut: `missing "======="`,
				},
				{
					name:    "empty stdin",
					stdin:   "",
					args:    []string{"update", "diff-err", "--diff"},
					wantOut: "no SEARCH/REPLACE blocks found",
				},
				{
					name:    "section not found",
					stdin:   block("x", "y"),
					args:    []string{"update", "diff-err", "--diff", "--section", "Nope"},
					wantOut: "section not found",
				},
				{
					name:    "diff with --content rejected",
					stdin:   "",
					args:    []string{"update", "diff-err", "--diff", "--content", "foo"},
					wantOut: "cannot be combined with --content or --stdin",
				},
				{
					name:    "diff with --stdin rejected",
					stdin:   "anything",
					args:    []string{"update", "diff-err", "--diff", "--stdin"},
					wantOut: "cannot be combined with --content or --stdin",
				},
			}

			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					out, err := e.runStdin(tc.stdin, tc.args...)
					if err == nil {
						t.Fatalf("expected error, got success:\n%s", out)
					}
					assertContains(t, out, tc.wantOut)
				})
			}

			// Atomicity: after all failed edits, the article is unchanged.
			if got := e.readArticle(t, "diff-err"); got != orig {
				t.Fatalf("article changed after failed diffs:\ngot:  %q\nwant: %q", got, orig)
			}
		})
	}
}
