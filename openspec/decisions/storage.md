# Storage Decision

Version: 1  
Updated: 2026-07-17

## Agreement

Glow uses two storage backends behind a `Store` interface: **SQLite** (default, local file) and **rqlite** (distributed cluster). The default backend is SQLite using `modernc.org/sqlite` — a pure-Go, CGO-free driver. rqlite is the only supported network backend. Each wiki is stored as an independent database (one `.db` file per wiki for SQLite). FTS5 full-text search is enabled on both backends. Storage files follow XDG conventions, defaulting to `~/.local/share/glow/wiki/` (overridable via `GLOW_DATA`).

## Rationale

**Store interface + Searcher extension:** Decouples backend from commands. Backends that don't support FTS return an error on `glow search` rather than silently degrading.

**rqlite as the only network backend:** Provides distributed reads and offline tolerance without introducing a separate server component. HTTP transport on the MCP server was explicitly rejected — rqlite is the only recommended network path for multi-client access.

**One DB file per wiki:** Simple mental model; wikis are completely independent with no shared state.

## Constraints

- New backends must implement the `Store` interface; FTS support requires `Searcher`.
- No HTTP transport on the MCP server — rqlite is the network story.

## Compliance

All existing code complies. No remediation needed.

## Notes

- rqlite uses `?level=weak&disableClusterDiscovery=true` by default (required for reverse-proxy/LB setups).
- `GLOW_DATA` env or `db_path` config key overrides the data directory.
