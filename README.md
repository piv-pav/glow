# GLOW - Go LLM-Oriented Wiki

A simple CLI tool providing wiki-like access to markdown articles with full-text search and metadata management.

> ## ⚠️ Breaking Changes in 0.9.0
>
> Version **0.9.0** massively changes the way glow operates and is **not backward compatible** with 0.8.7 and below.
>
> See the [CHANGELOG](CHANGELOG.md) for the full nature of the changes.
>
> **If you are upgrading from 0.8.7 or earlier, you must migrate your data to avoid data loss:**
>
> 1. **Export** all your wikis *before* upgrading:
>    ```bash
>    glow export <wiki> /tmp/<wiki>-backup.tar.gz
>    ```
> 2. **Upgrade** glow to 0.9.0.
> 3. **Import** your wikis back:
>    ```bash
>    glow import <wiki> /tmp/<wiki>-backup.tar.gz
>    ```

## Features

- 📝 **Markdown Articles** - Store articles with YAML frontmatter metadata
- 🔍 **Full-Text Search** - FTS5 with BM25 ranking
- 🏷️ **Tagging** - Add/remove tags for organization and search
- 📁 **Nested Folders** - Organize articles in hierarchical structure
- ✂️ **Section Editing** - Update specific sections of articles
- 📚 **Multi-Wiki** - Manage multiple independent wikis
- 💾 **SQLite + rqlite** - Local file or distributed cluster
- 📦 **Export/Import** - Migrate articles between wikis

## Installation

```bash
go install codeberg.org/pivpav/glow@latest
```

Or build from source:

```bash
git clone https://codeberg.org/pivpav/glow
cd glow
just install  # Runs tests then installs

# Or build locally
just build    # Runs tests then builds
./glow --version
```

## Quick Start

```bash
# Initialize a wiki (defaults to "default" with sqlite backend)
glow init

# Or with a specific name
glow init work

# Create an article
glow create "my-first-article" --content "# Hello

My first wiki article." --tag go --tag wiki

# Search articles
glow search "search term tag:go"

# List all articles
glow list

# Use different wiki
glow -w work create "work-notes" --content "Notes"
```

## Usage

### Article Operations

```bash
# Create article
glow create "article-name" --content "Content"
glow create "folder/article-name" --content "Content" --tag go --tag glow

# Create with multiline content (preferred)
glow create "article-name" --content "# Title

Content here"
echo "# Title

Content" | glow create "article-name" --stdin

# \n escapes also work in --content
glow create "article-name" --content "# Title\n\nContent here"

# Read article
glow read "article-name"                    # Content only
glow read "article-name" --tags             # List only the article's tags (-t)
glow read "article-name" --section "Setup"  # Read specific section
glow read "article-name" --sections         # List all sections

# Update article
glow update "article-name" --content "New content"
glow update "article-name" --content "# Title

Multiline content"
echo "New content" | glow update "article-name" --stdin

# Update tags
glow update "article-name" --tag newtag
glow update "article-name" --untag oldtag
glow update "article-name" --tag a,b --untag c

# Update specific section
glow update "article-name" --section "Installation" --content "New section content"

# Update via SEARCH/REPLACE diff blocks from STDIN (AI-style edits)
printf '<<<<<<< SEARCH\nold text\n=======\nnew text\n>>>>>>> REPLACE\n' | glow update "article-name" --diff
# Scope the diff to one section
printf '<<<<<<< SEARCH\nold\n=======\nnew\n>>>>>>> REPLACE\n' | glow update "article-name" --diff --section "Status"

# Append content
glow append "article-name" --content "Additional content"
glow append "article-name" --content "Line 1

Line 2"
echo "Additional content" | glow append "article-name" --stdin

# Append to section
glow append "article-name" --section "Examples" --content "New example"

# Delete article
glow delete "article-name"

# Delete specific section
glow delete "article-name" --section "Section Heading"

# Move/rename article
glow move "old-name" "new-name"
glow move "article" "folder/article"
```

### Tag Management

```bash
# Add tags on create
glow create "article" --content "Content" --tag go --tag cli

# Add tags on update
glow update "article" --tag newtag
glow update "article" --tag a,b  # comma-separated

# Remove tags
glow update "article" --untag oldtag
glow update "article" --untag a,b  # comma-separated

# Combine add and remove
glow update "article" --tag new --untag old
```

### Search

```bash
# Simple search
glow search "golang"

# Search with filters
glow search "indexing tag:go tag:cli"
glow search "path:team/ meeting notes"

# Using explicit filters
glow search "query" --filter=tag:go -l 20
```

### Wiki Management

```bash
# Initialize wiki (interactive backend selection)
glow init              # creates "default" wiki
glow init work         # creates named wiki

# Quick create (sqlite, non-interactive)
glow wiki-create work

# Delete a wiki
glow wiki-delete work

# List all wikis
glow wiki-list

# Use specific wiki
glow -w work list
glow -w personal create "notes" --content "Notes"

# Export/Import
glow export default /tmp/backup.tar.gz
glow import work /tmp/backup.tar.gz
```

### Listing

```bash
# List all articles in current wiki
glow list

# List articles in specific wiki
glow -w work list
```

## Article Format

Articles are Markdown files with minimal YAML frontmatter (managed automatically):

```markdown
---
tags:
  - go
  - cli
created: 2026-05-19T10:30:00Z
modified: 2026-05-19T10:30:00Z
---

# Article Content

Your markdown content here...
```

## Data Storage

Glow supports two storage backends:

| Backend | Search | Best for |
|---------|--------|----------|
| **SQLite** (default) | FTS5 | Single-machine, fast, zero config |
| **rqlite** | FTS5 | Distributed, offline reads, cluster |

Select backend during `glow init` or set in `~/.config/glow/glow.yaml`:

```yaml
wikis:
  default:
    backend: sqlite
  distributed:
    backend: rqlite
    rqlite:
      url: "http://localhost:4001"
      user: glow
      password: secret
      level: weak          # none|weak|strong
```

By default, data is stored in XDG-compliant directories:

- **macOS**: `~/Library/Application Support/glow/wiki/`
- **Linux**: `~/.local/share/glow/wiki/`
- **Windows**: `%LOCALAPPDATA%\glow\wiki\`

Override with `GLOW_DATA` environment variable:

```bash
export GLOW_DATA=/path/to/your/wikis
glow list
```

### Directory Structure

```
~/Library/Application Support/glow/wiki/
├── default.db
├── notes.db
└── work.db
```

## Search Syntax

Filters can be embedded directly in the query:

- `tag:value` - Search by tag
- `path:folder/` - Search in specific folder (prefix match)

```bash
# Find Go CLI articles
glow search "tag:go tag:cli"

# Find in specific folder
glow search "path:team/ retrospective"

# Combine with text
glow search "kubernetes tag:devops"
```

## Development

### Project Structure

```
glow/
├── main.go
├── tools/             # Cobra command implementations
│   ├── register.go
│   ├── append.go
│   ├── create.go
│   ├── delete.go
│   ├── exportimport.go
│   ├── helpers.go
│   ├── list.go
│   ├── move.go
│   ├── read.go
│   ├── search.go
│   ├── update.go
│   └── wiki.go
├── internal/
│   ├── article/       # Article parsing, tags
│   ├── storage/       # Backend implementations (sqlite, rqlite)
│   └── config/        # Configuration, wiki registry
└── tests/             # Integration tests
```

### Dependencies

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [modernc.org/sqlite](https://modernc.org/sqlite) - Pure-Go SQLite driver (no CGO)
- [gorqlite](https://github.com/rqlite/gorqlite) - rqlite driver
- [go-yaml](https://gopkg.in/yaml.v3) - YAML parsing
- [xdg](https://github.com/adrg/xdg) - XDG directories

### Building

**Tests run automatically** before build/install:

```bash
# Build (runs tests first)
just build

# Install (runs tests first)
just install

# Run tests only
just test

# Format code
just fmt
```

Tests use isolated environment (`GLOW_DATA=/tmp/glow-test-wiki`) and cover:

- Create/append/update/delete operations
- Section-targeted edits
- Search with filters
- Metadata operations
- Move/rename articles

## Contributing

Contributions welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details.

Pavel Pivovarov - [Codeberg](https://codeberg.org/pivpav/glow)
