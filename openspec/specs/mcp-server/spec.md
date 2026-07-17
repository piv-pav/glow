## ADDED Requirements

### Requirement: MCP is a pure translation layer
The MCP server SHALL be a thin adapter over the existing storage and article logic — it MUST NOT introduce its own business logic, data models, or side effects. It SHALL call the same `withStore`, `modifyArticle`, and `article.*` functions that the CLI commands use. No new behaviour SHALL be reachable via MCP that is not already reachable via the CLI. CLI behaviour SHALL NOT be modified to accommodate MCP.

#### Scenario: MCP create uses same storage path as CLI create
- **WHEN** a `create` tool call is made via MCP
- **THEN** the article is stored identically to `glow create` using the same `store.Create` call

#### Scenario: MCP diff update uses same apply logic as CLI
- **WHEN** an `update` tool call with diff blocks is made via MCP
- **THEN** `article.ApplyDiff` is called — the same function used by `glow update --diff`

#### Scenario: No CLI changes needed to support MCP
- **WHEN** a new CLI command is added
- **THEN** MCP can expose it by wrapping the same storage calls with no changes to CLI code

### Requirement: Start MCP server over stdio
The system SHALL start an MCP server on stdio when `glow mcp` is run. The server SHALL expose all wiki operations as MCP tools. A startup banner SHALL be printed to stderr on start.

#### Scenario: Start MCP server
- **WHEN** `glow mcp` is run
- **THEN** server starts, banner "GLOW MCP Server vX.Y.Z — STDIO" is printed to stderr
- **THEN** server accepts JSON-RPC requests on stdin/stdout

### Requirement: MCP tools
The server SHALL expose the following tools: `search`, `list`, `read`, `create`, `update`, `append`, `delete`, `move`. Each tool SHALL accept an optional `wiki_name` parameter defaulting to `"default"`.

#### Scenario: MCP read tool
- **WHEN** a `read` tool call with `{"name": "my-article"}` is sent
- **THEN** the article content is returned

#### Scenario: MCP create tool
- **WHEN** a `create` tool call with name, content, and optional tags is sent
- **THEN** article is created and success response is returned

#### Scenario: wiki_name overrides default
- **WHEN** a tool call includes `{"wiki_name": "work"}`
- **THEN** operation targets the `work` wiki

### Requirement: --wiki flag sets server-wide default wiki
The system SHALL accept `--wiki <name>` on `glow mcp` to set the default wiki for all tool calls on that server instance.

#### Scenario: Default wiki override
- **WHEN** `glow mcp --wiki work` is run
- **THEN** tool calls without explicit `wiki_name` target the `work` wiki

### Requirement: No HTTP transport
The system SHALL only support stdio transport for MCP. HTTP transport SHALL NOT be available.

#### Scenario: No --port flag
- **WHEN** `glow mcp --port 8080` is attempted
- **THEN** command exits with error (unknown flag)
