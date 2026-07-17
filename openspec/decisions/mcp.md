# MCP Decision

Version: 1  
Updated: 2026-07-17

## Agreement

The MCP server (`glow mcp`) is a **pure translation layer** — a thin adapter over the existing storage and article logic. It exposes all wiki operations as MCP tools over **stdio transport only**. HTTP transport is not supported. No new behaviour is reachable via MCP that is not already reachable via the CLI. CLI code is never modified to accommodate MCP.

## Rationale

**Pure translation layer:** Prevents divergence between CLI and MCP behaviour. MCP calls the same `withStore`, `modifyArticle`, and `article.*` functions as CLI commands — no duplicate logic, no separate data models.

**Stdio only, no HTTP:** HTTP transport was added in 0.10.1 and removed in 0.10.3. It violated the local-first, single-writer architecture and was never tested with multiple concurrent clients. rqlite is the correct solution for network/multi-client access — not an HTTP MCP endpoint.

**No CLI changes for MCP:** MCP wraps existing storage calls. Adding a new CLI command does not require touching MCP code, and adding MCP tools does not require touching CLI code.

## Constraints

- MCP MUST NOT introduce its own business logic, data models, or side effects.
- HTTP transport MUST NOT be re-added. If network MCP access is needed, use rqlite backend + stdio MCP.
- Every MCP tool must directly call the same storage functions as its CLI equivalent.

## Compliance

All existing code complies following the 0.10.3 removal of HTTP transport. No remediation needed.

## Notes

- Startup banner printed to stderr: `GLOW MCP Server vX.Y.Z — STDIO`
- Each tool accepts optional `wiki_name` parameter, defaulting to `"default"`. `--wiki` flag sets server-wide default.
