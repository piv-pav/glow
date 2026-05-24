---
name: knowledge
description: Personal knowledge base using GLOW wiki. Store and retrieve learnings, project context, CBA work, engineering practices, preferences. Use to remember decisions, patterns, configurations across sessions.
---

# Knowledge Skill

GLOW wiki binary: `wiki` (installed via go install)
Wiki name: `default`

## Structure

- `cba/{team_name}/` - Team work, configs, decisions
- `projects/{project_name}` - Project tracking  
- `engineering/` - Engineering patterns
- `preferences/` - User workflows

- You can extend and change structure if necessary.

## Operations

### Search

```bash
# Search content and metadata
wiki search "search term"

# Search with filters
wiki search "kafka tag:eventhub"
wiki search "path:cba/eventhub/ architecture"
wiki search "project:eventhub terraform"

# More results
wiki search "topic" -l 20
```

### Read Articles

```bash
# List all articles
wiki list

# Read specific article - ALWAYS use wiki read command
wiki read "cba/eventhub/team-context"

# List sections in article
wiki read "cba/eventhub/team-context" --sections

# Read specific section only
wiki read "cba/eventhub/team-context" --section "Current State"

# Include frontmatter
wiki read "article-name" --raw
```

### Write/Update Articles

```bash
# Create new article (LLM-friendly, no editor)
wiki create "cba/eventhub/new-topic" --content "# Title\n\nContent here" --meta "tags:kafka" --meta "project:eventhub"

# Or from stdin
echo "# Title\n\nContent" | wiki create "article-name" --stdin --meta "tags:value"

# Update entire article (LLM-friendly, no editor)
wiki update "article-name" --content "New content"

# Update specific section only
wiki update "article-name" --section "Phase 2" --content "Updated section content"

# Or from stdin
echo "Updated content" | wiki update "article-name" --stdin
echo "Section content" | wiki update "article-name" --section "Phase 2" --stdin

# Append to article (end of file)
echo "Additional content" | wiki append "article-name" --stdin
wiki append "article-name" --content "Additional content"

# Append to specific section (under heading)
echo "New example" | wiki append "article-name" --section "Examples" --stdin
wiki append "article-name" --section "Examples" --content "New example"
```

### Metadata Operations

```bash
# Get metadata value
wiki meta get "article-name" tags
wiki meta get "article-name" status

# Add tags
wiki meta add "article-name" tags kafka eventhub

# Set metadata
wiki meta set "article-name" author "Pavel"
wiki meta set "article-name" status "active"

# Remove metadata
wiki meta delete "article-name" tags kafka
wiki meta delete "article-name" status
```

### Article Management

```bash
# Move/rename
wiki move "old-name" "new-name"
wiki move "article" "folder/article"

# Delete
wiki delete "article-name"
wiki delete "article-name" --section "Section Heading"
```

## Mandatory Rules

**Tagging**: Every new article MUST have `tags` metadata. Use `--meta "tags:..."` on create, or `wiki meta add` after.
**Cross-linking**: Reference related articles with `[[folder/article]]` wikilinks in content.
**Search first**: Always `wiki search` before writing to avoid duplicates.

## CLI Tips

- **`--stdin`**: Use `--stdin` flag to pipe content into `wiki create`, `wiki update`, or `wiki append`. Example: `echo "content" | wiki append "name" --stdin`
- **`--content`**: Use `--content` flag for inline content on `wiki create`, `wiki update`, or `wiki append`. Example: `wiki append "name" --content "text"`
- **`wiki meta get`**: Get a metadata field value. Example: `wiki meta get "name" tags`
- **`--` separator**: Use `--` before article names that could be parsed as flags (e.g., starting with `-`).

## Behavior Guidelines

**On every user request**: Check if related to projects/team/CBA/engineering/preferences/past work
→ If yes: Search knowledge base FIRST before answering

**After learning/solving/being corrected**:
→ Update knowledge base immediately
- Search first to find existing articles
- Read existing content to avoid conflicts
- Use append for additions, or read + edit for updates
- Add metadata (tags, projects) per mandatory rules above
- Use wikilinks `[[folder/article]]` in content per mandatory rules above

## Common Workflows

### Updating Knowledge

```bash
# 1. Search for existing article
wiki search "logstash"

# 2. List sections to find what to update
wiki read "cba/eventhub/logstash-decommission" --sections

# 3. Read specific section
wiki read "cba/eventhub/logstash-decommission" --section "Current State"

# 4. Append new information (to end or specific section)
echo "## Update 2026-05-19

New development: ..." | wiki append "cba/eventhub/logstash-decommission" --stdin

# OR append to specific section
echo "- New blocker: ..." | wiki append "cba/eventhub/logstash-decommission" --section "Build Blockers" --stdin

# OR update entire section if replacing content
wiki update "cba/eventhub/logstash-decommission" --section "Current State" --content "Updated state info"
```

### Creating New Articles

```bash
# Create article with content (no editor)
wiki create "cba/eventhub/new-topic" \
  --content "# New Topic\n\nContent here..." \
  --meta "tags:kafka" \
  --meta "project:eventhub"

# Or use stdin
echo "# Title\n\nContent" | wiki create "article-name" --stdin --meta "tags:value"
```

## Index Management

```bash
# Verify index health
wiki wiki-verify

# Rebuild if corrupted
wiki wiki-rebuild
```

---

**Work silently.** Only report wiki activities if asked or critical context found.
