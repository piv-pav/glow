package article

import "testing"

func TestApplyDiff_SingleBlock(t *testing.T) {
	content := "alpha\nbeta\ngamma"
	diff := "<<<<<<< SEARCH\nbeta\n=======\nBETA\n>>>>>>> REPLACE"
	got, n, err := ApplyDiff(content, diff)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("want 1 block, got %d", n)
	}
	if got != "alpha\nBETA\ngamma" {
		t.Fatalf("unexpected result: %q", got)
	}
}

func TestApplyDiff_MultiBlock(t *testing.T) {
	content := "# Title\n\none\ntwo\nthree"
	diff := "<<<<<<< SEARCH\n# Title\n=======\n# New\n>>>>>>> REPLACE\n\n<<<<<<< SEARCH\ntwo\n=======\nTWO\nextra\n>>>>>>> REPLACE"
	got, n, err := ApplyDiff(content, diff)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("want 2 blocks, got %d", n)
	}
	want := "# New\n\none\nTWO\nextra\nthree"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestApplyDiff_NotFound(t *testing.T) {
	_, _, err := ApplyDiff("abc", "<<<<<<< SEARCH\nxyz\n=======\nq\n>>>>>>> REPLACE")
	if err == nil {
		t.Fatal("expected error for missing search text")
	}
}

func TestApplyDiff_Ambiguous(t *testing.T) {
	_, _, err := ApplyDiff("x\nx", "<<<<<<< SEARCH\nx\n=======\ny\n>>>>>>> REPLACE")
	if err == nil {
		t.Fatal("expected error for ambiguous match")
	}
}

func TestApplyDiff_EmptySearchPrepends(t *testing.T) {
	got, _, err := ApplyDiff("body", "<<<<<<< SEARCH\n=======\nheader\n>>>>>>> REPLACE")
	if err != nil {
		t.Fatal(err)
	}
	if got != "headerbody" {
		t.Fatalf("unexpected result: %q", got)
	}
}

func TestParseDiff_Malformed(t *testing.T) {
	if _, err := ParseDiff("<<<<<<< SEARCH\nfoo\n>>>>>>> REPLACE"); err == nil {
		t.Fatal("expected error for missing separator")
	}
	if _, err := ParseDiff("no markers here"); err == nil {
		t.Fatal("expected error for no blocks")
	}
}

func TestApplyDiff_MultilineSearch(t *testing.T) {
	content := "a\nb\nc\nd"
	diff := "<<<<<<< SEARCH\nb\nc\n=======\nB\nC\n>>>>>>> REPLACE"
	got, _, err := ApplyDiff(content, diff)
	if err != nil {
		t.Fatal(err)
	}
	if got != "a\nB\nC\nd" {
		t.Fatalf("unexpected result: %q", got)
	}
}

func TestApplyDiffToSection_ScopesToSection(t *testing.T) {
	a := &Article{Content: "# Doc\n\n## Alpha\nvalue\n\n## Beta\nvalue"}
	// "value" is globally ambiguous but unique within Beta.
	n, err := a.ApplyDiffToSection("Beta", "<<<<<<< SEARCH\nvalue\n=======\nBV\n>>>>>>> REPLACE")
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("want 1 block, got %d", n)
	}
	want := "# Doc\n\n## Alpha\nvalue\n\n## Beta\nBV"
	if a.Content != want {
		t.Fatalf("got %q want %q", a.Content, want)
	}
}

func TestApplyDiffToSection_SectionNotFound(t *testing.T) {
	a := &Article{Content: "## Alpha\nx"}
	if _, err := a.ApplyDiffToSection("Beta", "<<<<<<< SEARCH\nx\n=======\ny\n>>>>>>> REPLACE"); err == nil {
		t.Fatal("expected section-not-found error")
	}
}
