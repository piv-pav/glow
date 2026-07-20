# CLI Design Decision

Version: 2  
Updated: 2026-07-21

## Agreement

Glow is a single-binary CLI. Commands are flat (no deep subcommand nesting). All content input uses `--content` (inline) or `--stdin` (pipe) — these are mutually exclusive. `--diff` is a third exclusive input mode for SEARCH/REPLACE editing. The `-w <name>` global flag targets a named wiki; without it, `default` is used. Wiki management commands are prefixed `wiki-` (`wiki-create`, `wiki-delete`, `wiki-list`).

`glow wiki-create <name>` is the single command for wiki creation. It requires a positional `<name>` argument and exactly one of two mutually exclusive mode flags:
- `--interactive` / `-i` — interactive mode; prompts for backend and connection details
- `--backend <b>` / `-b <b>` — non-interactive mode; creates wiki with the specified backend

Omitting both flags is an error. `glow init` does not exist.

## Rationale

**Flat command structure:** Keeps the CLI discoverable and scriptable. Commands like `glow create`, `glow update`, `glow search` are direct and composable in shell pipelines.

**--content / --stdin / --diff exclusivity:** Three clearly separated input modes prevent ambiguous combinations. `--diff` is explicitly incompatible with content flags — atomic SEARCH/REPLACE semantics require the full article to be in memory before applying blocks.

**-w global flag over per-command flag:** Consistent wiki targeting across all commands without repeating the flag definition. Mirrors how tools like `kubectl -n` work.

**wiki- prefix for management commands:** Separates lifecycle commands (wiki-create, wiki-delete, wiki-list) from article operation commands (create, read, update, etc.).

**Single `wiki-create` command, explicit mode flags:** Glow is AI-first. Implicit interactivity (prompting when flags are absent) breaks AI workflows by blocking on stdin. Interactivity is opt-in via `-i`. A human who wants prompts passes `-i`; an AI agent always passes `-b`. The two modes are clearly separated — no ambiguous fallback behaviour.

**Name always required:** AI agents always know the wiki name. Requiring it explicitly makes every invocation self-documenting and prevents silent defaulting.

**`-b` and `-i` shorthands:** Consistent shorthand flags reduce typing in interactive shell use. `-b` for `--backend`, `-i` for `--interactive`.

## Constraints

- New commands MUST follow the flat structure — no `glow article create` style nesting.
- Content input MUST use `--content` / `--stdin` / `--diff` pattern — no positional content arguments.
- `--diff` MUST remain exclusive from `--content` and `--stdin`.
- No command MAY prompt interactively without an explicit `-i` / `--interactive` flag. AI-safe by default.

## Compliance

`glow init` has been removed as part of this change. All wiki creation now goes through `wiki-create`. No other non-compliance exists.

## Notes

- Aliases `show` and `cat` are accepted for `glow read`.
- `--sections` lists headings; `--section <name>` scopes to a section for read/update/append/delete/diff.
