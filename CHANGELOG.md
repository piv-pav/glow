# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
