# EXAM-P5-T01 — Self-paced Attempt Foundation (models/repos + snapshot-bound create/get)

Status: IMPLEMENTED  
- Started: 2026-07-28  
- Implemented: 2026-07-28

## Standards loaded

AGENTS.md, CLAUDE.md, docs/standards/index.md (+ backend/testing/security), ADR-024 §2–§4/§7, Product/Engineering roadmaps, EXAM-P5 phase ledger, EXAM-P4-T04 snapshot contract evidence, Kratos auth + QuizPermission conventions.

## Task understanding

EXAM-P5-T01 introduces the **self-paced attempt foundation**: Go models/repos for `assessment_attempts` / `attempt_answers`, additive binding to immutable `test_snapshots`, and authenticated create/get/list APIs. Content for an attempt comes only from the bound snapshot (question order + learner-safe items). This is **not** the full Attempt Engine — autosave/resume (T02), submit/scoring (T03), and attempt-linked snapshot wiring (T04) remain deferred.

## Frozen acceptance

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

## Attempt architecture

```text
Immutable Test Snapshot (EXAM-P4-T04)
        ↓ bind FK + copy question_order
Assessment Attempt (IN_PROGRESS)
        ↓ later EXAM-P5-T02/T03
Answer writes → Submit → Score
```

Live `user_played_quizzes` / socket session tables are **not** reused for self-paced semantics.

## Lifecycle and state transitions (T01)

| State | Introduced in T01 | Notes |
|---|---|---|
| `IN_PROGRESS` | Yes (create) | Only create transition shipped |
| `SUBMITTED` | Schema only | Transition deferred to T03 |
| `AUTO_SUBMITTED` | Schema only | Deferred |
| `ABANDONED` | Schema only | Deferred; does not consume `max_attempts` |

Create eligibility: quiz `assessment_mode=SELF_PACED`; learners require `status=PUBLISHED`; editors with write/share may create on draft for preview.

One-active policy: unique partial index on `(quiz_id, user_id) WHERE status='IN_PROGRESS'`; repeat create returns existing attempt (HTTP 200).

## Snapshot integration

- Required `test_snapshot_id` FK → `test_snapshots` ON DELETE RESTRICT.
- `question_order` jsonb copied from snapshot item positions at create; never re-resolved from collections/bank.
- Learner payloads embed `TestSnapshotLearnerView` (no keys).
- Editor payloads embed full frozen snapshot (keys included).

## Security and entitlement model

| Concern | Behaviour |
|---|---|
| Auth | `KratosAuthenticated` on all attempt routes |
| Owner | Taken from session `ContextUser.ID` only |
| Learner entitlement | Published SELF_PACED (any authenticated user) |
| Editor preview | Creator or write/share permission |
| Foreign attempt ID | Owner get returns 404 |
| Answer keys | Stripped from learner create/get/list |
| Editor inspection | Separate `.../editor` route + `QuizPermission` + `VerifyQuizEditAccess` |

## Migration summary

| Migration | Purpose |
|---|---|
| `20260728100000_alter_assessment_attempts_add_snapshot` | `test_snapshot_id` FK, indexes, one-active unique index |
| `20260728100100_alter_attempt_answers_restrict` | ADR-024 §7: attempt FK ON DELETE RESTRICT |

Run: `gk-circle migrate up`

## API changes (additive)

| Method | Path | Notes |
|---|---|---|
| POST | `/v1/quizzes/:quiz_id/attempts` | Create or return active; body `{snapshot_id}` |
| GET | `/v1/quizzes/:quiz_id/attempts` | List caller's attempts |
| GET | `/v1/quizzes/:quiz_id/attempts/:attempt_id` | Owner learner view |
| GET | `/v1/quizzes/:quiz_id/attempts/:attempt_id/editor` | Editor view with keys |

## Checks (2026-07-28)

```text
go test ./models/ ./services/ ./controllers/api/v1/ -run "AssessmentAttempt|AttemptAnswer" -count=1 → PASS
go test ./... -count=1 → PASS
go build ./... → PASS
```

No frontend changes → frontend lint/test/build not required.

## Answer-leak verification

- Learner create response omits `answers` / `official_answer` / `authoritative_answer` / `answer_review_status` (controller + service tests).
- Editor get includes frozen keys by design.

## Compatibility verification

- Live quiz / `user_played_quizzes` untouched.
- `test_snapshots` contract unchanged except consumption.
- Historical LIVE assessment_mode quizzes reject self-paced attempt create.

## Production impact assessment

| Item | Value |
|---|---|
| Breaking changes | **NO** |
| Migration required | **YES** (additive) |
| Downtime | None expected |
| Risk | Low |

## Deferred work

| Item | Task |
|---|---|
| Start / resume / autosave | EXAM-P5-T02 |
| Idempotent submit + PCS scoring | EXAM-P5-T03 |
| Attempt-linked snapshot persistence extras | EXAM-P5-T04 |
| Student player | EXAM-P6 |
| Course QUIZ_REFERENCE enrolment gate | ADR-024 §6 / later |

## Production source modified by EXAM-P5-T01: YES

## Stop condition

EXAM-P5-T02 **not started**. Awaiting explicit manual review.
