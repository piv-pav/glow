# GLOW - Go LLM-Oriented Wiki

A simple CLI tool providing wiki-like access to markdown articles with full-text search and metadata management.

## Features

- 📝 **Markdown Articles** - Store articles with YAML frontmatter metadata
- 🔍 **Full-Text Search** - Powered by Bleve search engine
- 🏷️ **Flexible Metadata** - Tags, projects, custom fields, aliases
- 📁 **Nested Folders** - Organize articles in hierarchical structure
- ✂️ **Section Editing** - Update specific sections of articles
- 📚 **Multi-Wiki** - Manage multiple independent wikis
- 🔧 **Index Management** - Verify and rebuild search index

## Installation

```bash
go install github.com/pavelpivovarov/glow/cmd/wiki@latest
```

Or build from source:

```bash
git clone https://github.com/pavelpivovarov/glow
cd glow
just install

# Or build locally
just build
./wiki --version
```

## Quick Start

```bash
# Create an article
wiki create "my-first-article"

# Add metadata
wiki meta add my-first-article tags go cli
wiki meta set my-first-article author "Your Name"

# Search articles
wiki search "search term tag:go"

# List all articles
wiki list

# Create a new wiki
wiki wiki-create work

# Use different wiki
wiki -w work create "work-notes"
```

## Usage

### Article Operations

```bash
# Create article
wiki create "article-name"
wiki create "folder/article-name" --meta tags:go,cli --meta project:glow

# Update article (opens editor)
wiki update "article-name"

# Update specific section
wiki update "article-name" --section="Installation"

# Append content
wiki append "article-name" "Additional content here"

# Append to section
wiki append "article-name" --section="Examples" "New example"

# Delete article
wiki delete "article-name"

# Move/rename article
wiki move "old-name" "new-name"
wiki move "article" "folder/article"
```

### Metadata Management

```bash
# Set scalar field
wiki meta set article-name author "Pavel"
wiki meta set article-name status "draft"

# Add to array field
wiki meta add article-name tags go cli
wiki meta add article-name projects glow

# Delete field
wiki meta delete article-name status              # Delete entire field
wiki meta delete article-name tags go             # Remove from array
```

### Search

```bash
# Simple search
wiki search "golang"

# Search with filters
wiki search "indexing tag:go tag:cli"
wiki search "project:glow documentation"
wiki search "path:team/ meeting notes"

# Using explicit filters
wiki search "query" --filter=tag:go --filter=project:glow -l 20
```

### Wiki Management

```bash
# List all wikis
wiki wiki-list

# Create new wiki
wiki wiki-create work

# Use specific wiki
wiki -w work list
wiki -w personal create "notes"

# Verify index health
wiki wiki-verify

# Rebuild index (if corrupted)
wiki wiki-rebuild
```

### Listing

```bash
# List all articles in current wiki
wiki list

# List articles in specific wiki
wiki -w work list
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

Override with `WIKI_DATA` environment variable:

```bash
export WIKI_DATA=/path/to/your/wikis
wiki list
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
wiki search "tag:go tag:cli"

# Find project documentation
wiki search "project:glow architecture"

# Find in specific folder
wiki search "path:team/ retrospective"

# Multiple criteria
wiki search "kubernetes author:Pavel tag:devops"
```

## Development

### Project Structure

```
glow/
├── cmd/               # Cobra commands
│   ├── root.go
│   ├── create.go
│   ├── update.go
│   ├── append.go
│   ├── delete.go
│   ├── move.go
│   ├── search.go
│   ├── list.go
│   ├── meta.go
│   └── wiki.go
├── internal/
│   ├── article/       # Article parsing, metadata
│   ├── storage/       # File operations
│   ├── index/         # Bleve indexing
│   └── config/        # Path management
└── main.go
```

### Dependencies

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Bleve](https://github.com/blevesearch/bleve) - Full-text search
- [go-yaml](https://gopkg.in/yaml.v3) - YAML parsing
- [xdg](https://github.com/adrg/xdg) - XDG directories

### Building

```bash
# Build
just build

# Run tests
just test

# Install locally
just install
```

## Contributing

Contributions welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details.

## Roadmap

- [ ] Export to different formats (PDF, HTML)
- [ ] Git integration for version control
- [ ] Web interface
- [ ] Article templates
- [ ] Link validation
- [ ] Backlinks tracking
- [ ] Graph visualization
- [ ] Import from other wikis/note systems

## Author

Pavel Pivovarov - [GitHub](https://github.com/pavelpivovarov)
