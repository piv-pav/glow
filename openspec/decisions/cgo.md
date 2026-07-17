# CGO Decision

Version: 1  
Updated: 2026-07-17

## Agreement

Glow is and MUST remain **CGO-free**. No dependency that requires CGO is permitted. The pure-Go SQLite driver (`modernc.org/sqlite`) is used instead of `mattn/go-sqlite3`. All other dependencies must also be pure Go.

## Rationale

CGO-free enables `go install codeberg.org/pivpav/glow@latest` as the sole distribution mechanism — no C toolchain required on the user's machine. It also simplifies cross-platform builds and removes a class of build environment problems entirely.

## Constraints

- `mattn/go-sqlite3` and any other CGO-dependent package are permanently banned.
- Every new dependency MUST be verified CGO-free before adoption.
- Build tooling MUST NOT introduce CGO (no `cgo` flags, no `#include`).

## Compliance

All existing dependencies comply. No remediation needed.

## Notes

- `modernc.org/sqlite` provides full FTS5 support — functional parity with `mattn/go-sqlite3` for glow's use case.
