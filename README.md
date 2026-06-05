# GLOW - Go LLM-Oriented Wiki

A simple CLI tool providing wiki-like access to markdown articles with full-text search and metadata management.

## Features

- 📝 **Markdown Articles** - Store articles with YAML frontmatter metadata
- 🔍 **Full-Text Search** - Powered by Bleve search engine
- 🏷️ **Flexible Metadata** - Tags, projects, custom fields, aliases
- 📁 **Nested Folders** - Organize articles in hierarchical structure
- ✂️ **Section Editing** - Update specific sections of articles
- 📚 **Multi-Wiki** - Manage multiple independent wikis
- 🔧 **Index Management** - Rebuild search index if corrupted

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
# Create an article
glow create "my-first-article" --content "# Hello

My first wiki article."

# Add metadata
glow meta add my-first-article tags go cli
glow meta set my-first-article author "Your Name"

# Search articles
glow search "search term tag:go"

# List all articles
glow list

# Create a new wiki
glow wiki-create work

# Use different wiki
glow -w work create "work-notes" --content "Notes"
```

## Usage

### Article Operations

```bash
# Create article
glow create "article-name" --content "Content"
glow create "folder/article-name" --content "Content" --meta "tags:go" --meta "project:glow"

# Create with multiline content (preferred)
glow create "article-name" --content "# Title

Content here"
echo "# Title

Content" | glow create "article-name" --stdin

# \n escapes also work in --content
glow create "article-name" --content "# Title\n\nContent here"

# Read article
glow read "article-name"                    # Content only
glow read "article-name" --raw              # Include frontmatter
glow read "article-name" --section "Setup"  # Read specific section
glow read "article-name" --sections         # List all sections

# Update article
glow update "article-name" --content "New content"
glow update "article-name" --content "# Title

Multiline content"
echo "New content" | glow update "article-name" --stdin

# Update specific section
glow update "article-name" --section "Installation" --content "New section content"

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

### Metadata Management

```bash
# Set scalar field
glow meta set article-name author "Pavel"
glow meta set article-name status "draft"

# Add to array field
glow meta add article-name tags go cli
glow meta add article-name projects glow

# Delete field
glow meta delete article-name status              # Delete entire field
glow meta delete article-name tags go             # Remove from array
```

### Search

```bash
# Simple search
glow search "golang"

# Search with filters
glow search "indexing tag:go tag:cli"
glow search "project:glow documentation"
glow search "path:team/ meeting notes"

# Using explicit filters
glow search "query" --filter=tag:go --filter=project:glow -l 20
```

### Wiki Management

```bash
# List all wikis
glow wiki-list

# Create new wiki
glow wiki-create work

# Use specific wiki
glow -w work list
glow -w personal create "notes" --content "Notes"

# Rebuild index (if corrupted)
glow rebuild
```

### Listing

```bash
# List all articles in current wiki
glow list

# List articles in specific wiki
glow -w work list
```

## Article Format

Articles are stored as Markdown files with YAML frontmatter:

```markdown
---
title: "Article Title"
tags: [go, cli, wiki]
aliases: [alt-name, another-name]
projects: [glow]
author: "Pavel"
created: 2026-05-19T10:30:00Z
modified: 2026-05-19T10:30:00Z
path: folder/article
custom-field: "custom value"
---

# Article Content

Your markdown content here...

## Section 1

Content for section 1.

## Section 2

Content for section 2.
```

## Data Storage

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
├── default/
│   ├── articles/
│   │   ├── article1.md
│   │   ├── article2.md
│   │   └── projects/
│   │       └── glow.md
│   └── index.bleve/
└── work/
    ├── articles/
    │   └── meetings/
    │       └── 2026-05.md
    └── index.bleve/
```

## Search Syntax

### Embedded Filters

Filters can be embedded directly in the query:

- `tag:value` - Search by tag
- `project:value` - Search by project
- `author:value` - Search by author
- `path:folder/` - Search in specific folder (prefix match)
- Any custom metadata field: `fieldname:value`

### Examples

```bash
# Find Go CLI articles
glow search "tag:go tag:cli"

# Find project documentation
glow search "project:glow architecture"

# Find in specific folder
glow search "path:team/ retrospective"

# Multiple criteria
glow search "kubernetes author:Pavel tag:devops"
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
│   ├── helpers.go
│   ├── list.go
│   ├── meta.go
│   ├── move.go
│   ├── read.go
│   ├── search.go
│   ├── update.go
│   └── wiki.go
├── internal/
│   ├── article/       # Article parsing, metadata
│   ├── storage/       # File operations
│   ├── index/         # Bleve indexing
│   └── config/        # Path management
└── tests/             # Integration tests
```

### Dependencies

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Bleve](https://github.com/blevesearch/bleve) - Full-text search
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
