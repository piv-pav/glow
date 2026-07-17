## ADDED Requirements

### Requirement: Tag articles on create
The system SHALL accept one or more `--tag` flags on `glow create`. Tags SHALL be stored in article frontmatter. Each `--tag` value MAY be a comma-separated list (e.g. `--tag go,cli`) or repeated flags (e.g. `--tag go --tag cli`) — both are equivalent.

#### Scenario: Single tag
- **WHEN** `glow create my-article --content "x" --tag go` is run
- **THEN** article has tag `go`

#### Scenario: Comma-separated tags
- **WHEN** `--tag go,cli,wiki` is provided
- **THEN** article has tags `go`, `cli`, and `wiki`

#### Scenario: Multiple flags
- **WHEN** `--tag go --tag cli` is provided
- **THEN** article has both tags

### Requirement: Add and remove tags on update
The system SHALL add tags via `--tag` and remove tags via `--untag` on `glow update`. Both flags accept comma-separated values or repeated flags. Either flag MAY be used independently. A tag-only update (no content flags) SHALL be valid.

#### Scenario: Add tag
- **WHEN** `glow update my-article --tag newtag` is run
- **THEN** `newtag` is added; existing tags unchanged

#### Scenario: Remove tag
- **WHEN** `glow update my-article --untag oldtag` is run
- **THEN** `oldtag` is removed; other tags unchanged

#### Scenario: Comma-separated untag
- **WHEN** `glow update my-article --untag go,llm` is run
- **THEN** both `go` and `llm` are removed

#### Scenario: Add and remove in one call
- **WHEN** `glow update my-article --tag a --untag b` is run
- **THEN** `a` is added and `b` is removed atomically

### Requirement: Add and remove tags on append
The system SHALL accept `--tag` and `--untag` on `glow append`, with the same comma-separated / repeated flag syntax. A tag-only append (no content) SHALL be valid.

#### Scenario: Tag-only append
- **WHEN** `glow append my-article --tag extra` is run
- **THEN** tag `extra` is added; content is unchanged

#### Scenario: Comma-separated tags on append
- **WHEN** `glow append my-article --tag go,llm` is run
- **THEN** both `go` and `llm` are added

### Requirement: Read tags
The system SHALL print article tags one per line when `glow read --tags` (or `-t`) is used.

#### Scenario: Read tags
- **WHEN** `glow read my-article --tags` is run
- **THEN** each tag is printed on its own line

### Requirement: Auto-managed timestamps
The system SHALL automatically set `created` (RFC3339) on article creation and update `modified` (RFC3339) on every write. These fields SHALL NOT require user input.

#### Scenario: Created set on create
- **WHEN** an article is created
- **THEN** its frontmatter contains `created` with the creation timestamp

#### Scenario: Modified updated on update
- **WHEN** an article is updated
- **THEN** its `modified` timestamp is bumped to current time
