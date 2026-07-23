## ADDED Requirements

### Requirement: Upgrade to latest release
The system SHALL check for a newer version on GitHub and upgrade the binary via `go install` when `glow upgrade` is run.

#### Scenario: Upgrade available
- **WHEN** `glow upgrade` is run and a newer version exists on GitHub
- **THEN** the new version is installed via `go install github.com/piv-pav/glow@<tag>`

#### Scenario: Already up to date
- **WHEN** `glow upgrade` is run and current version matches latest
- **THEN** "Already up to date" (or similar) message is printed; no install performed

#### Scenario: Upgrade failure
- **WHEN** GitHub is unreachable
- **THEN** command exits with an error describing the failure
