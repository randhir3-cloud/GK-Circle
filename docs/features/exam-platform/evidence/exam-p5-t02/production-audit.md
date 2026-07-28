# EXAM-P5-T02 — Production audit

## Breaking Changes: NO

## Database migration status
- No new migration for T02.
- Relies on existing `attempt_answers_unique_per_question` and T01 snapshot binding.

## Runtime risk
- Additive routes only.
- Answer writes limited to owned `IN_PROGRESS` attempts.

## Rollback
- Revert route/handlers; existing answer rows remain valid data.
