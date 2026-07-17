## ADDED Requirements

### Requirement: Read section by heading
The system SHALL return only the content of a named section when `--section` is passed to `glow read`. If the heading is not found, an error SHALL be returned. The heading line itself SHALL NOT be included in the output.

#### Scenario: Read existing section
- **WHEN** `glow read my-article --section "Usage"` is run
- **THEN** only the content under the "Usage" heading is printed (heading line excluded)

#### Scenario: Read missing section
- **WHEN** `glow read my-article --section "Nonexistent"` is run
- **THEN** command exits with error "section not found: Nonexistent"

### Requirement: Update section by heading
The system SHALL replace only the named section's body when `--section` is passed to `glow update`. Other sections SHALL remain unchanged.

#### Scenario: Update existing section
- **WHEN** `glow update my-article --section "Notes" --content "new notes"` is run
- **THEN** the "Notes" section content is replaced; other sections are unchanged

#### Scenario: Update missing section
- **WHEN** `glow update my-article --section "Ghost" --content "x"` is run
- **THEN** command exits with error

### Requirement: Append to section by heading
The system SHALL append content at the end of the named section when `--section` is passed to `glow append`.

#### Scenario: Append to existing section
- **WHEN** `glow append my-article --section "Notes" --content "- new item"` is run
- **THEN** "- new item" is appended inside the "Notes" section

### Requirement: Delete section by heading
The system SHALL remove the named section (heading + body) when `--section` is passed to `glow delete`.

#### Scenario: Delete existing section
- **WHEN** `glow delete my-article --section "Old Section"` is run
- **THEN** that section is removed; article otherwise unchanged

#### Scenario: Delete missing section
- **WHEN** `glow delete my-article --section "Ghost"` is run
- **THEN** command exits with error

### Requirement: List section headings
The system SHALL list all section headings in an article when `--sections` is passed to `glow read`.

#### Scenario: List sections
- **WHEN** `glow read my-article --sections` is run
- **THEN** all headings are printed with their level (e.g., `## Notes`)
