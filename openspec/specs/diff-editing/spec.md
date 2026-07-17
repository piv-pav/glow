## ADDED Requirements

### Requirement: Apply SEARCH/REPLACE diff blocks
The system SHALL apply one or more SEARCH/REPLACE blocks to an article when `glow update --diff` is used. Blocks are read from STDIN. Each SEARCH string MUST match exactly once in the target content; zero or multiple matches SHALL be an error. All blocks SHALL be applied atomically — if any block fails, the article is left unchanged.

Format:
```
<<<<<<< SEARCH
exact existing text
=======
replacement text
>>>>>>> REPLACE
```

#### Scenario: Single block applied
- **WHEN** a single SEARCH/REPLACE block is piped to `glow update my-article --diff`
- **THEN** the matching text is replaced and "Applied 1 diff block(s) to article: my-article" is printed

#### Scenario: Multiple blocks applied
- **WHEN** two SEARCH/REPLACE blocks are piped
- **THEN** both replacements are applied and count reflects 2 blocks

#### Scenario: SEARCH not found
- **WHEN** the SEARCH text does not exist in the article
- **THEN** command exits with error; article is unchanged

#### Scenario: Ambiguous SEARCH (multiple matches)
- **WHEN** the SEARCH text matches more than once
- **THEN** command exits with error; article is unchanged

#### Scenario: Empty SEARCH prepends
- **WHEN** SEARCH block is empty
- **THEN** replacement text is prepended to the article

### Requirement: Scope diff to section
The system SHALL scope SEARCH/REPLACE matching to a named section when `--section` is combined with `--diff`.

#### Scenario: Diff scoped to section
- **WHEN** `glow update my-article --diff --section "Notes"` is used
- **THEN** SEARCH matches only within the "Notes" section content
- **THEN** "Applied N diff block(s) to section "Notes" in article: my-article" is printed

### Requirement: --diff is exclusive input mode
The system SHALL reject combining `--diff` with `--content` or `--stdin`.

#### Scenario: --diff with --content rejected
- **WHEN** `glow update my-article --diff --content "x"` is run
- **THEN** command exits with error "--diff cannot be combined with --content or --stdin"
