# GK Circle Documentation Folder Rules

Version: 1.0

Status: Mandatory Standard

Owner: Engineering

Last Updated: 2026-07-13

---

# Purpose

Define where documentation belongs so agents do not create duplicate or
misplaced documents.

This standard works with:

- `docs/standards/documentation-rules.md`
- `docs/standards/documentation-governance.md`

---

# Folder Placement Rule

Documentation must be placed in the folder owned by the feature, decision, or
evidence type.

Use these locations:

| Documentation Type | Location |
|---|---|
| Architecture decisions | `docs/architecture/ADR/` |
| Standards | `docs/standards/` |
| Active session context | `docs/active-context/` |
| Feature specifications | `docs/features/<feature>/` |
| Release or certification evidence | `docs/production-validation/` or the approved evidence folder for that workstream |
| Bug bash evidence | `docs/bug-bash/` |
| Testing governance | `docs/testing/` |

Do not create markdown files in the repository root.

---

# Duplicate Prevention Rule

Before creating a documentation file:

1. Search for an existing document with the same purpose.
2. Update the existing owner document when possible.
3. Create a new document only when it has a distinct governance, feature, or
   evidence purpose.
4. Link related documents instead of copying large sections.

---

# Evidence Folder Rule

Evidence must stay with the workstream that produced it.

For nested curriculum Phase C evidence, use:

```text
docs/production-validation/nested-course-curriculum/
```

For Course terminology standards alignment, use:

```text
docs/production-validation/nested-course-curriculum/course-terminology-standards-alignment/
```

---

# Current Documentation Rule

Current documentation must describe the active architecture.

Historical or superseded documents may retain old terminology only when they are
clearly classified as historical, archived, or superseded by a newer ADR or
standard.

Do not use superseded documents as implementation guidance.
