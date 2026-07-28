# EXAM-P2-T01 Production Audit

Audit date: 2026-07-27

## Deployment prerequisites

- Apply migration `20260727120000_question_revisions_and_answer_authority` via `gk-circle migrate up`.
- Existing question rows are backfilled with lineage, CONFIRMED status, and revision 1.

## Behaviour changes

| Change | Impact |
|---|---|
| New `questions` authority columns | Present on create/list/get for editors |
| `question_revisions` append-only history | Created on question create and edit |
| `GET .../questions/:id/revisions` | Editor-only (existing quiz edit middleware) |
| EditQuestion UI authority panel | Surfaces status, official/authoritative keys, history |

## Non-changes (compatibility)

- Live quiz scoring still uses `answers`.
- Historical `user_quiz_responses` / sessions still reference prior question IDs after edit rewire.
- No scoring algorithm changes.

## Rollback

1. Deploy previous application binary/assets.
2. Run migration down only if no dependent later data requires revisions (destructive to revision history).

## Security note

Revision listing requires quiz edit access. Unauthenticated answer-key endpoints remain an EXAM-P2-T03 concern (not in T01 scope).
