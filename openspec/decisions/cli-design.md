# CLI Design Decision

Version: 1  
Updated: 2026-07-17

## Agreement

Glow is a single-binary CLI. Commands are flat (no deep subcommand nesting). All content input uses `--content` (inline) or `--stdin` (pipe) — these are mutually exclusive. `--diff` is a third exclusive input mode for SEARCH/REPLACE editing. The `-w <name>` global flag targets a named wiki; without it, `default` is used. Wiki management commands are prefixed `wiki-` (`wiki-create`, `wiki-delete`, `wiki-list`). `glow init` is the interactive wiki creation entry point.

## Rationale

**Flat command structure:** Keeps the CLI discoverable and scriptable. Commands like `glow create`, `glow update`, `glow search` are direct and composable in shell pipelines.

**--content / --stdin / --diff exclusivity:** Three clearly separated input modes prevent ambiguous combinations. `--diff` is explicitly incompatible with content flags — atomic SEARCH/REPLACE semantics require the full article to be in memory before applying blocks.

**-w global flag over per-command flag:** Consistent wiki targeting across all commands without repeating the flag definition. Mirrors how tools like `kubectl -n` work.

**wiki- prefix for management commands:** Separates lifecycle commands (init, delete, list wikis) from article operation commands (create, read, update, etc.).

**glow init as interactive entry point:** Single command handles both interactive TTY and scripted (flag-driven) wiki creation. `wiki-create` is a legacy shortcut for SQLite-only non-interactive creation.

## Constraints

- New commands MUST follow the flat structure — no `glow article create` style nesting.
- Content input MUST use `--content` / `--stdin` / `--diff` pattern — no positional content arguments.
- `--diff` MUST remain exclusive from `--content` and `--stdin`.

## Compliance

All existing commands comply. No remediation needed.

## Notes

- Aliases `show` and `cat` are accepted for `glow read`.
- `--sections` lists headings; `--section <name>` scopes to a section for read/update/append/delete/diff.
