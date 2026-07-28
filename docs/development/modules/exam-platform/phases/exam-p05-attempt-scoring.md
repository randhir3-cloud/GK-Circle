# EXAM-P5 — Attempt and Scoring Engine

* **Status**: VERIFIED
* **Weight**: 15%
* **Started**: 2026-07-28
* **Completed tasks**: T01–T04 (2026-07-28)
* **Closed**: 2026-07-28 (formal approval)

## Objective

Server-authoritative self-paced attempt lifecycle with transactional scoring and negative marking. Ships **before** the full student player UI (P6).

## Planned tasks

| ID | Title | Status |
|---|---|---|
| EXAM-P5-T01 | Go models/repos for assessment_attempts / attempt_answers | IMPLEMENTED |
| EXAM-P5-T02 | Start / resume / autosave APIs | IMPLEMENTED |
| EXAM-P5-T03 | Idempotent submit + PCS scoring (negative marks) | IMPLEMENTED |
| EXAM-P5-T04 | Resolved question snapshot persistence | IMPLEMENTED |

## EXAM-P5-T01 acceptance (frozen)

<!-- TASK:EXAM-P5-T01:ACCEPTANCE:START -->
- [x] Go models/repos for existing `assessment_attempts` and `attempt_answers` tables (do not overload live `user_played_quizzes`).
- [x] Additive migration binds attempts to immutable `test_snapshots` (`test_snapshot_id` + RESTRICT).
- [x] Authenticated create attempt associates ownership from session identity (never client-supplied user id) and freezes `question_order` from the snapshot.
- [x] One-active-attempt policy: repeat create while `IN_PROGRESS` returns the existing attempt (idempotent).
- [x] Reject foreign/empty snapshots, exceeded `max_attempts`, and unauthorised callers.
- [x] Learner get/list returns attempt metadata + snapshot learner items with **no answer keys**.
- [x] Editor inspection of attempts is a separate authorised representation.
- [x] Attempt answer writes, autosave, submit, and scoring remain out of scope (T02/T03).
- [x] `go test ./...` and `go build ./...` pass; evidence under `docs/features/exam-platform/evidence/exam-p5-t01/`.
<!-- TASK:EXAM-P5-T01:ACCEPTANCE:END -->

## Frozen product rules for this phase

- Attempts consume immutable test snapshots (ADR-024 §3/§4); never live collections or mutable bank rows for exam content.
- Live-session played-quiz tables remain separate; self-paced uses `assessment_attempts`.
- Scoring and answer persistence ship in later P5 tasks; T01 is foundation + create/get only.

## EXAM-P5-T02 acceptance (frozen)

<!-- TASK:EXAM-P5-T02:ACCEPTANCE:START -->
- [x] Authenticated autosave upserts one answer row per `(attempt_id, question_id)` for an owned `IN_PROGRESS` attempt.
- [x] Question must exist in the attempt’s immutable test snapshot; option keys must match frozen snapshot options; single-answer cardinality enforced.
- [x] Explicit clear (null/empty selected options) is supported; `score` / `is_correct` are never written by autosave.
- [x] Resume API returns learner-safe snapshot, question order, saved answers, status/timing, and progress counts — with **no answer keys**.
- [x] Terminal attempts (`SUBMITTED` / `AUTO_SUBMITTED` / `ABANDONED`) reject answer mutation.
- [x] Duplicate/concurrent upserts do not create multiple rows (unique constraint + ON CONFLICT).
- [x] Ownership from session only; foreign quiz/attempt/question rejected; unauthenticated rejected.
- [x] Submit/scoring remain out of scope (EXAM-P5-T03).
- [x] `go test ./...` and `go build ./...` pass; evidence under `docs/features/exam-platform/evidence/exam-p5-t02/`.
<!-- TASK:EXAM-P5-T02:ACCEPTANCE:END -->

## Frozen product rules for this phase

- Attempts consume immutable test snapshots (ADR-024 §3/§4); never live collections or mutable bank rows for exam content.
- Live-session played-quiz tables remain separate; self-paced uses `assessment_attempts`.
- T01: foundation + create/get. T02: autosave + resume. T03: submit + scoring.

## EXAM-P5-T03 acceptance (frozen)

<!-- TASK:EXAM-P5-T03:ACCEPTANCE:START -->
- [x] Authenticated owner can submit an owned `IN_PROGRESS` attempt; transition to `SUBMITTED` with `submitted_at`.
- [x] Scoring uses only frozen snapshot items (authoritative key if present, else operational `answers`); never live bank/collections; never `official_answer` alone.
- [x] Single-answer: correct → +points; incorrect → −`negative_marks_per_question`; unanswered → 0. Survey/unscored types contribute 0 and leave `is_correct` null.
- [x] Aggregate `total_score` / `max_score` persisted; floor total at 0; per-answer `score`/`is_correct` written in the same transaction.
- [x] Repeated submit is idempotent (returns stored result, no rescore). Concurrent submit cannot score twice (row lock).
- [x] Client-supplied score/correctness/status rejected; non-owner/unauthenticated rejected; learner result omits answer keys.
- [x] Auto-submit, player UI, and EXAM-P5-T04 remain out of scope.
- [x] `go test ./...` and `go build ./...` pass; evidence under `docs/features/exam-platform/evidence/exam-p5-t03/`.
<!-- TASK:EXAM-P5-T03:ACCEPTANCE:END -->

## EXAM-P5-T04 acceptance (frozen)

<!-- TASK:EXAM-P5-T04:ACCEPTANCE:START -->
- [x] At attempt create, copy resolved snapshot items into attempt-linked rows (IDs, revision, order, stem/options/keys, points) without replacing `test_snapshots`.
- [x] Freeze scoring config on the attempt (`negative_marks_per_question`, `expected_max_score`) at create; later quiz edits must not change in-flight or historical scoring.
- [x] Autosave validation, resume, and submit/scoring read attempt-linked snapshot items (not live bank/collections).
- [x] Submitted attempts remain immutable; attempt-linked snapshot rows are insert-once (no update API).
- [x] Learner payloads remain answer-key-safe; frozen scoring metadata may be exposed without keys.
- [x] Phase 6 player / analytics / auto-submit remain out of scope.
- [x] `go test ./...` and `go build ./...` pass; evidence under `docs/features/exam-platform/evidence/exam-p5-t04/`.
<!-- TASK:EXAM-P5-T04:ACCEPTANCE:END -->

