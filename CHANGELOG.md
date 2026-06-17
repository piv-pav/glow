# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.8.3] - 2026-06-17

### Added
- **rqlite**: `disable_discovery` option for reverse proxy setups (prevents gorqlite from resolving internal container addresses)

## [0.8.2] - 2026-06-17

### Added
- **rqlite backend**: Distributed SQLite via rqlite cluster support with FTS5 search, automatic write forwarding to leader, and local read-replica availability

## [0.8.1] - 2026-06-17

### Changed
- **Auto-discovery**: No longer prompts for confirmation — detected wikis are registered automatically with informational output

## [0.8.0] - 2026-06-17

### Added
- **Multi-backend storage**: SQLite (default), PostgreSQL, and file-based backends
- **`glow init`**: Interactive wiki creation with backend selection (sqlite/pgsql/files)
- **`glow wiki-delete`**: Remove a wiki from config (and local data for files/sqlite backends)
- **PostgreSQL support**: Full-text search via tsvector + GIN index, `ts_rank` relevance ordering
- **Auto-discovery**: On first run without config, discovers existing wikis in data directory and registers them automatically
- **Wiki name validation**: Names must match `[a-zA-Z0-9][a-zA-Z0-9_-]*` — prevents path traversal
- **Export/Import**: `glow export <wiki> <file>` and `glow import <wiki> <file>` for migration between backends
- **Configuration file**: `~/.config/glow/glow.yaml` stores wiki registry and backend settings

### Changed
- **Bleve index only for files backend**: SQLite/PgSQL use native full-text search, no Bleve overhead (~2x faster writes)
- **Search output**: Removed score display (not useful for AI consumers)
- **`glow init`**: Defaults to wiki name "default" when no argument given

### Refactored
- **Storage layer deduplicated**: Extracted shared `sqlStore` base (Create/Read/Update/Delete/Move/List) — SQLite and PgSQL only differ in placeholder style and Search implementation
- **Removed `NewSearcher` wrapper**: Search command type-asserts Store to Searcher directly
- **`strings.Cut`** replaces manual filter parsing loop
- **`filepath.Ext`** replaces `strings.HasSuffix` for file detection
- **Pre-allocated slices** in index search (avoids repeated growth)
- **sqlite.go**: 294 → 134 lines
- **pgsql.go**: 279 → 128 lines

### Performance (100 articles benchmark)
| Operation | Files | SQLite | PgSQL |
|-----------|-------|--------|-------|
| CREATE 100 | 12.2s | 5.7s | 7.1s |
| SEARCH 10 | 0.75s | 0.64s | 0.78s |
| UPDATE 20 | 2.3s | 1.2s | 1.6s |
| DELETE 100 | 8.0s | 5.5s | 6.9s |

## [0.7.1] - 2026-06-09

### Fixed
- `AddTags()` now deduplicates against existing tags
- `Serialize()` no longer mutates article timestamps (moved to storage layer)

### Improved
- Search: keyword analyzer for `tags` and `path` fields — exact match, no false positives from stemming
- Search: tags indexed as individual terms (array) instead of space-joined string
- `List()` uses `filepath.WalkDir` instead of `filepath.Walk` (avoids extra stat per entry)
- `move` command no longer performs redundant file read before move
- Extracted `articleToDoc()` helper — eliminates duplicate metadata→index conversion

### Refactored
- Renamed `Article.Metadata` → `Article.Frontmatter` (clarity: it's YAML frontmatter, not generic metadata)
- Renamed `SearchResult.Metadata` → `SearchResult.Fields`
- Removed dead generic metadata functions (`SetMetadata`, `AddMetadata`, `DeleteMetadata`, `GetMetadataString`, `GetMetadataArray`)
- Renamed `metadata.go` → `tags.go`, `metadata_test.go` → `tags_test.go`
- Added `GetTags()` method as the single way to read tags

## [0.7.0] - 2026-06-05

### Breaking Changes
- **Removed `meta` command**: `glow meta set/add/delete/get` no longer exist.
- **Removed `--meta` flag**: Use `--tag` and `--untag` instead.
- **Tags only**: Metadata is now limited to `created`, `modified`, and `tags` fields.

### Added
- `--tag` flag on `create` and `update`: Add tags (comma-separated or repeated: `--tag go --tag cli`)
- `--untag` flag on `update`: Remove tags (same syntax as `--tag`)
- `article.AddTags()`, `article.RemoveTags()`, `article.SetTags()` methods

### Changed
- `update` can now modify tags without `--content` (e.g., `glow update "x" --tag new --untag old`)
- Updated tests to cover single, repeated, and comma-separated `--tag`/`--untag` operations

### Refactored
- Extracted `modifyArticle()` helper for read→modify→update→index pattern
- Removed `splitLines`/`joinLines` wrappers (inlined in `read.go`)
- Removed `parseMeta()` helper (no longer needed)

## [0.6.0] - 2026-06-05

### Breaking Changes
- **`update`**: Removed `--meta` flag. Use `glow meta set/add` for all metadata changes.
- **`append --section`**: Now errors if section not found (previously silently created section with hardcoded `##` prefix)

### Fixed
- `AppendToSection` no longer silently creates malformed sections on missing heading
- `DeleteSection` had wrong doc comment (copy-paste artifact from `AppendToSection`)

### Refactored
- `readContent()` helper eliminates duplicated stdin/`--content` branching across create/update/append
- `parseMeta()` helper eliminates duplicated metadata parsing in create
- `article.New()` now receives content directly instead of post-assignment
- `os.ReadFile("/dev/stdin")` replaced with `io.ReadAll(os.Stdin)` (more portable)
- Removed stale `cmd/wiki/` directory (dead code with old module path)
- `search_test.go`: replaced fragile `sh -c echo` pipe with direct `runWiki` call

### Docs
- `--content` examples now use multiline strings (preferred over `\n` escapes)
- `\n` in `--content` is still interpreted as fallback
- Removed stale `verify` command, alias commands (`show`, `cat`), `printf` examples

## [0.5.2] - 2026-05-27

### Changed
- **Repository**: Migrated to Codeberg (https://codeberg.org/pivpav/glow)
- **Module path**: `git.netra.pivpav.com/public/glow` → `codeberg.org/pivpav/glow`
- All imports and documentation updated to new repository location
- Fresh commit history for public release

## [0.5.1] - 2026-05-26

### Fixed
- **UpdateSection**: Fixed duplicate headers when content includes section heading
  - Detect when newContent starts with matching header
  - Skip adding existing heading to prevent duplicates
  - Improved blank line spacing between sections

## [0.5.0] - 2026-05-25

### Changed
- **Performance**: Index fields cached at creation time, reused in search (eliminates repeated field enumeration)
- **Performance**: Metadata flattening inlined in index operations (removes redundant map allocations)
- **Performance**: `unescapeContent()` fast path for strings without escape sequences
- **Code quality**: Introduced `withIndex()` helper pattern across all commands (eliminates boilerplate, guarantees cleanup)
- **Test organization**: Split 1159-line integration test into 5 focused files (helpers, article, metadata, search, wiki)

### Removed
- `--editor` flag and interactive editor support (never needed in LLM/automation context)
- `verify` command (unnecessary index health check)
- `Index.Verify()` method
- `Article.GetAllMetadataForIndex()` method (logic inlined)

## [0.4.5] - 2026-05-24

### Fixed
- `--content` flag now interprets escape sequences (`\n`, `\t`, etc.) instead of storing them literally
- Invalid escape sequences in `--content` return a clear error suggesting `--stdin` as alternative

## [0.4.4] - 2026-05-24

### Changed
- Renamed `wiki-verify` → `verify` and `wiki-rebuild` → `rebuild`

## [0.4.3] - 2026-05-24

### Changed
- Simplified text search: all metadata fields (including `tags`, custom fields) use same boost 1.5, content stays at 1.0
- Text search now dynamically discovers all indexed fields via Bleve API — no hardcoded field list

## [0.4.2] - 2026-05-24

### Fixed
- Tags and metadata fields ignored in plain text search — now searches across `tags`, `project`, `path` in addition to `content`

### Changed
- Search scoring priority: tags (2.0) > project/path (1.5) > content (1.0)

## [0.3.1] - 2026-05-24

### Fixed
- `glow --version` shows `v0.3.1` instead of `dev` when installed via plain `go install` (embed VERSION file, set version in init())

## [0.3.0] - 2026-05-24

### Changed
- **Binary renamed**: `wiki` → `glow` (matches repo name). Install with `go install codeberg.org/pivpav/glow@latest`
- **Module path**: `github.com/pavelpivovarov/glow` → `codeberg.org/pivpav/glow`
- **Restructured**: `main.go` at repo root, subcommands in `tools/` package, removed `cmd/`
- **Build system**: justfile targets use `glow` binary name
- `--meta` flags changed from `StringSliceVar` to `StringArrayVar` (commas in values preserved)

### Added
- `update --meta` flag for updating metadata alongside content changes
- Integration tests for `list`, `read`, `wiki-create`, `wiki-list`, `wiki-verify`, `update --meta` (30 total)

### Fixed
- `update` no longer opens editor when only `--meta` flags provided (hangs in non-interactive use)
- All import paths updated to new module path

## [0.2.1] - 2026-05-24

### Added
- `--section` flag for `wiki delete` to remove a specific section from an article
- `Article.DeleteSection()` method for section-level deletion
- Integration test for `delete --section` functionality

## [0.2.0] - 2026-05-23

### Added
- `--stdin` flag for `wiki append` to read content from stdin
- `--content` flag for `wiki append` to provide content directly
- `wiki meta get` subcommand to read metadata field values

### Changed
- `wiki append` now requires explicit `--content` or `--stdin` flag (positional content arg removed)
- Consistent CLI API across `create`, `update`, and `append` — all use `--content` / `--stdin`

### Removed
- Positional content argument from `wiki append`

## [0.1.4] - 2026-05-19

### Changed
- Made `wiki create` require explicit flag: `--content`, `--stdin`, or `--editor`
- Prevents accidental editor invocation in non-interactive contexts (LLM tools, automation)

### Fixed
- Editor no longer opens by default when no content flags provided

## [0.1.3] - 2026-05-19

### Added
- `--content` flag for `wiki create` to provide content directly without editor
- `--stdin` flag for `wiki create` to read content from stdin
- `--content` flag for `wiki update` to provide content directly without editor
- `--stdin` flag for `wiki update` to read content from stdin

### Changed
- Updated integration tests to use `--stdin` flag explicitly
- Updated README.md with new LLM-friendly usage examples
- Updated SKILL.md with non-interactive workflows

### Fixed
- Editor hang when `wiki create` or `wiki update` used in non-interactive contexts (pipes, LLM tools)

## [0.1.2] - 2026-05-18

### Added
- Initial stable release with core wiki functionality

## [0.1.1] - 2026-05-17

### Added
- Section-based operations (read, update, append)
- Metadata management commands

## [0.1.0] - 2026-05-16

### Added
- Initial release
- Article creation, reading, updating, deletion
- Full-text search with Bleve
- Multi-wiki support
- Metadata (tags, projects, custom fields)
- Nested folder structure
