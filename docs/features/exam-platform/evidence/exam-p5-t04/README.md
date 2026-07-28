# EXAM-P5-T04 — Resolved Question Snapshot Persistence

Status: IMPLEMENTED  
- Started: 2026-07-28  
- Implemented: 2026-07-28

## Standards loaded

AGENTS.md, CLAUDE.md, docs/standards/index.md (+ backend/testing/security), ADR-024 §3/§4, Product/Engineering roadmaps, EXAM-P5 ledger, EXAM-P4-T04 snapshot contract, EXAM-P5-T01–T03 evidence, existing attempt/scoring services.

## Task understanding

EXAM-P5-T04 completes the Attempt Engine foundation by **copying resolved snapshot items onto each attempt** at create and **freezing scoring configuration** (`negative_marks_per_question`, `expected_max_score`) so later quiz/bank edits cannot alter in-flight or historical scoring. Autosave, resume, and submit read attempt-linked freezes. Does **not** replace `test_snapshots`. Phase 6 player / analytics / auto-submit remain out of scope.

## Frozen acceptance

<!-- TASK:EXAM-P5-T04:ACCEPTANCE:START -->
- [x] At attempt create, copy resolved snapshot items into attempt-linked rows (IDs, revision, order, stem/options/keys, points) without replacing `test_snapshots`.
- [x] Freeze scoring config on the attempt (`negative_marks_per_question`, `expected_max_score`) at create; later quiz edits must not change in-flight or historical scoring.
- [x] Autosave validation, resume, and submit/scoring read attempt-linked snapshot items (not live bank/collections).
- [x] Submitted attempts remain immutable; attempt-linked snapshot rows are insert-once (no update API).
- [x] Learner payloads remain answer-key-safe; frozen scoring metadata may be exposed without keys.
- [x] Phase 6 player / analytics / auto-submit remain out of scope.
- [x] `go test ./...` and `go build ./...` pass; evidence under `docs/features/exam-platform/evidence/exam-p5-t04/`.
<!-- TASK:EXAM-P5-T04:ACCEPTANCE:END -->

## Snapshot integration architecture

```text
test_snapshots + test_snapshot_items  (shared immutable freeze — ADR-024)
        │
        │  copy at CreateInProgress (same TX)
        ▼
assessment_attempts
  • test_snapshot_id (FK, retained)
  • question_order (frozen)
  • negative_marks_per_question (frozen at create)
  • expected_max_score (frozen at create)
        │
        ▼
assessment_attempt_snapshot_items  (insert-once per attempt)
  • stem / options / keys / points / revision / order
        │
        ├─ Autosave validates options against attempt items
        ├─ Resume projects learner items from attempt items
        └─ Submit scores from attempt items + frozen neg marks
```

Shared `test_snapshots` remain the quiz-level contract. Attempt-linked rows are a **per-attempt copy**, not a parallel snapshot system.

## Attempt state / immutability

| Concern | Behaviour |
|---|---|
| Snapshot items | Insert-once in create TX; no update/delete API |
| Scoring config | Frozen on attempt at create |
| After submit | Attempt + answers immutable (existing T03 rules) |
| Pre-T04 attempts | Empty attempt-item set falls back to shared snapshot for learner projection |

## Security model

| Concern | Behaviour |
|---|---|
| Learner payloads | Omit answer keys; may expose frozen scoring metadata |
| Autosave | Validates against attempt-linked options only |
| Submit | Uses attempt-linked keys + frozen neg marks; never live bank |
| Ownership | Unchanged session identity rules |

## Migration summary

| Migration | Change |
|---|---|
| `20260728120000_create_assessment_attempt_snapshot_items` | Add `negative_marks_per_question`, `expected_max_score` on `assessment_attempts`; create `assessment_attempt_snapshot_items` with RESTRICT FKs and unique (attempt, position) / (attempt, question) |

Additive; down script drops table and columns.

## API changes

No new routes. Existing create / resume / get / result envelopes expose:

- `negative_marks_per_question`
- `expected_max_score` (when set)

Behaviour change: create persists attempt-linked items; autosave/resume/submit consume them.

## Checks (2026-07-28)

```text
go test ./... -count=1 → PASS
go build ./... → PASS
```

No frontend / Phase 6 player changes.

## Compatibility verification

- Shared `test_snapshots` / `test_snapshot_items` unchanged as the quiz freeze contract.
- T01–T03 routes preserved; scoring authority rules preserved with attempt-local key source.
- Live quiz / `user_played_quizzes` untouched.
- Pre-T04 attempts: learner projection falls back to shared snapshot when attempt items are empty.

## Production impact assessment

| Item | Value |
|---|---|
| Breaking changes | **NO** (additive schema + behaviour on new attempts) |
| Migration required | **YES** (`20260728120000`) |
| Downtime | None expected (additive DDL) |
| Risk | Low–medium (create path transactional; existing in-progress attempts without items use fallback) |

## Deferred work

| Item | Task |
|---|---|
| Student player UI | EXAM-P6 |
| Auto-submit on expiry | later |
| Analytics / leaderboards | later phases |
| Full answer-key review policy | EXAM-P7 / later |
| Backfill attempt items for pre-T04 rows | optional ops if needed |

## Production source modified by EXAM-P5-T04: YES

## Stop condition

**EXAM-P6 not started.** Awaiting explicit manual review and approval before any Phase 6 work.
