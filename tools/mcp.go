package tools

import (
	"context"
	"fmt"
	"strings"

	"codeberg.org/pivpav/glow/internal/article"
	"codeberg.org/pivpav/glow/internal/storage"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP server (stdio)",
	Long:  `Start a Model Context Protocol server over stdio, exposing all wiki operations as tools.`,
	Args:  cobra.NoArgs,
	RunE:  runMCP,
}

var mcpWiki string

func init() {
	mcpCmd.Flags().StringVarP(&mcpWiki, "wiki", "w", "default", "Wiki to expose via MCP")
}

func runMCP(cmd *cobra.Command, args []string) error {
	s := server.NewMCPServer("glow", cmd.Root().Version)

	wikiParam := mcp.WithString("wiki_name",
		mcp.Description(`Wiki name to use (default: "default"). Only specify if targeting a non-default wiki.`),
	)

	// helper: resolve wiki_name from request, fall back to --wiki flag
	wiki := func(req mcp.CallToolRequest) string {
		w := req.GetString("wiki_name", "")
		if w == "" {
			return mcpWiki
		}
		return w
	}

	// ── search ──────────────────────────────────────────────────────────────
	s.AddTool(mcp.NewTool("search",
		mcp.WithDescription("Search wiki articles by content and tags. Embed tag: and path: filters in query, e.g. \"kafka tag:eventhub\"."),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query, optionally with tag: and path: filters")),
		mcp.WithInteger("limit", mcp.Description("Max results (default 10)")),
		wikiParam,
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query := req.GetString("query", "")
		limit := req.GetInt("limit", 10)
		wikiName := wiki(req)

		query, filters := parseEmbeddedFilters(query)

		store, err := storage.New(wikiName)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		defer store.Close()

		searcher, ok := store.(storage.Searcher)
		if !ok {
			return mcp.NewToolResultError("backend does not support search"), nil
		}
		out, err := searcher.Search(query, filters, limit)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		var sb strings.Builder
		if len(out.Results) == 0 {
			sb.WriteString("No results found")
		} else {
			if out.Total > len(out.Results) {
				fmt.Fprintf(&sb, "Found %d results (showing top %d):\n\n", out.Total, len(out.Results))
			} else {
				fmt.Fprintf(&sb, "Found %d results:\n\n", out.Total)
			}
			for i, r := range out.Results {
				fmt.Fprintf(&sb, "%d. %s\n", i+1, r.Name)
				if r.Snippet != "" {
					fmt.Fprintf(&sb, "   %s\n", r.Snippet)
				}
				if len(r.Tags) > 0 {
					fmt.Fprintf(&sb, "   [tags: %s]\n", strings.Join(r.Tags, ", "))
				}
				sb.WriteByte('\n')
			}
		}
		return mcp.NewToolResultText(sb.String()), nil
	})

	// ── list ─────────────────────────────────────────────────────────────────
	s.AddTool(mcp.NewTool("list",
		mcp.WithDescription("List all articles in the wiki."),
		wikiParam,
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		wikiName := wiki(req)
		var result string
		err := withStore(wikiName, func(store storage.Store) error {
			articles, err := store.List()
			if err != nil {
				return err
			}
			if len(articles) == 0 {
				result = fmt.Sprintf("No articles in wiki '%s'", wikiName)
				return nil
			}
			var sb strings.Builder
			fmt.Fprintf(&sb, "Articles in wiki '%s' (%d):\n\n", wikiName, len(articles))
			for _, a := range articles {
				fmt.Fprintf(&sb, "  %s\n", a)
			}
			result = sb.String()
			return nil
		})
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(result), nil
	})

	// ── read ─────────────────────────────────────────────────────────────────
	s.AddTool(mcp.NewTool("read",
		mcp.WithDescription("Read an article's content. Optionally retrieve a specific section or list all sections/tags."),
		mcp.WithString("name", mcp.Required(), mcp.Description("Article name")),
		mcp.WithString("section", mcp.Description("Read only this section heading")),
		mcp.WithBoolean("sections", mcp.Description("List all section headings")),
		mcp.WithBoolean("tags", mcp.Description("List only the article's tags")),
		wikiParam,
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name := req.GetString("name", "")
		section := req.GetString("section", "")
		sections := req.GetBool("sections", false)
		tags := req.GetBool("tags", false)
		wikiName := wiki(req)

		var result string
		err := withStore(wikiName, func(store storage.Store) error {
			art, err := store.Read(name)
			if err != nil {
				return err
			}
			switch {
			case tags:
				result = strings.Join(art.GetTags(), "\n")
			case sections:
				var sb strings.Builder
				fmt.Fprintf(&sb, "Sections in %s:\n\n", name)
				for _, s := range art.ParseSections() {
					if s.Heading == "" {
						sb.WriteString("  (preamble)\n")
					} else {
						fmt.Fprintf(&sb, "  %s %s\n", strings.Repeat("#", s.Level), s.Heading)
					}
				}
				result = sb.String()
			case section != "":
				s := art.FindSection(section)
				if s == nil {
					return fmt.Errorf("section not found: %s", section)
				}
				lines := strings.Split(s.Content, "\n")
				if len(lines) > 1 {
					result = strings.Join(lines[1:], "\n")
				}
			default:
				result = art.Content
			}
			return nil
		})
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(result), nil
	})

	// ── create ───────────────────────────────────────────────────────────────
	s.AddTool(mcp.NewTool("create",
		mcp.WithDescription("Create a new article. Article name can include folders (e.g. folder/article)."),
		mcp.WithString("name", mcp.Required(), mcp.Description("Article name")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Article content (markdown)")),
		mcp.WithArray("tags", mcp.Description("Tags to apply"), mcp.WithStringItems()),
		wikiParam,
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name := req.GetString("name", "")
		content := req.GetString("content", "")
		tags := req.GetStringSlice("tags", nil)
		wikiName := wiki(req)

		art := article.New(content)
		if len(tags) > 0 {
			art.AddTags(tags...)
		}

		var result string
		err := withStore(wikiName, func(store storage.Store) error {
			if err := store.Create(name, art); err != nil {
				return err
			}
			result = fmt.Sprintf("Created article: %s", name)
			return nil
		})
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(result), nil
	})

	// ── update ───────────────────────────────────────────────────────────────
	s.AddTool(mcp.NewTool("update",
		mcp.WithDescription(`Update an existing article. Provide content to replace article body (or a specific section).
For surgical edits use diff_blocks: one or more SEARCH/REPLACE pairs in the format:
<<<<<<< SEARCH
exact existing text
=======
replacement text
>>>>>>> REPLACE`),
		mcp.WithString("name", mcp.Required(), mcp.Description("Article name")),
		mcp.WithString("content", mcp.Description("New content (replaces article or section)")),
		mcp.WithString("diff_blocks", mcp.Description("SEARCH/REPLACE diff blocks for surgical edits")),
		mcp.WithString("section", mcp.Description("Limit update to this section heading")),
		mcp.WithArray("tags", mcp.Description("Tags to add"), mcp.WithStringItems()),
		mcp.WithArray("untags", mcp.Description("Tags to remove"), mcp.WithStringItems()),
		wikiParam,
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name := req.GetString("name", "")
		content := req.GetString("content", "")
		diffBlocks := req.GetString("diff_blocks", "")
		section := req.GetString("section", "")
		tags := req.GetStringSlice("tags", nil)
		untags := req.GetStringSlice("untags", nil)
		wikiName := wiki(req)

		var result string
		appliedBlocks := 0

		err := modifyArticleQuiet(wikiName, name, func(art *article.Article) error {
			switch {
			case diffBlocks != "" && section != "":
				n, err := art.ApplyDiffToSection(section, diffBlocks)
				if err != nil {
					return err
				}
				appliedBlocks = n
			case diffBlocks != "":
				res, n, err := article.ApplyDiff(art.Content, diffBlocks)
				if err != nil {
					return err
				}
				art.Content = res
				appliedBlocks = n
			case content != "" && section != "":
				if err := art.UpdateSection(section, content); err != nil {
					return err
				}
			case content != "":
				art.Content = content
			}
			if len(tags) > 0 {
				art.AddTags(tags...)
			}
			if len(untags) > 0 {
				art.RemoveTags(untags...)
			}
			return nil
		})
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		switch {
		case diffBlocks != "" && section != "":
			result = fmt.Sprintf("Applied %d diff block(s) to section %q in article: %s", appliedBlocks, section, name)
		case diffBlocks != "":
			result = fmt.Sprintf("Applied %d diff block(s) to article: %s", appliedBlocks, name)
		case section != "":
			result = fmt.Sprintf("Updated section %q in article: %s", section, name)
		default:
			result = fmt.Sprintf("Updated article: %s", name)
		}
		return mcp.NewToolResultText(result), nil
	})

	// ── append ───────────────────────────────────────────────────────────────
	s.AddTool(mcp.NewTool("append",
		mcp.WithDescription("Append content to an article or a specific section."),
		mcp.WithString("name", mcp.Required(), mcp.Description("Article name")),
		mcp.WithString("content", mcp.Description("Content to append")),
		mcp.WithString("section", mcp.Description("Append to this section heading")),
		mcp.WithArray("tags", mcp.Description("Tags to add"), mcp.WithStringItems()),
		mcp.WithArray("untags", mcp.Description("Tags to remove"), mcp.WithStringItems()),
		wikiParam,
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name := req.GetString("name", "")
		content := req.GetString("content", "")
		section := req.GetString("section", "")
		tags := req.GetStringSlice("tags", nil)
		untags := req.GetStringSlice("untags", nil)
		wikiName := wiki(req)

		msg := fmt.Sprintf("Appended to article: %s", name)
		if section != "" {
			msg = fmt.Sprintf("Appended to section %q in article: %s", section, name)
		}

		err := modifyArticle(wikiName, name, func(art *article.Article) error {
			if content != "" {
				if section != "" {
					if err := art.AppendToSection(section, content); err != nil {
						return err
					}
				} else {
					if art.Content != "" && art.Content[len(art.Content)-1] != '\n' {
						art.Content += "\n"
					}
					art.Content += "\n" + content
				}
			}
			if len(tags) > 0 {
				art.AddTags(tags...)
			}
			if len(untags) > 0 {
				art.RemoveTags(untags...)
			}
			return nil
		}, msg)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(msg), nil
	})

	// ── delete ───────────────────────────────────────────────────────────────
	s.AddTool(mcp.NewTool("delete",
		mcp.WithDescription("Delete an article or a specific section within an article."),
		mcp.WithString("name", mcp.Required(), mcp.Description("Article name")),
		mcp.WithString("section", mcp.Description("Delete only this section heading")),
		wikiParam,
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name := req.GetString("name", "")
		section := req.GetString("section", "")
		wikiName := wiki(req)

		var result string
		if section != "" {
			msg := fmt.Sprintf("Deleted section: %s from article: %s", section, name)
			err := modifyArticle(wikiName, name, func(art *article.Article) error {
				return art.DeleteSection(section)
			}, msg)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			result = msg
		} else {
			err := withStore(wikiName, func(store storage.Store) error {
				return store.Delete(name)
			})
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			result = fmt.Sprintf("Deleted article: %s", name)
		}
		return mcp.NewToolResultText(result), nil
	})

	// ── move ─────────────────────────────────────────────────────────────────
	s.AddTool(mcp.NewTool("move",
		mcp.WithDescription("Move or rename an article."),
		mcp.WithString("old_name", mcp.Required(), mcp.Description("Current article name")),
		mcp.WithString("new_name", mcp.Required(), mcp.Description("New article name")),
		wikiParam,
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		oldName := req.GetString("old_name", "")
		newName := req.GetString("new_name", "")
		wikiName := wiki(req)

		err := withStore(wikiName, func(store storage.Store) error {
			return store.Move(oldName, newName)
		})
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Moved article: %s -> %s", oldName, newName)), nil
	})

	return server.ServeStdio(s)
}
