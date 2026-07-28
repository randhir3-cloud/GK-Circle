# EXAM-P5-T04 — Production audit

## Breaking Changes: NO

Additive schema and behaviour. Existing routes unchanged. New attempts freeze scoring config and copy snapshot items; learner envelopes may include `negative_marks_per_question` / `expected_max_score`.

## Database migration status

- **Required**: `20260728120000_create_assessment_attempt_snapshot_items`
- Adds columns on `assessment_attempts`
- Creates `assessment_attempt_snapshot_items` with RESTRICT FKs to attempts and `test_snapshot_items`
- Down script drops table then columns

## Runtime risk

- Create is transactional (attempt + items).
- Submit/autosave no longer re-read quiz negative marks for scoring/validation of content.
- Pre-T04 attempts without attempt items fall back to shared snapshot for learner projection; submit requires attempt items (empty → empty-snapshot error). Ops may need backfill if any IN_PROGRESS attempts exist across cutover.

## Rollback

1. Stop serving new create traffic if needed.
2. Revert application revision.
3. Run migration down only if no dependent attempt-item rows must be retained (destructive to T04 data).
