# EXAM-P2-T01 — Question Versioning and Answer Authority Foundation

Status: VERIFIED
- Started: 2026-07-27
- Verified: 2026-07-27

## Frozen Acceptance

- [x] Question revision model and immutable `question_revisions` storage (`ON DELETE RESTRICT`).
- [x] Answer authority fields per ADR-024 (`official_answer`, `authoritative_answer`, `answer_review_status`, `answer_revision_reason`, `answer_revision_source`) plus `lineage_id` / `revision_number`.
- [x] Question create/update records a revision; edit rewiring remains backward-compatible.
- [x] Question editor support: authority fields + read-only revision history.
- [x] `GET /v1/quizzes/:quiz_id/questions/:question_id/revisions` for editors.
- [x] Unit and controller tests; frontend editor tests for authority payload.
- [x] Evidence pack under `docs/features/exam-platform/evidence/exam-p2-t01/`.

## Standards loaded

AGENTS.md, CLAUDE.MD, docs/standards/index.md, ADR-024, Product/Engineering roadmaps, EXAM-P2 phase ledger.

## Architecture notes

| Concept | Implementation |
|---|---|
| Stable lineage | `questions.lineage_id` persists across edit rewires |
| Current head | Live `questions` row holds authority fields + `revision_number` |
| History | Append-only `question_revisions` (no update/delete API) |
| Live compatibility | `answers` remains operational key for live quiz; `authoritative_answer` defaults to same keys |
| Edit flow | New question row + `quiz_questions` rewire; revision_number increments within lineage |

## Migration summary

| Migration | Purpose |
|---|---|
| `20260727120000_question_revisions_and_answer_authority` | Authority columns on `questions`, `question_revisions` table, backfill revision 1 |

Run: `gk-circle migrate up`

## API changes (additive)

| Method | Path | Notes |
|---|---|---|
| GET | `/v1/quizzes/:quiz_id/questions/:question_id/revisions` | Editor revision history (`VerifyQuizEditAccess`) |
| POST/PUT | `/v1/quizzes/:quiz_id/questions[...]` | Optional authority fields on body |

Question analytics responses include: `lineage_id`, `revision_number`, `official_answer`, `authoritative_answer`, `answer_review_status`, `answer_revision_reason`, `answer_revision_source`.

## Checks (2026-07-27)

```text
go test ./models/ -run "Question|Answer" -count=1                    → PASS
go test ./controllers/api/v1/ -run "ListQuestionRevisions" -count=1 → PASS
go build ./...                                                        → PASS
npm test -- --run EditQuestion                                        → PASS (2)
eslint (EditQuestion + QuestionFormCard + test)                       → PASS
```

## Compatibility verification

- Existing quizzes backfilled: `lineage_id = id`, status `CONFIRMED`, revision 1 inserted.
- Edit flow still rewires `quiz_questions`; historical sessions keep old `question_id`.
- Live scoring unchanged (`answers` column).
- CSV import / collections / attempt engine untouched.

## Production impact assessment

| Item | Value |
|---|---|
| Breaking changes | NO |
| Migration required | YES |
| Downtime | None expected (additive DDL) |
| Risk | Low |

## Out of scope (per authorisation)

EXAM-P2-T02+, CSV import, collections, test builder, player, analytics, revision queue, PYQ filtering, attempt engine, scoring changes.

## Production source modified by EXAM-P2-T01: YES
