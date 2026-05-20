---
name: knowledge
description: Personal knowledge base using GLOW wiki. Store and retrieve learnings, project context, CBA work, engineering practices, preferences. Use to remember decisions, patterns, configurations across sessions.
---

# Knowledge Skill

GLOW wiki binary: `wiki` (installed via go install)
Wiki name: `default`

## Structure

- `cba/eventhub/` - Team work, configs, decisions
- `projects/` - Project tracking  
- `engineering/` - Engineering patterns
- `preferences/` - User workflows

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

# Only use built-in read tool if wiki read fails
read /Users/pavel.pivovarov/Library/Application\ Support/wiki/wiki/default/articles/cba/eventhub/team-context.md
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
wiki append "article-name" "Additional content"

# Append to specific section (under heading)
wiki append "article-name" --section "Examples" "New example"
```

### Metadata Operations

```bash
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
```

## Behavior Guidelines

**On every user request**: Check if related to projects/team/CBA/engineering/preferences/past work
→ If yes: Search knowledge base FIRST before answering

**After learning/solving/being corrected**:
→ Update knowledge base immediately
- Search first to find existing articles
- Read existing content to avoid conflicts
- Use append for additions, or read + edit for updates
- Add metadata (tags, projects) for better search
- Use wikilinks `[[folder/article]]` in content

## Common Workflows

### Finding Information

```bash
# Search for EventHub context
wiki search "path:cba/eventhub/"

# Search by topic
wiki search "kafka MSK architecture"

# Search by tag
wiki search "tag:terraform"
```

### Updating Knowledge

```bash
# 1. Search for existing article
wiki search "logstash"

# 2. List sections to find what to update
wiki read "cba/eventhub/logstash-decommission" --sections

# 3. Read specific section
wiki read "cba/eventhub/logstash-decommission" --section "Current State"

# 4. Append new information (to end or specific section)
wiki append "cba/eventhub/logstash-decommission" "## Update 2026-05-19

New development: ..."

# OR append to specific section
wiki append "cba/eventhub/logstash-decommission" --section "Build Blockers" "- New blocker: ..."

# OR update entire section if replacing content
wiki update "cba/eventhub/logstash-decommission" --section "Current State" --content "Updated state info"

# OR edit directly if major changes needed
edit /Users/pavel.pivovarov/Library/Application\ Support/wiki/wiki/default/articles/cba/eventhub/logstash-decommission.md
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

# For complex content, use write tool directly
write /Users/pavel.pivovarov/Library/Application\ Support/wiki/wiki/default/articles/cba/eventhub/new-topic.md "---
tags: [kafka, eventhub]
project: eventhub
created: 2026-05-19T10:00:00Z
---

# New Topic

Content here..."

# Rebuild index after direct writes
wiki wiki-rebuild
```

## EventHub Context

**Role**: Principal Engineer, EventHub team (Kafka/MSK platform at CBA)

**Key articles** (read before EventHub work):
- `cba/eventhub/team-context` - Current state, priorities
- `cba/eventhub/platform-overview` - Technical overview
- `cba/eventhub/quick-start` - Commands, troubleshooting

**Platform**: 7 environments (Sandbox→Prod), ~100s tenants, AWS MSK + legacy on-prem Confluent

## Direct File Access

Articles stored at: `/Users/pavel.pivovarov/Library/Application Support/wiki/wiki/default/articles/`

Use built-in `read`, `write`, `edit` tools for direct file manipulation when:
- Batch operations needed
- Complex edits required
- Editor-based commands not suitable for agent

## Taskwarrior Integration

Pavel uses `task` command for todos. Add tasks when discovering action items:

```bash
task add project:eventhub priority:H "Description"
```

Projects: `eventhub`, `kafka-gateway`  
Priority: H (this week), M (soon), L (backlog)

**Don't duplicate in wiki.** Tasks = actions, wiki = knowledge.

## Index Management

```bash
# Verify index health
wiki wiki-verify

# Rebuild if corrupted
wiki wiki-rebuild
```

---

**Work silently.** Only report wiki activities if asked or critical context found.
