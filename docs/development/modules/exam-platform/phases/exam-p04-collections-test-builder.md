# EXAM-P4 — Collections and Visual Test Builder

* **Status**: IN_PROGRESS
* **Weight**: 12%
* **Started**: 2026-07-27

## Objective

STATIC and DYNAMIC Question Collections; visual Test Builder; inline **+ Add New Question** that persists into the Question Bank and links to the current test.

## Planned tasks

| ID | Title | Status |
|---|---|---|
| EXAM-P4-T01 | Collection entities (STATIC + DYNAMIC filters) | VERIFIED |
| EXAM-P4-T02 | Visual Test Builder UI | VERIFIED |
| EXAM-P4-T03 | Inline Add New Question → bank + link to test | VERIFIED |
| EXAM-P4-T04 | Attempt/test snapshot contract hooks | IMPLEMENTED |

## EXAM-P4-T01 acceptance (frozen)

<!-- TASK:EXAM-P4-T01:ACCEPTANCE:START -->
- [x] Additive migration for `question_collections` and `question_collection_members`.
- [x] `STATIC` and `DYNAMIC` kinds with DB constraints (STATIC has no filter; DYNAMIC requires `filter_json`).
- [x] Quiz-scoped CRUD APIs with editor authentication.
- [x] STATIC ordered membership replace validates questions are linked to the quiz.
- [x] DYNAMIC filter schema validation and resolve preview endpoint.
- [x] Metadata filters return `METADATA_PENDING` until question bank taxonomy ships (EXAM-P10).
- [x] Empty dynamic filter resolves to all quiz-linked questions.
- [x] Model + controller tests; `go test ./...` and `go build ./...` pass.
- [x] Evidence: `docs/features/exam-platform/evidence/exam-p4-t01/`
<!-- TASK:EXAM-P4-T01:ACCEPTANCE:END -->

## EXAM-P4-T02 acceptance (frozen)

<!-- TASK:EXAM-P4-T02:ACCEPTANCE:START -->
- [x] Nuxt Visual Test Builder on the quiz manage page consumes T01 collection APIs only (no parallel client storage).
- [x] Editors can list, create, update, and delete STATIC and DYNAMIC collections for the current quiz.
- [x] STATIC builder manages ordered membership from existing quiz-bank questions and persists via `PUT .../members`.
- [x] DYNAMIC builder edits filter criteria (subject, topic, year, difficulty, PYQ status) and persists via `PATCH`.
- [x] Resolve preview clearly distinguishes ordered STATIC membership from DYNAMIC resolved question IDs, including `METADATA_PENDING` and empty-filter messaging.
- [x] Answer keys are not re-fetched or displayed through collection resolve (question IDs / titles from existing quiz list only).
- [x] Composable + component Vitest coverage; lint and frontend tests pass.
- [x] Evidence: `docs/features/exam-platform/evidence/exam-p4-t02/`
<!-- TASK:EXAM-P4-T02:ACCEPTANCE:END -->

## EXAM-P4-T03 acceptance (frozen)

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

## EXAM-P4-T04 acceptance (frozen)

<!-- TASK:EXAM-P4-T04:ACCEPTANCE:START -->
- [x] Additive `test_snapshots` + `test_snapshot_items` schema freezing composed question IDs, revision identifiers, order, and answer keys.
- [x] Editor API creates a quiz-scoped immutable snapshot by resolving STATIC/DYNAMIC collections once (no later re-evaluation).
- [x] Snapshot creation is transactional; empty or unresolved (`METADATA_PENDING`) collections are rejected explicitly.
- [x] Duplicate question IDs across composed collections are rejected; ordering is deterministic (collection position, then member/resolve order).
- [x] Snapshot remains unchanged after later collection membership/filter edits and after live question revision/bank mutations.
- [x] Editor read returns frozen payload (authorised); learner-safe read omits answer keys.
- [x] No attempt start/submit/scoring/player APIs (those remain EXAM-P5+).
- [x] Model/controller tests cover immutability, duplicates, empty/unresolved, and answer-key safety; `go test ./...` and `go build ./...` pass.
- [x] Evidence: `docs/features/exam-platform/evidence/exam-p4-t04/`
<!-- TASK:EXAM-P4-T04:ACCEPTANCE:END -->

## Frozen product rules for this phase

- Question Bank remains the single store (ADR-024 §8).
- Collections support STATIC fixed membership and DYNAMIC filter criteria (ADR-024 §3).
- Attempt snapshots are contract hooks only in T04; full persistence in P5.
- Visual builder and inline question create are separate tasks (T02, T03).
