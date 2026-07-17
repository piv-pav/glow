## ADDED Requirements

### Requirement: Create named wiki
The system SHALL provide two commands for creating a wiki:

- `glow init [name]` — interactive; prompts for backend (`sqlite`/`rqlite`) and, if rqlite, prompts for URL, user, and password (password input without echo). Name defaults to `default` if omitted.
- `glow wiki-create <name>` — non-interactive; always creates a SQLite wiki. No prompts.

#### Scenario: init prompts for backend
- **WHEN** `glow init mywiki` is run
- **THEN** user is prompted "Storage backend [sqlite/rqlite] (default: sqlite):"

#### Scenario: init with rqlite prompts for connection details
- **WHEN** user enters `rqlite` at the backend prompt
- **THEN** user is prompted for URL, user, and password (password read without echo)

#### Scenario: init defaults to sqlite
- **WHEN** user presses enter at the backend prompt
- **THEN** wiki is created with SQLite backend

#### Scenario: init defaults wiki name to default
- **WHEN** `glow init` is run with no name argument
- **THEN** wiki is created with name `default`

#### Scenario: wiki-create always uses SQLite
- **WHEN** `glow wiki-create mywiki` is run
- **THEN** a new SQLite wiki `mywiki` is created without prompts

### Requirement: Delete named wiki
The system SHALL delete a named wiki's database file via `glow wiki-delete <name>`.

#### Scenario: Delete wiki
- **WHEN** `glow wiki-delete mywiki` is run
- **THEN** the wiki's `.db` file is removed

### Requirement: List wikis
The system SHALL list all configured wikis via `glow wiki-list`.

#### Scenario: List wikis
- **WHEN** `glow wiki-list` is run
- **THEN** names of all registered wikis are printed

### Requirement: Select wiki with -w flag
The system SHALL target a named wiki when the `-w <name>` global flag is provided on any command. Without `-w`, the `default` wiki is used.

#### Scenario: Read from named wiki
- **WHEN** `glow -w work read my-article` is run
- **THEN** the article is read from the `work` wiki

#### Scenario: Default wiki used without flag
- **WHEN** `glow read my-article` is run without `-w`
- **THEN** the article is read from the `default` wiki

### Requirement: Wiki storage layout
SQLite wikis SHALL be stored as a single `.db` file at `<data-dir>/<name>.db`. The data directory SHALL default to `~/.local/share/glow/wiki/` following XDG conventions (overridable via `GLOW_DATA` env or `db_path` config key). rqlite wikis have no local file — they target a remote cluster via URL.

#### Scenario: SQLite wiki file location
- **WHEN** a SQLite wiki named `work` is created
- **THEN** its database is at `~/.local/share/glow/wiki/work.db`

#### Scenario: rqlite wiki has no local file
- **WHEN** a wiki is configured with `backend: rqlite`
- **THEN** no `.db` file is created; all operations go to the rqlite cluster URL

### Requirement: Wiki name validation
The system SHALL reject wiki names that could cause path traversal (e.g., names containing `..` or `/`).

#### Scenario: Invalid wiki name rejected
- **WHEN** `glow wiki-create "../evil"` is run
- **THEN** command exits with a validation error
