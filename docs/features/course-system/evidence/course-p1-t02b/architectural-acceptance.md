# COURSE-P1-T02B Architectural Acceptance

- ADR: `ADR-023`
- ADR Version: 1
- Schema Version: 1
- Accepted: 2026-07-25T12:52:50.0300196+05:30
- Decision: ACCEPTED
- Task status: VERIFIED manually after technical verification

Technical verification passed before acceptance. The supporting record is
`technical-verification.md` and its machine-readable companion.

ADR-023 is now authoritative for the Course hierarchy. It resolves the former
CourseNode versus CourseSubject/CourseTopic persistence conflict while leaving
existing Course persistence and APIs unchanged.

This acceptance authorizes no CourseNode code or migration. T03 remains
NOT_STARTED until separately approved.
