# Search Decision

Version: 1  
Updated: 2026-07-17

## Agreement

Search uses **FTS5 full-text search with BM25 ranking**. Multiple text terms are OR'd. Results include a content snippet with matched terms highlighted and a total count. Filters (`tag:value`, `path:prefix/`) are **glow-native tokens** parsed out of the query string before FTS execution — they are not passed to FTS5. Raw FTS5 syntax (NEAR, NOT, column filters) is not exposed. The dedicated `--filter` flag was removed in favour of inline filter syntax in the query string.

## Rationale

**FTS5 + BM25:** Provides relevance-ranked full-text search without external dependencies. Both SQLite and rqlite backends support it natively.

**OR semantics for multiple terms:** More useful for a personal wiki than AND — partial matches still surface relevant articles.

**Inline filter syntax (`tag:go path:projects/`):** Composable in a single query string argument. More natural for shell use than a separate `--filter` flag. The `--filter` flag was removed in 0.9.5 as a duplicate.

**No raw FTS5 syntax exposure:** Keeps the query model simple and consistent across backends. FTS5 operator syntax would leak implementation details and differ between backends.

## Constraints

- Filter tokens (`tag:`, `path:`) MUST be stripped before passing query to FTS5.
- Raw FTS5 syntax MUST NOT be passed through to the engine.
- `--filter` flag MUST NOT be re-added — use inline `tag:`/`path:` syntax.

## Compliance

All existing code complies. No remediation needed.

## Notes

- Terms are treated as literals — `go-yaml` and `self-hosting` work as expected without quoting.
- Search limit configurable via `--limit`; total count shown even when results are truncated.
