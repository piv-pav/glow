## ADDED Requirements

### Requirement: Create article
The system SHALL create a new article with a given name, content, and optional tags. Article name MAY include slash-delimited path segments (e.g., `folder/article`). Content MUST be provided via `--content` flag or `--stdin`; omitting both SHALL be an error. Duplicate names SHALL be rejected by the storage backend.

#### Scenario: Create with inline content
- **WHEN** `glow create my-article --content "# Hello"` is run
- **THEN** article is stored and "Created article: my-article" is printed

#### Scenario: Create with stdin
- **WHEN** content is piped via `echo "body" | glow create my-article --stdin`
- **THEN** article is stored with that content

#### Scenario: Create with tags
- **WHEN** `glow create my-article --content "body" --tag go --tag cli` is run
- **THEN** article is stored with tags `go` and `cli`

#### Scenario: Create without content fails
- **WHEN** `glow create my-article` is run with no `--content` or `--stdin`
- **THEN** command exits with error "must specify one of: --content or --stdin"

#### Scenario: Nested path article
- **WHEN** `glow create projects/myproject --content "body"` is run
- **THEN** article stored under path `projects/myproject`

### Requirement: Read article
The system SHALL output the full content of an article. Aliases `show` and `cat` SHALL be accepted.

#### Scenario: Read full content
- **WHEN** `glow read my-article` is run
- **THEN** article content is printed to stdout

#### Scenario: Read non-existent article
- **WHEN** `glow read missing` is run
- **THEN** command exits with error

#### Scenario: List tags only
- **WHEN** `glow read my-article --tags` (or `-t`) is run
- **THEN** only tags are printed, one per line

#### Scenario: List sections
- **WHEN** `glow read my-article --sections` is run
- **THEN** section headings are listed with level indicators

### Requirement: Update article
The system SHALL replace article content or a named section. Content MUST be provided via `--content`, `--stdin`, `--diff`, or tag flags; omitting all SHALL be an error. `--diff` SHALL NOT be combined with `--content` or `--stdin`.

#### Scenario: Update full content
- **WHEN** `glow update my-article --content "new body"` is run
- **THEN** article content is replaced and "Updated article: my-article" is printed

#### Scenario: Update section
- **WHEN** `glow update my-article --section "Heading" --content "new"` is run
- **THEN** only that section's content is replaced

#### Scenario: Tag-only update
- **WHEN** `glow update my-article --tag newtag --untag oldtag` is run with no content flags
- **THEN** tags are updated without touching content

#### Scenario: --diff with --content fails
- **WHEN** `glow update my-article --diff --content "x"` is run
- **THEN** command exits with error "--diff cannot be combined with --content or --stdin"

### Requirement: Append to article
The system SHALL append content to an existing article, or to a named section. Tags MAY be modified in the same operation.

#### Scenario: Append content
- **WHEN** `glow append my-article --content "more text"` is run
- **THEN** content is appended to the article body

#### Scenario: Append to section
- **WHEN** `glow append my-article --section "Notes" --content "item"` is run
- **THEN** content is appended within that section

#### Scenario: Append with tag
- **WHEN** `glow append my-article --content "body" --tag new` is run
- **THEN** content is appended and tag `new` is added

#### Scenario: Tag-only append
- **WHEN** `glow append my-article --tag foo` is run with no content
- **THEN** tag `foo` is added; content unchanged

### Requirement: Delete article
The system SHALL delete an article by name, or delete a named section from an article.

#### Scenario: Delete article
- **WHEN** `glow delete my-article` is run
- **THEN** article is removed and "Deleted article: my-article" is printed

#### Scenario: Delete section
- **WHEN** `glow delete my-article --section "Old Section"` is run
- **THEN** only that section is removed; article remains

#### Scenario: Delete non-existent article
- **WHEN** `glow delete missing` is run
- **THEN** command exits with error

### Requirement: Move/rename article
The system SHALL move or rename an article. All `[[oldName]]` wikilinks in other articles SHALL be rewritten to `[[newName]]` (best-effort; anchor variants are not rewritten).

#### Scenario: Rename article
- **WHEN** `glow move old-name new-name` is run
- **THEN** article is accessible at `new-name`; old name is gone

#### Scenario: Move to folder
- **WHEN** `glow move article projects/article` is run
- **THEN** article is accessible at `projects/article`

#### Scenario: Wikilink rewriting
- **WHEN** article `foo` is moved to `bar` and another article contains `[[foo]]`
- **THEN** that reference is rewritten to `[[bar]]`
