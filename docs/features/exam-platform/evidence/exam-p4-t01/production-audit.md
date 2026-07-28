# EXAM-P4-T01 — Production Audit

## Risk summary

Additive schema and APIs only. Collections are editor-scoped under existing quiz permission middleware. No learner-facing or scoring changes.

## Data integrity

- STATIC members reference `questions` with `ON DELETE RESTRICT`.
- Collections cascade-delete with parent quiz.
- Member replace is transactional (delete + insert + commit).

## Security

- All collection routes require Kratos auth + quiz edit access.
- Resolve preview does not expose answer keys (question IDs only).

## Rollback

Run migration down: `gk-circle migrate down 1` (drops collection tables; no impact on questions/quizzes).

## Unverified at runtime

Browser/UI verification not applicable (backend-only task). Full-stack test builder verification deferred to EXAM-P4-T02.
