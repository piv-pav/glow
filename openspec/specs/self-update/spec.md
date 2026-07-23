## ADDED Requirements

### Requirement: Upgrade to latest release
The system SHALL print manual migration instructions to GitHub when `glow upgrade` is run.

#### Scenario: Upgrade run
- **WHEN** `glow upgrade` is run
- **THEN** manual migration instructions are printed directing user to `go install github.com/piv-pav/glow@latest`

#### Scenario: Already on GitHub
- **WHEN** binary was installed from GitHub (v0.11.2+)
- **THEN** `glow upgrade` fetches latest GitHub tag and upgrades via `go install github.com/piv-pav/glow@<tag>`
