# GLOW - Go LLM-Oriented Wiki

## Overview
Simple CLI tool providing wiki-like access to markdown articles with full-text search and metadata management.

## Features
- Create/update/delete/move articles
- Append to existing articles
- Full-text search using Bleve (content + metadata)
- Flexible metadata management (YAML frontmatter)
- Multi-wiki support with namespaces
- Single binary compilation

## Project Structure
```
glow/
├── cmd/
│   ├── root.go          # Cobra root, wiki name flag
│   ├── create.go        # Create article
│   ├── update.go        # Update article content + metadata
│   ├── append.go        # Append to article content
│   ├── delete.go        # Delete article
│   ├── move.go          # Move/rename article
│   ├── search.go        # Search articles + metadata
│   ├── list.go          # List articles
│   └── meta.go          # Metadata operations (set/add/delete)
├── internal/
│   ├── article/
│   │   ├── article.go   # Article struct, frontmatter parsing
│   │   └── metadata.go  # Metadata operations helpers
│   ├── storage/
│   │   ├── storage.go   # File operations, XDG paths
│   │   └── wiki.go      # Wiki struct, article CRUD
│   ├── index/
│   │   ├── index.go     # Bleve indexing (content + metadata)
│   │   └── search.go    # Search operations
│   └── config/
│       └── config.go    # Path resolution, wiki name resolution
├── main.go
├── go.mod
└── go.sum
```

## Data Layout
```
~/.local/share/glow/           # XDG_DATA_HOME/glow or $WIKI_DATA
├── wiki/
│   ├── default/
│   │   ├── articles/
│   │   │   ├── article1.md
│   │   │   ├── article2.md
│   │   │   └── projects/
│   │   │       ├── glow.md
│   │   │       └── team/
│   │   │           └── notes.md
│   │   └── index.bleve/           # Bleve index
│   └── work/
│       ├── articles/
│       │   └── meetings/
│       │       └── 2026-05.md
│       └── index.bleve/
```

## Article Format
```markdown
---
title: "Article Title"
tags: [go, cli, wiki]
aliases: [alt-name, another-name]
projects: [glow, other-project]
author: "Pavel"
created: 2026-05-19T10:30:00Z
modified: 2026-05-19T10:30:00Z
custom-field: "custom value"
---

# Article content here

Markdown content with full formatting support.
```

## CLI Commands

### Basic Operations
```bash
# Create article (optionally with initial metadata)
wiki create "article-name"
wiki create "folder/article-name"
wiki create "project/team/notes" --meta tags:go,cli --meta project:glow

# Update article (opens editor or accepts content)
wiki update "article-name"

# Update specific section by heading
wiki update "article-name" --section="Installation"
wiki update "article-name" --section="## API Reference"

# Append content to article
wiki append "article-name" "Additional content here"

# Append to specific section
wiki append "article-name" --section="Examples" "New example here"

# Delete article
wiki delete "article-name"

# Move/rename article
wiki move "old-name" "new-name"

# List articles
wiki list
```

### Search
```bash
# Search with embedded filters
wiki search "LLM tag:go tag:cli"
wiki search "project:glow indexing"
wiki search "alias:alt-name"
wiki search "author:Pavel custom-field:value"
wiki search "path:team/ meeting notes"    # Search in folder

# Search with explicit filter flags
wiki search "query" --filter=tag:go --filter=project:glow
```

### Metadata Operations
```bash
# Set field (scalar, overwrites)
wiki meta set article-name key value
wiki meta set my-article author "Pavel"

# Add to array field (appends, creates array if needed)
wiki meta add article-name tags go
wiki meta add article-name tags cli python
wiki meta add article-name projects glow

# Delete from field
wiki meta delete article-name tags go        # Remove from array
wiki meta delete article-name author         # Delete entire field
```

### Multi-Wiki Support
```bash
# Use specific wiki (default is "default")
wiki work create "article-name"
wiki work search "query"
wiki personal list

# If no wiki name specified, uses "default"
wiki create "article-name"  # Creates in default wiki
```

## Technical Implementation

### Storage
- **Path resolution**: XDG_DATA_HOME/glow or `$WIKI_DATA` environment variable
- **Wiki path**: `<base>/wiki/<wiki-name>/`
- **Articles**: `<wiki-path>/articles/**/*.md` (supports nested folders)
- **Article paths**: Can include folders: `folder/subfolder/article`
- **Index**: `<wiki-path>/index.bleve/`
- **Auto-create**: Parent directories created automatically

### Article Structure
- **Struct**: `Article` with `Metadata map[string]interface{}` and `Content string`
- **Parsing**: YAML frontmatter extraction/serialization
- **Auto-timestamps**: `created`, `modified` managed automatically
- **Section parsing**: Split markdown by headings for targeted updates
  - `ParseSections()` → map heading to content range
  - `UpdateSection(heading, newContent)` → replace section, preserve rest
  - `AppendToSection(heading, content)` → append within section
  - Section matching: case-insensitive, ignore `#` count
  - Section not found → error or create at end

### Metadata
- **Flexible schema**: Any key-value pairs supported
- **Common fields**: tags, aliases, projects (arrays), title, author (scalars)
- **Operations**:
  - `set`: Overwrites with scalar value
  - `add`: Appends to array (creates array if doesn't exist)
  - `delete`: Removes from array or deletes field entirely

### Search/Indexing
- **Engine**: Bleve full-text search
- **Indexed fields**: Article content + all metadata fields + path
- **Article ID**: Full relative path (e.g., `project/notes/meeting`)
- **Query parser**: Extract `field:value` patterns from query string
- **Path filtering**: `path:folder/` finds all articles in folder (prefix match)
- **Boolean queries**: Combine term queries (metadata) + match queries (content)
- **Auto-update**: Index updated on every article create/update/delete/move

### Dependencies
- **CLI**: github.com/spf13/cobra
- **Search**: github.com/blevesearch/bleve/v2
- **YAML**: gopkg.in/yaml.v3
- **XDG**: github.com/adrg/xdg (or manual XDG implementation)

## Implementation Notes
1. Each command in separate file under `cmd/`
2. Cobra command registration in respective files
3. Internal packages handle business logic
4. Index update triggered automatically on article mutations
5. Metadata operations preserve types (strings vs arrays)
6. Search query parser handles dynamic field filtering
7. Wiki name resolved from command prefix or defaults to "default"
8. Section-targeted updates: parse MD by headings, replace/append to specific sections
9. Section matching is case-insensitive and ignores heading level markdown (`#`, `##`, etc.)
10. Article paths support nested folders: `folder/subfolder/article`
11. Parent directories auto-created with `os.MkdirAll`
12. Article ID in index = full relative path from articles directory
13. List command can show tree structure or flat list with paths


## Implementation Status

✅ **COMPLETED** - All core features implemented and tested!
