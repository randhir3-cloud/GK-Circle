# EXAM-P5-T03 — Production audit

## Breaking Changes: NO

## Database migration status
- No new migration.
- Uses existing `assessment_attempts` score/status columns and `attempt_answers.score` / `is_correct`.

## Runtime risk
- Additive submit/result routes.
- Scoring reads quiz `negative_marks_per_question` at first submit only; repeats return stored totals.

## Rollback
- Revert handlers; submitted rows remain valid historical data.
