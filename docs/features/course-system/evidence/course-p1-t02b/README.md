# COURSE-P1-T02B Canonical Evidence

Task: `COURSE-P1-T02B — Course hierarchy architecture decision`

Status: VERIFIED

Accepted: 2026-07-25T12:52:50.0300196+05:30

Canonical ADR:
`docs/architecture/ADR/ADR-023-canonical-course-hierarchy-model.md`

This directory is the authoritative evidence store for the hierarchy
architecture resolution. Module status and decision files contain summaries and
links only.

## Evidence index

- `technical-verification.md`: human-readable verification and acceptance record.
- `technical-verification.json`: machine-readable command and result record.
- `architectural-acceptance.md`: separate governance acceptance record.
- `commands/`: sanitized command output.
- `hashes/`: recursive ledger hashes proving read-only check and no-op sync.

## Safety

- CourseNode code created: No
- CourseNode migration created or run: No
- Production accessed: No
- NUC accessed: No
- Breaking change introduced: No

## Conclusion

Technical verification passed before the governance action. ADR-023 was then
accepted and T02B was manually promoted to VERIFIED. T03 remains NOT_STARTED
and requires separate explicit approval.
