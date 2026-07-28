# EXAM-P5-T02 — Start / Resume / Autosave APIs

Status: IMPLEMENTED  
- Started: 2026-07-28  
- Implemented: 2026-07-28

## Standards loaded

AGENTS.md, CLAUDE.md, docs/standards/index.md (+ backend/testing/security), ADR-024, Product/Engineering roadmaps, EXAM-P5 ledger, EXAM-P5-T01 evidence, EXAM-P4-T04 snapshot contract, attempt models and answer-leak conventions.

## Task understanding

EXAM-P5-T02 adds **answer autosave** and **attempt resume** for owned `IN_PROGRESS` self-paced attempts. Answers bind to the attempt’s immutable snapshot item (not live bank/collections). Submit, auto-submit, and scoring remain EXAM-P5-T03.

## Frozen acceptance

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

## Autosave architecture

```text
Authenticated learner
        ↓
Owned IN_PROGRESS attempt (FOR UPDATE)
        ↓
Snapshot item validation (options + cardinality)
        ↓
Transactional upsert attempt_answers
        ↓
Touch attempt.updated_at
        ↓
Learner-safe answer response (no score/keys)
```

## Resume contract

`GET .../attempts/:attempt_id/resume` returns:

- attempt status, timing, question order
- learner-safe snapshot items (no keys)
- saved `selected_options` / `is_marked_review` / timestamps
- progress: total / answered / marked / unanswered

Must not expose: operational keys, official/authoritative answers, correctness, scores, review-status authority fields.

## State and validation rules

| Rule | Behaviour |
|---|---|
| Mutable status | `IN_PROGRESS` only |
| Terminal | reject autosave |
| Question scope | must be in bound snapshot |
| Single answer | exactly one option when set |
| Survey | 1..N distinct valid options |
| Clear | `clear=true` or `selected_options: []` |
| Forbidden body | `user_id`, `score`, `is_correct` |

## Concurrency and idempotency

- Unique `(attempt_id, question_id)` prevents duplicate rows.
- Insert race → detect unique violation → update existing row.
- Repeated identical saves update `updated_at` and converge to same selection.
- Last write wins (no optimistic version token in T02).

## Security model

| Concern | Behaviour |
|---|---|
| Auth | Kratos session required |
| Owner | session user only |
| Foreign attempt | 404 |
| Foreign question | 400 |
| Terminal mutate | 400 |
| Answer keys | stripped on resume |

## Migration summary

| Migration | Purpose |
|---|---|
| None | Reuses `attempt_answers` unique constraint from existing schema |

## API changes (additive)

| Method | Path | Notes |
|---|---|---|
| PUT | `/v1/quizzes/:quiz_id/attempts/:attempt_id/answers/:question_id` | Autosave upsert |
| GET | `/v1/quizzes/:quiz_id/attempts/:attempt_id/resume` | Resume payload |

## Checks (2026-07-28)

```text
go test (scoped AttemptAnswer|AssessmentAttempt|Autosave|Resume) → PASS
go test ./... -count=1 → PASS
go build ./... → PASS
```

No frontend changes.

## Answer-leak verification

Resume and autosave responses omit `official_answer`, `authoritative_answer`, `answer_review_status`, `score`, `is_correct` (asserted in tests).

## Compatibility verification

- T01 create/idempotent/one-active unchanged.
- Snapshot immutability unchanged.
- Live `user_played_quizzes` untouched.
- Editor attempt inspection route unchanged.

## Production impact assessment

| Item | Value |
|---|---|
| Breaking changes | **NO** |
| Migration required | **NO** (for T02) |
| Downtime | None |
| Risk | Low |

## Deferred work

| Item | Task |
|---|---|
| Idempotent submit + PCS scoring | EXAM-P5-T03 |
| Attempt-linked snapshot extras | EXAM-P5-T04 |
| Student player | EXAM-P6 |

## Production source modified by EXAM-P5-T02: YES

## Stop condition

EXAM-P5-T03 **not started**. Awaiting explicit manual review.
