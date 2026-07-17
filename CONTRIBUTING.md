# Contributing

## Workflow

This project uses [OpenSpec](https://github.com/Fission-AI/OpenSpec) with the `spec-driven-with-decisions` schema.

Every behavioral change must be accompanied by:

- Updated or new delta specs under `openspec/changes/<name>/specs/`
- A decisions review under `openspec/changes/<name>/decisions/` (even if the outcome is "no decision changes")

**PRs that modify behavior without updated specs will not be accepted.**

## Getting the Schema

Install `spec-driven-with-decisions` globally:

```bash
mkdir -p ~/.local/share/openspec/schemas
cp -r openspec/schemas/spec-driven-with-decisions ~/.local/share/openspec/schemas/
```

Or clone directly from [codeberg.org/pivpav/openspec-schemas](https://codeberg.org/pivpav/openspec-schemas).

Verify:

```bash
openspec schema which spec-driven-with-decisions
```

## Starting a Change

```bash
openspec new change "<your-change-name>"
openspec status --change "<your-change-name>"
```

Then follow `/opsx:propose` or `/opsx:apply` in your AI coding assistant.
