# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.1] - 2026-05-24

### Fixed
- `glow --version` shows `v0.3.1` instead of `dev` when installed via plain `go install` (embed VERSION file, set version in init())

## [0.3.0] - 2026-05-24

### Changed
- **Binary renamed**: `wiki` → `glow` (matches repo name). Install with `go install git.netra.pivpav.com/public/glow@latest`
- **Module path**: `github.com/pavelpivovarov/glow` → `git.netra.pivpav.com/public/glow`
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
