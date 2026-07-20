# Contributing

## Workflow

This project uses [OpenSpec](https://github.com/Fission-AI/OpenSpec) with the `spec-driven-with-decisions` schema.

Every behavioral change must be accompanied by:

- Updated or new delta specs under `openspec/changes/<name>/specs/`
- A decisions review under `openspec/changes/<name>/decisions/` (even if the outcome is "no decision changes")

**PRs that modify behavior without updated specs will not be accepted.**

## Getting the Schema

The `spec-driven-with-decisions` schema is included as a git submodule at `openspec/schemas/`.
If you cloned without `--recurse-submodules`, initialise it:

```bash
git submodule update --init
```

Verify:

```bash
openspec schema which spec-driven-with-decisions
# Source: project
```

## Starting a Change

```bash
openspec new change "<your-change-name>"
openspec status --change "<your-change-name>"
```

Then follow `/opsx:propose` or `/opsx:apply` in your AI coding assistant.
