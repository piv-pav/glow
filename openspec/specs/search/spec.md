## ADDED Requirements

### Requirement: Full-text search with BM25 ranking
The system SHALL search article content using FTS5 full-text search with BM25 ranking. Multiple terms SHALL be treated as OR. Results SHALL include a content snippet with matched terms highlighted. Total count SHALL be shown even when results are limited.

#### Scenario: Basic search
- **WHEN** `glow search "golang wiki"` is run
- **THEN** articles matching "golang" or "wiki" are returned, ranked by relevance

#### Scenario: Search with limit
- **WHEN** `glow search "query" --limit 5` is run
- **THEN** at most 5 results are returned; total count still shown if more exist

#### Scenario: No results
- **WHEN** `glow search "xyzzy-not-in-any-article"` is run
- **THEN** "No results found" is printed

### Requirement: Tag filter in search query
The system SHALL filter results by tag when `tag:<value>` is embedded in the search query. Tag filter SHALL be parsed out before FTS execution.

#### Scenario: Tag filter
- **WHEN** `glow search "query tag:go"` is run
- **THEN** only articles tagged `go` matching the query text are returned

#### Scenario: Tag-only filter
- **WHEN** `glow search "tag:go"` is run
- **THEN** all articles tagged `go` are returned

### Requirement: Path filter in search query
The system SHALL filter results to articles whose name starts with the given prefix when `path:<value>` is embedded in the search query.

#### Scenario: Path filter
- **WHEN** `glow search "query path:projects/"` is run
- **THEN** only articles under the `projects/` path prefix are returned

### Requirement: List all articles
The system SHALL list all articles in the wiki via `glow list`.

#### Scenario: List articles
- **WHEN** `glow list` is run
- **THEN** all article names are printed

#### Scenario: List on empty wiki
- **WHEN** `glow list` is run on a wiki with no articles
- **THEN** command exits cleanly with no output or "No articles found"
