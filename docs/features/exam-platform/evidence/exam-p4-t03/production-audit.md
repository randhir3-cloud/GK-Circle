# EXAM-P4-T03 — Production Audit

## Risk summary

Frontend-only extension of the Visual Test Builder. Question persistence continues to use the authenticated Question Bank create API. Optional STATIC membership append uses the existing T01 members API.

## Security / integrity

- Create path remains server-validated (`VerifyQuizEditAccess`).
- Lineage, revision, and answer authority remain server-owned via `AppendQuestionsToQuiz`.
- No direct DB writes from the client.
- Answer keys are authored through the shared MCQ editor, not via collection resolve.

## Rollback

Hide or remove the inline Add New Question controls in `VisualTestBuilder`. Questions already created remain in the bank; STATIC membership links remain editable via T02 membership UI.

## Unverified at runtime

IronBee browser MCP was not available in this session. Unit/component tests and production build verify wiring; live browser smoke is deferred to manual review.
