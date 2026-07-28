# EXAM-P5-T01 — Production audit

## Breaking Changes: NO

## Database migration status
- Additive only.
- `assessment_attempts.test_snapshot_id` nullable for apply safety; application create always sets it.
- Partial unique index enforces one `IN_PROGRESS` attempt per quiz+user.
- `attempt_answers.attempt_id` FK changed CASCADE → RESTRICT (ADR-024 §7).

## Runtime risk
- New routes only; existing live-quiz paths unchanged.
- Requires migrate up before attempt APIs succeed against DBs missing the new column/index.

## Rollback
- Use down migrations in reverse order after draining new attempt traffic (if any).
- No live-quiz data reinterpretation.
