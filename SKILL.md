---
name: knowledge
description: Personal knowledge base using GLOW wiki. Store and retrieve learnings, project context, engineering practices, preferences. Use to remember decisions, patterns, configurations across sessions.
---

# Knowledge Skill

GLOW stands for Golang LLM-Oriented Wiki.

GLOW wiki binary: `glow` (installed via go install . or go install git.netra.pivpav.com/public/glow@latest)
Wiki name: `default`

## Structure

No fixed structure. AI decides organization based on what's effective for each topic.

Guidelines (not rules):
- **Projects** → `projects/{name}/` if it's about a specific project
- **Engineering patterns** → `engineering/` 
- **Machine/device info** → `machine/`
- But AI can create ANY path, folder, or naming scheme it finds effective
- Cross-link everything with `[[path/to/article]]` wikilinks
- Tags and project metadata should be added, but structure is freeform

## Operations

### Search

```bash
# Search content and metadata
glow search "search term"

# Search with filters
glow search "kafka tag:eventhub"
glow search "path:projects/eventhub/ architecture"
glow search "project:eventhub terraform"

# More results
glow search "topic" -l 20
```

### Read Articles

```bash
# List all articles
glow list

# Read specific article - ALWAYS use glow read command
glow read "projects/eventhub/team-context"

# List sections in article
glow read "projects/eventhub/team-context" --sections

# Read specific section only
glow read "projects/eventhub/team-context" --section "Current State"

# Include frontmatter
glow read "article-name" --raw
```

### Write/Update Articles

```bash
# Create new article (LLM-friendly, no editor)
glow create "projects/eventhub/new-topic" --content "# Title\n\nContent here" --meta "tags:kafka" --meta "project:eventhub"

# Or from stdin
echo "# Title\n\nContent" | glow create "article-name" --stdin --meta "tags:value"

# Update entire article (LLM-friendly, no editor)
glow update "article-name" --content "New content"

# Update specific section only
glow update "article-name" --section "Phase 2" --content "Updated section content"

# Or from stdin
echo "Updated content" | glow update "article-name" --stdin
echo "Section content" | glow update "article-name" --section "Phase 2" --stdin

# Append to article (end of file)
echo "Additional content" | glow append "article-name" --stdin
glow append "article-name" --content "Additional content"

# Append to specific section (under heading)
echo "New example" | glow append "article-name" --section "Examples" --stdin
glow append "article-name" --section "Examples" --content "New example"
```

### Metadata Operations

```bash
# Get metadata value
glow meta get "article-name" tags
glow meta get "article-name" status

# Add tags
glow meta add "article-name" tags kafka eventhub

# Set metadata
glow meta set "article-name" author "Pavel"
glow meta set "article-name" status "active"

# Remove metadata
glow meta delete "article-name" tags kafka
glow meta delete "article-name" status
```

### Article Management

```bash
# Move/rename
glow move "old-name" "new-name"
glow move "article" "folder/article"

# Delete
glow delete "article-name"
glow delete "article-name" --section "Section Heading"
```

## Mandatory Rules

**Tagging**: Every new article MUST have `tags` metadata. Use `--meta "tags:..."` on create, or `glow meta add` after.
**Cross-linking**: Reference related articles with `[[folder/article]]` wikilinks in content.
**Search first**: Always `glow search` before writing to avoid duplicates.

## CLI Tips

- **`--stdin`**: Use `--stdin` flag to pipe content into `glow create`, `glow update`, or `glow append`. Example: `echo "content" | glow append "name" --stdin`
- **`--content`**: Use `--content` flag for inline content on `glow create`, `glow update`, or `glow append`. Example: `glow append "name" --content "text"`
- **`glow meta get`**: Get a metadata field value. Example: `glow meta get "name" tags`
- **`--` separator**: Use `--` before article names that could be parsed as flags (e.g., starting with `-`).

## Behavior Guidelines

**On every user request**: Check if related to projects/team/engineering/preferences/past work
→ If yes: Search knowledge base FIRST before answering

**After learning/solving/being corrected**:
→ Update knowledge base immediately
- Search first to find existing articles
- Read existing content to avoid conflicts
- Use append for additions, or read + edit for updates
- Add metadata (tags, projects) per mandatory rules above
- Use glowlinks `[[folder/article]]` in content per mandatory rules above

## Common Workflows

### Updating Knowledge

```bash
# 1. Search for existing article
glow search "logstash"

# 2. List sections to find what to update
glow read "projects/eventhub/logstash-decommission" --sections

# 3. Read specific section
glow read "projects/eventhub/logstash-decommission" --section "Current State"

# 4. Append new information (to end or specific section)
echo "## Update 2026-05-19

New development: ..." | glow append "projects/eventhub/logstash-decommission" --stdin

# OR append to specific section
echo "- New blocker: ..." | glow append "projects/eventhub/logstash-decommission" --section "Build Blockers" --stdin

# OR update entire section if replacing content
glow update "projects/eventhub/logstash-decommission" --section "Current State" --content "Updated state info"
```

### Creating New Articles

```bash
# Create article with content (no editor)
glow create "projects/eventhub/new-topic" \
  --content "# New Topic\n\nContent here..." \
  --meta "tags:kafka" \
  --meta "project:eventhub"

# Or use stdin
echo "# Title\n\nContent" | glow create "article-name" --stdin --meta "tags:value"
```

## Index Management

```bash
# Verify index health
glow glow-verify

# Rebuild if corrupted
glow glow-rebuild
```

---

**Work silently.** Only report glow activities if asked or critical context found.
