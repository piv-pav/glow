# Build Decision

Version: 1  
Updated: 2026-07-17

## Agreement

Glow is built with Go 1.25+, **CGO disabled**, using `just` as the task runner. Tests run automatically before every build or install (`just build`, `just install`). The module path is `github.com/piv-pav/glow`. Distribution is via `go install github.com/piv-pav/glow@latest`. Self-upgrade is provided by `glow upgrade` (fetches latest GitHub tag and installs via `go install`).

## Rationale

**Tests before build:** Prevents accidental distribution of broken builds. `just build` / `just install` always run the test suite first.

**go install distribution:** Zero infrastructure — no release binaries, no package manager. Users upgrade via `glow upgrade` which shells out to `go install`.

**Codeberg as canonical home:** Source of truth for version tags, release detection, and `go install` path.

## Constraints

- Tests MUST pass before any build/install step.
- Module path MUST remain `github.com/piv-pav/glow`.

## Compliance

All existing code complies. No remediation needed.

## Notes

- Integration tests run with `GLOW_DATA=/tmp/glow-test-wiki` for isolation.
- `just fmt` formats code; `just test` runs tests only.
