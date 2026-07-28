# EXAM-P5-T03 — Idempotent Submit + PCS Scoring

Status: IMPLEMENTED  
- Started: 2026-07-28  
- Implemented: 2026-07-28

## Standards loaded

AGENTS.md, CLAUDE.md, docs/standards/index.md (+ backend/testing/security), ADR-024 §2/§4, Product/Engineering roadmaps, EXAM-P5 ledger, EXAM-P5-T01/T02 evidence, EXAM-P4-T04 snapshot contract, Phase 2 answer-authority rules.

## Task understanding

EXAM-P5-T03 adds **transactional final submission** and **mark-based PCS scoring** for self-paced attempts. Scoring uses frozen snapshot keys only. Auto-submit, player UI, and EXAM-P5-T04 remain out of scope.

## Frozen acceptance

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

## Submission architecture

```text
Authenticated owner
        ↓
Lock owned IN_PROGRESS attempt (FOR UPDATE)
        ↓
Load frozen snapshot items + persisted answers
        ↓
ScorePCSAttempt (deterministic)
        ↓
Persist per-answer outcomes + aggregate totals
        ↓
Transition → SUBMITTED
        ↓
Learner-safe result (no answer keys)
```

## Scoring authority and rules

| Source | Role |
|---|---|
| Snapshot `authoritative_answer` | Preferred scoring key when non-empty |
| Snapshot `answers` | Operational fallback |
| Snapshot `official_answer` | Documentary only — **not** used to score |
| Quiz `negative_marks_per_question` | Penalty for incorrect answered singles |
| Snapshot `points` | Award on correct (default 1 if null) |

| Case | Outcome |
|---|---|
| Correct single | +points |
| Incorrect single | −negative marks |
| Unanswered | 0, `is_correct` null |
| Survey | unscored (0, `is_correct` null) |
| Aggregate | floor at 0; round to 2 dp |

## Attempt state transitions

`IN_PROGRESS` → submit → `SUBMITTED`  
`AUTO_SUBMITTED` deferred. Terminal attempts immutable; autosave already rejects non-`IN_PROGRESS`.

## Transaction and idempotency

- Single transaction: lock → score → write answers → finalize.
- Repeat submit on `SUBMITTED` returns stored result (HTTP 200).
- Finalize `WHERE status=IN_PROGRESS`; 0 rows → conflict path returns stored result.
- Failure rolls back completely.

## Security model

| Concern | Behaviour |
|---|---|
| Auth | Kratos required |
| Owner | session only |
| Client score/status | rejected |
| Result keys | omitted (correctness/marks allowed) |
| Editor inspection | unchanged separate route |

## Migration summary

None for T03.

## API changes (additive)

| Method | Path | Notes |
|---|---|---|
| POST | `/v1/quizzes/:quiz_id/attempts/:attempt_id/submit` | Submit / idempotent result |
| GET | `/v1/quizzes/:quiz_id/attempts/:attempt_id/result` | Read stored result |

## Checks (2026-07-28)

```text
go test ./... -count=1 → PASS
go build ./... → PASS
```

No frontend changes.

## Frozen-snapshot scoring verification

Unit tests assert authoritative key preference over operational/live-differing keys; survey unscored; negative marks + floor.

## Answer-leak verification

Result/submit payloads omit `official_answer` / `authoritative_answer` / `answer_review_status`.

## Compatibility verification

- T01/T02 create/autosave/resume preserved.
- Live quiz scoring untouched.
- Snapshot immutability preserved.

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
| Attempt-linked snapshot extras | EXAM-P5-T04 |
| Auto-submit on expiry | later / T03 explicitly deferred |
| Student player | EXAM-P6 |
| Full answer-key review policy | EXAM-P7 |

## Production source modified by EXAM-P5-T03: YES

## Stop condition

EXAM-P5-T04 **not started**. Awaiting explicit manual review.
