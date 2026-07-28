# Architecture Decision Records

Architecture Decision Records (ADRs) preserve repository-wide architectural
decisions and their history.

## Identifier policy

- Identifiers are sequential repository-wide numbers formatted as `ADR-NNN`.
- An identifier is permanent once assigned and is never reused.
- Rejected, deprecated, and superseded records retain their identifiers.
- ADR filenames begin with the assigned identifier.
- A superseding ADR receives a new identifier and names every record it
  supersedes.

Historical references establish that ADR-001 through ADR-022 have already been
assigned. Missing historical files do not make their numbers available for
reuse. Assigned in-repo: ADR-023 (course hierarchy), ADR-024 (exam platform
domain). The next assigned identifier is ADR-025.

## Lifecycle

- **Proposed**: under review and not authoritative.
- **Accepted**: approved and authoritative for implementation.
- **Superseded**: replaced by a named later ADR and retained as history.
- **Deprecated**: discouraged without a complete replacement.

An ADR may move from Proposed to Accepted only after its technical evidence and
acceptance criteria are complete. Incompatible architectural changes require a
new ADR; accepted records are not rewritten to conceal architectural history.

