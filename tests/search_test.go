package tests

import (
	"testing"
)

func TestWikiSearch(t *testing.T) {
	for _, backend := range backends {
		t.Run(backend, func(t *testing.T) {
			e := newEnv(t, backend)

			e.mustRun(t, "create", "golang-basics", "--content", "# Go Basics\n\nLearn Golang fundamentals.", "--tag", "go", "--tag", "programming")
			e.mustRun(t, "create", "python-intro", "--content", "# Python Intro\n\nPython programming language.", "--tag", "python", "--tag", "programming")
			e.mustRun(t, "create", "cli-tools", "--content", "# CLI Tools\n\nBuilding CLI with Go.", "--tag", "go", "--tag", "cli")

			// Files backend needs index built; sqlite searches natively
			if backend == "files" {
				e.mustRun(t, "rebuild")
			}

			tests := []struct {
				name        string
				args        []string
				wantInside  []string
				wantOutside []string
			}{
				{
					name:       "full-text search",
					args:       []string{"search", "Golang"},
					wantInside: []string{"golang-basics"},
				},
				{
					name:        "tag filter",
					args:        []string{"search", "tag:go"},
					wantInside:  []string{"golang-basics", "cli-tools"},
					wantOutside: []string{"python-intro"},
				},
				{
					name:        "tag filter programming",
					args:        []string{"search", "tag:programming"},
					wantInside:  []string{"golang-basics", "python-intro"},
					wantOutside: []string{"cli-tools"},
				},
				{
					name:       "path filter",
					args:       []string{"search", "path:golang"},
					wantInside: []string{"golang-basics"},
				},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					out := e.mustRun(t, tt.args...)
					for _, want := range tt.wantInside {
						assertContains(t, out, want)
					}
					for _, notwant := range tt.wantOutside {
						assertNotContains(t, out, notwant)
					}
				})
			}
		})
	}
}
