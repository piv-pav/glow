# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
