## ADDED Requirements

### Requirement: Export wiki to tar.gz
The system SHALL export all articles in a wiki to a `.tar.gz` archive of markdown files via `glow export <wiki-name> <output.tar.gz>`. Each article SHALL be written as `<article-name>.md` with frontmatter preserved. The target wiki MUST exist.

#### Scenario: Export default wiki
- **WHEN** `glow export default /tmp/backup.tar.gz` is run
- **THEN** a tar.gz archive is created containing one `.md` file per article

#### Scenario: Export non-existent wiki fails
- **WHEN** `glow export missing /tmp/out.tar.gz` is run
- **THEN** command exits with error "wiki not found: missing"

### Requirement: Import wiki from tar.gz
The system SHALL import articles from a `.tar.gz` archive into an existing wiki via `glow import <wiki-name> <input.tar.gz>`. Only `.md` files are processed. Articles that already exist in the target wiki SHALL be skipped with a warning. The target wiki MUST exist before importing.

#### Scenario: Import to existing wiki
- **WHEN** `glow import work /tmp/backup.tar.gz` is run
- **THEN** articles from the archive are written to the `work` wiki

#### Scenario: Already-existing articles skipped
- **WHEN** an article from the archive already exists in the target wiki
- **THEN** it is skipped and a warning is printed to stderr; import continues

#### Scenario: Import to non-existent wiki fails
- **WHEN** `glow import missing /tmp/backup.tar.gz` is run
- **THEN** command exits with error prompting to create the wiki first

#### Scenario: Import enables backend migration
- **WHEN** an SQLite wiki is exported and the archive is imported into an rqlite wiki
- **THEN** all articles are accessible via the rqlite backend
