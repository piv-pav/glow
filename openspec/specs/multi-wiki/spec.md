# Multi-Wiki Specification

## Purpose

Glow supports multiple named wikis on a single machine. Each wiki is an isolated store of articles with its own backend configuration. The `wiki-create`, `wiki-delete`, and `wiki-list` commands manage wikis. The `-w` global flag selects which wiki article commands target.

## Requirements

### Requirement: Create named wiki
The system SHALL provide a single command for creating a wiki: `glow wiki-create <name>`.

The `<name>` positional argument is always required. The command requires exactly one of two mutually exclusive mode flags:
- `--interactive` / `-i` — interactive mode; prompts for backend (`sqlite`/`rqlite`) and, if rqlite, prompts for URL, user, and password (password input without echo).
- `--backend <b>` / `-b <b>` — non-interactive mode; creates wiki with the specified backend (`sqlite` or `rqlite`).

For rqlite non-interactive mode, the following additional flags are available:
- `--url <url>` — rqlite cluster URL (required for rqlite)
- `--user <user>` — rqlite username (optional)
- `--password <pass>` — rqlite password (optional)
- `--password-stdin` — read rqlite password from stdin instead of `--password` (mutually exclusive with `--password`)
- `--level <level>` — rqlite consistency level: `none`, `weak`, `strong` (optional; default: `weak`)

Omitting both `-i` and `-b` is an error. Providing both is an error.

#### Scenario: wiki-create with -b sqlite creates wiki
- **WHEN** `glow wiki-create mywiki -b sqlite` is run
- **THEN** a new SQLite wiki named `mywiki` is created and exits 0

#### Scenario: wiki-create with -b rqlite creates wiki
- **WHEN** `glow wiki-create mywiki -b rqlite --url http://localhost:4001` is run
- **THEN** a new rqlite wiki named `mywiki` is created and exits 0

#### Scenario: wiki-create rqlite requires --url
- **WHEN** `glow wiki-create mywiki -b rqlite` is run without `--url`
- **THEN** command exits non-zero with an error indicating `--url` is required

#### Scenario: wiki-create rqlite with credentials
- **WHEN** `glow wiki-create mywiki -b rqlite --url http://localhost:4001 --user alice --password secret` is run
- **THEN** a new rqlite wiki named `mywiki` is created with the supplied credentials and exits 0

#### Scenario: wiki-create rqlite with --password-stdin
- **WHEN** `glow wiki-create mywiki -b rqlite --url http://localhost:4001 --user alice --password-stdin` is run with password piped to stdin
- **THEN** a new rqlite wiki named `mywiki` is created using the piped password and exits 0

#### Scenario: --password and --password-stdin together is an error
- **WHEN** `glow wiki-create mywiki -b rqlite --url http://localhost:4001 --password secret --password-stdin` is run
- **THEN** command exits non-zero with an error indicating the flags are mutually exclusive

#### Scenario: wiki-create rqlite with consistency level
- **WHEN** `glow wiki-create mywiki -b rqlite --url http://localhost:4001 --level strong` is run
- **THEN** a new rqlite wiki named `mywiki` is created with `strong` consistency and exits 0

#### Scenario: wiki-create with -i prompts for backend
- **WHEN** `glow wiki-create mywiki -i` is run interactively
- **THEN** user is prompted "Storage backend [sqlite/rqlite] (default: sqlite):"

#### Scenario: wiki-create -i with rqlite prompts for connection details
- **WHEN** user enters `rqlite` at the backend prompt
- **THEN** user is prompted for URL, user, and password (password read without echo)

#### Scenario: wiki-create -i defaults to sqlite
- **WHEN** user presses enter at the backend prompt
- **THEN** wiki is created with SQLite backend

#### Scenario: wiki-create with no mode flag is an error
- **WHEN** `glow wiki-create mywiki` is run with neither `-i` nor `-b`
- **THEN** command exits non-zero with an error message indicating `-i` or `-b` is required

#### Scenario: wiki-create with no name is an error
- **WHEN** `glow wiki-create` is run with no positional argument
- **THEN** command exits non-zero with an error message indicating name is required

#### Scenario: wiki-create -i and -b together is an error
- **WHEN** `glow wiki-create mywiki -i -b sqlite` is run
- **THEN** command exits non-zero with an error indicating the flags are mutually exclusive

#### Scenario: wiki-create duplicate name is an error
- **WHEN** `glow wiki-create mywiki -b sqlite` is run and `mywiki` already exists
- **THEN** command exits non-zero with an error indicating the wiki already exists

#### Scenario: -b shorthand accepted
- **WHEN** `glow wiki-create mywiki -b sqlite` is run
- **THEN** it behaves identically to `glow wiki-create mywiki --backend sqlite`

#### Scenario: -i shorthand accepted
- **WHEN** `glow wiki-create mywiki -i` is run
- **THEN** it behaves identically to `glow wiki-create mywiki --interactive`

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
- **WHEN** `glow wiki-create "../evil" -b sqlite` is run
- **THEN** command exits with a validation error
