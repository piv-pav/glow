---
name: glow
description: GLOW wiki operations. Use for ALL glow commands, wiki searches, knowledge storage/retrieval. Essential for remembering context, project info, engineering patterns, configurations. Must use when user mentions wiki, memory, remember, or need persistent storage.
---

# GLOW Wiki Skill

**CRITICAL: Use this skill for ANY `glow` command or wiki operation**

## When to Use This Skill

**ALWAYS use this skill when**:
- User says: "remember", "note", "memory", "wiki", "save this", "store"
- Running ANY `glow` command (search, read, create, update, append, etc.)
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

Second paragraph." --tag "value"
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

### Diff-based Update (SEARCH/REPLACE blocks)

Update an article by piping SEARCH/REPLACE blocks via STDIN — the format most AI tools emit for text edits. Good for surgical, multi-spot edits without re-pasting whole sections:

```bash
printf '<<<<<<< SEARCH\nexact old text\n=======\nnew text\n>>>>>>> REPLACE\n' | glow update "article-name" --diff

# Scope to a single section (SEARCH only needs to be unique within that section)
printf '<<<<<<< SEARCH\nold\n=======\nnew\n>>>>>>> REPLACE\n' | glow update "article-name" --diff --section "Status"
```

- Multiple blocks applied in order; each SEARCH must match **exactly once** (errors on 0 or >1 matches).
- Edits are **atomic** — article unchanged if any block fails.
- `--diff` reads only from STDIN; cannot combine with `--content`/`--stdin`. Combines with `--section`, `--tag`, `--untag`.

### Tags
```bash
glow update "article-name" --tag kafka
glow update "article-name" --untag oldtag
glow update "article-name" --tag a,b --untag c
```

### Management
```bash
glow move "old-name" "new-name"
glow delete "article-name"
glow delete "article-name" --section "Section Heading"
```

### Wiki Management
```bash
# List wikis
glow wiki-list

# Use a specific wiki
glow -w work search "topic"
glow -w work read "article"

# Create/delete wikis
glow init mywork
glow wiki-delete mywork

# Export/Import
glow export default /tmp/backup.json
glow import work /tmp/backup.json
```
## Mandatory Rules

- **Tagging**: Every article MUST have tags. Use `--tag` on create/update.
- **Cross-linking**: Use `[[folder/article]]` wikilinks in content
- **Search first**: Always search before writing to avoid duplicates

---

**Work silently.** Only report glow activities if asked or critical context found.