## ADDED Requirements

### Requirement: SQLite backend (default)
The system SHALL use a CGO-free SQLite backend (modernc.org/sqlite) by default. The database file SHALL be located at `<data-dir>/<wiki-name>.db`. FTS5 full-text search SHALL be enabled.

#### Scenario: Default backend is SQLite
- **WHEN** a wiki is created without specifying a backend
- **THEN** storage uses SQLite

#### Scenario: CGO-free build
- **WHEN** glow binary is built
- **THEN** no CGO is required (uses modernc.org/sqlite, not mattn/go-sqlite3)

#### Scenario: Search works on SQLite
- **WHEN** `glow search "term"` is run against an SQLite wiki
- **THEN** FTS5 full-text search returns ranked results

### Requirement: rqlite backend
The system SHALL support rqlite as an alternative distributed backend. rqlite connection SHALL be configured via config file with `url`, `user`, `password`, and optional `level` (consistency: `none`/`weak`/`strong`, default `weak`) fields. Cluster discovery SHALL be disabled by default (required for reverse-proxy/LB setups). rqlite wikis have no local `.db` file.

#### Scenario: rqlite backend configured
- **WHEN** config specifies `backend: rqlite` with a valid URL
- **THEN** all operations target the rqlite cluster

#### Scenario: Consistency level defaults to weak
- **WHEN** `level` is not set in rqlite config
- **THEN** connections use `?level=weak&disableClusterDiscovery=true`

#### Scenario: Custom consistency level
- **WHEN** `level: strong` is set
- **THEN** connections use `?level=strong&disableClusterDiscovery=true`

#### Scenario: Search works on rqlite
- **WHEN** `glow search "term"` is run against a rqlite wiki
- **THEN** FTS5 search returns ranked results via rqlite

### Requirement: Store interface
The system SHALL define a `Store` interface that both backends implement. A `Searcher` interface SHALL extend `Store` for backends that support FTS search. Backends that do not implement `Searcher` SHALL return an error on `glow search`.

#### Scenario: Backend selected by config
- **WHEN** config specifies a backend
- **THEN** the factory returns the correct Store implementation

### Requirement: Config file
The system SHALL read wiki backend configuration from `~/.config/glow/glow.yaml` (XDG config path). Config SHALL support multiple named wikis with independent backends.

#### Scenario: Config with multiple wikis
- **WHEN** config defines `default` (sqlite) and `distributed` (rqlite) wikis
- **THEN** each wiki uses its own backend independently
