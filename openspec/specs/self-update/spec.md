## ADDED Requirements

### Requirement: Upgrade to latest release
The system SHALL check for a newer version on Codeberg and upgrade the binary via `go install` when `glow upgrade` is run.

#### Scenario: Upgrade available
- **WHEN** `glow upgrade` is run and a newer version exists on Codeberg
- **THEN** the new version is installed via `go install codeberg.org/pivpav/glow@<tag>`

#### Scenario: Already up to date
- **WHEN** `glow upgrade` is run and current version matches latest
- **THEN** "Already up to date" (or similar) message is printed; no install performed

#### Scenario: Upgrade failure
- **WHEN** Codeberg is unreachable
- **THEN** command exits with an error describing the failure
