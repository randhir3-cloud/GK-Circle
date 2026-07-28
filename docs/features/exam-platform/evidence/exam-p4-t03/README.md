# EXAM-P4-T03 — Inline Add New Question → Bank + Link

Status: IMPLEMENTED
- Started: 2026-07-27
- Implemented: 2026-07-27

## Standards loaded

AGENTS.md, CLAUDE.md, docs/standards/index.md (+ frontend/testing/security), ADR-024 §8, Product/Engineering roadmaps, EXAM-P4 phase ledger, EXAM-P4-T01/T02 evidence, existing Question Bank create path.

## Task understanding

EXAM-P4-T03 adds **inline Add New Question** inside the Visual Test Builder. Create must reuse the shared Question Bank path (`POST /v1/quizzes/:quiz_id/questions` → `AppendQuestionsToQuiz`) so lineage, revision 1, answer authority, and revision history are preserved. Optionally append to a selected STATIC collection via T01 `PUT .../members`. No parallel store, no T04 snapshots.

## Frozen acceptance

<!-- TASK:EXAM-P4-T03:ACCEPTANCE:START -->
- [x] Visual Test Builder supports inline **Add New Question** without leaving the builder.
- [x] Create uses the existing Question Bank path (`POST /v1/quizzes/:quiz_id/questions` → `AppendQuestionsToQuiz`) so lineage, revision 1, answer authority, and revision history are recorded.
- [x] Newly created questions are linked to the current quiz via the shared bank create path (no parallel store).
- [x] When a STATIC collection is selected, editors can optionally append the new question to that collection’s ordered membership via `PUT .../members`.
- [x] DYNAMIC collections: question enters the quiz bank; membership is not forced (filter resolution remains server-owned).
- [x] Reuses shared `QuestionFormCard` / `McqQuestionEditor` and `useQuizQuestionsApi.createQuestion`.
- [x] Parent quiz question list refreshes after create.
- [x] Vitest coverage for inline create + STATIC link; lint/build/tests pass.
- [x] Evidence: `docs/features/exam-platform/evidence/exam-p4-t03/`
<!-- TASK:EXAM-P4-T03:ACCEPTANCE:END -->

## Architecture notes

```text
VisualTestBuilder
  └── Add New Question
        └── QuestionFormCard / McqQuestionEditor
              └── useQuizQuestionsApi.createQuestion
                    └── POST /v1/quizzes/:quiz_id/questions
                          └── QuizService.AppendQuestionsToQuiz
                                └── registerQuestions + RecordRevision + quiz_questions

If STATIC selected + “Also link…” checked:
  └── PUT /v1/quizzes/:quiz_id/collections/:id/members
        └── append new question_id to ordered membership

Parent page @question-created → refresh() quiz question list
```

## Migration summary

None.

## API changes

None. Reuses existing create-question and collection-members APIs.

## Checks (2026-07-27)

```text
npm test -- --run VisualTestBuilder.test.js          → PASS (6)
eslint (changed T03 files)                           → PASS
npm run build                                        → PASS
```

## Compatibility verification

- Existing page-level “Add Question” form unchanged.
- Import wizard / revision history unchanged.
- No backend schema or route changes.
- Collection CRUD from T02 unchanged aside from inline create affordance.

## Production impact assessment

| Item | Value |
|---|---|
| Breaking changes | **NO** |
| Migration required | **NO** |
| Downtime | None |
| Risk | Low |

## Deferred work

| Item | Task |
|---|---|
| Attempt/test snapshot contract hooks | EXAM-P4-T04 |
| Attempt engine / player / analytics | Later phases |
| AI-generated questions / bulk inline import | Out of programme |

## Production source modified by EXAM-P4-T03: YES
