---
name: glow
description: GLOW wiki operations. Use for ALL glow commands, wiki searches, knowledge storage/retrieval. Essential for remembering context, project info, engineering patterns, configurations. Must use when user mentions wiki, memory, remember, or need persistent storage.
---

# GLOW Wiki Skill

**CRITICAL: Use this skill for ANY `glow` command or wiki operation**

## When to Use This Skill

**ALWAYS use this skill when**:
- User says: "remember", "note", "memory", "wiki", "save this", "store"
- Running ANY `glow` command (search, read, create, update, append, meta, etc.)
- Looking up project context, team info, engineering patterns
- User asks about past work, decisions, configurations
- Need to persist learnings across sessions

**MANDATORY: Start every conversation with glow search**:
1. Extract key terms from user's request (project names, technologies, concepts)
2. Search glow wiki: `glow search "extracted terms"`
3. Read relevant articles found
4. Use that context in your response
5. If no relevant context found, proceed normally

## Operations

### Search
```bash
glow search "search term"
glow search "kafka tag:eventhub"
glow search "path:projects/eventhub/ architecture"
```

### Read Articles
```bash
glow list
glow read "article-name"
glow read "article-name" --sections
glow read "article-name" --section "Section Name"
```

### Write/Update Articles

**Prefer multiline `--content` over `\n` escape sequences** — cleaner and less error-prone:
```bash
# Create
glow create "article-name" --content "# Title

First paragraph.

Second paragraph." --meta "tags:value"
echo "Content" | glow create "article-name" --stdin

# Update
glow update "article-name" --content "New content"
glow update "article-name" --section "Section" --content "Section content"
echo "New content" | glow update "article-name" --stdin

# Append
glow append "article-name" --content "More content"
glow append "article-name" --section "Section" --content "New item"
echo "More content" | glow append "article-name" --stdin
```

Note: `\n` in `--content` is also interpreted if needed, but multiline strings are preferred.

### Metadata
```bash
glow meta get "article-name" tags
glow meta add "article-name" tags kafka eventhub
glow meta set "article-name" status "active"
```

### Management
```bash
glow move "old-name" "new-name"
glow delete "article-name"
```

## Mandatory Rules

- **Tagging**: Every article MUST have `tags` metadata
- **Cross-linking**: Use `[[folder/article]]` wikilinks in content
- **Search first**: Always search before writing to avoid duplicates

---

**Work silently.** Only report glow activities if asked or critical context found.