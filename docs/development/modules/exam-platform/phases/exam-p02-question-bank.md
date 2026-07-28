# EXAM-P2 — Question Bank, Versioning and Security

* **Status**: IN_PROGRESS
* **Weight**: 12%
* **Started**: 2026-07-27

## Objective

Versioned MCQ bank with PYQ metadata fields and secured answer-key / review endpoints.

## Tasks

| ID | Title | Status | Dependencies | Evidence |
|---|---|---|---|---|
| EXAM-P2-T01 | Question revision model + answer authority fields | VERIFIED | EXAM-P1-T01 | `docs/features/exam-platform/evidence/exam-p2-t01/` |
| EXAM-P2-T02 | Admin MCQ editor (Nuxt) | VERIFIED | EXAM-P2-T01 | `docs/features/exam-platform/evidence/exam-p2-t02/` |
| EXAM-P2-T03 | Secure answer-key and review endpoints | VERIFIED | EXAM-P1-T01 | `docs/features/exam-platform/evidence/exam-p2-t03/` |
| EXAM-P2-T04 | Answer-key protection tests (pre-release leak proofs) | VERIFIED | EXAM-P2-T03 | `docs/features/exam-platform/evidence/exam-p2-t04/` |

## EXAM-P2-T01 acceptance (frozen)

<!-- TASK:EXAM-P2-T01:ACCEPTANCE:START -->
- [x] Question revision model and immutable `question_revisions` storage.
- [x] Answer authority fields per ADR-024.
- [x] Question editor support (authority fields + revision history list).
- [x] Backward-compatible with existing quizzes and edit rewiring.
- [x] Evidence: `docs/features/exam-platform/evidence/exam-p2-t01/`
<!-- TASK:EXAM-P2-T01:ACCEPTANCE:END -->

## EXAM-P2-T02 acceptance (frozen)

<!-- TASK:EXAM-P2-T02:ACCEPTANCE:START -->
- [x] Shared `McqQuestionEditor` for MCQ create/edit with full ADR-024 answer authority fields.
- [x] `QuestionRevisionHistory` via T01 revisions API.
- [x] Quiz inline (`QuestionFormCard`) and full-page (`EditQuestion`) editors use shared component.
- [x] Independent official and authoritative answer keys in the editor.
- [x] `quiz_questions` composable + `question_authority` utils.
- [x] Vitest coverage for utils, composable, and editor components.
- [x] Evidence: `docs/features/exam-platform/evidence/exam-p2-t02/`
<!-- TASK:EXAM-P2-T02:ACCEPTANCE:END -->

## EXAM-P2-T03 acceptance (frozen)

<!-- TASK:EXAM-P2-T03:ACCEPTANCE:START -->
- [x] `GET /v1/analytics_board/user`, `GET /v1/final_score/user`, and `GET /v1/user_played_quizes/:user_played_quiz_id` require authentication.
- [x] Unauthenticated callers receive HTTP 401 (no answer-key payload).
- [x] Authenticated participants and live session hosts retain review access to their played-quiz data.
- [x] Quiz editors (creator, shared read/write/share, public-quiz admin) retain per-participant review preview.
- [x] Existing admin aggregate endpoints (`/analytics_board/admin`, `/final_score/admin`) unchanged.
- [x] Unit/middleware tests prove unauthenticated rejection and participant allow paths.
- [x] Evidence: `docs/features/exam-platform/evidence/exam-p2-t03/`
<!-- TASK:EXAM-P2-T03:ACCEPTANCE:END -->

## EXAM-P2-T04 acceptance (frozen)

<!-- TASK:EXAM-P2-T04:ACCEPTANCE:START -->
- [x] Security regression suite proves unauthenticated review endpoints return 401 with no answer-key fields in the body.
- [x] Security regression suite proves authenticated unauthorised callers receive 403 with no answer-key fields.
- [x] Security regression suite proves authorised participant, host, and editor review paths still return answer-key fields when allowed.
- [x] Pre-release live question WebSocket delivery payload excludes answer keys (`BuildLiveQuestionDeliveryPayload` + tests).
- [x] Nested JSON leak detection helper (`security.AssertNoSensitiveAnswerKeyFields`) with explicit field deny-list.
- [x] Route inventory documents protected review, pre-release, and editor answer-key surfaces.
- [x] `go test ./...` and `go build ./...` pass.
- [x] Evidence: `docs/features/exam-platform/evidence/exam-p2-t04/`
<!-- TASK:EXAM-P2-T04:ACCEPTANCE:END -->

## Frozen product rules for this phase

- Question versioning is mandatory (not deferred to results).
- Answer authority: `officialAnswer`, `authoritativeAnswer`, `answerReviewStatus`, `answerRevisionReason`, `answerRevisionSource`.
- Scorer (later) uses snapshot keys only.
- EXAM-P2-T03 must prove unauthenticated review is rejected; editors retain preview; live flows remain authenticated.
