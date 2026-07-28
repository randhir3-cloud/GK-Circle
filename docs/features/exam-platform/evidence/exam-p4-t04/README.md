# EXAM-P4-T04 — Attempt/Test Snapshot Contract Hooks

Status: IMPLEMENTED
- Started: 2026-07-27
- Implemented: 2026-07-27

## Standards loaded

AGENTS.md, CLAUDE.md, docs/standards/index.md (+ backend/testing/security), ADR-024 §3/§4, Product/Engineering roadmaps, EXAM-P4 phase ledger, EXAM-P4-T01–T03 evidence, collection resolve + question revision services.

## Task understanding

EXAM-P4-T04 adds **immutable test composition snapshot contract hooks**: resolve STATIC/DYNAMIC collections once, freeze question IDs + revision identifiers + order + answer keys into `test_snapshots` / `test_snapshot_items`, and expose editor + learner-safe read APIs. This is **not** the Attempt Engine (no start/submit/score/player — those remain EXAM-P5).

## Frozen acceptance

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

## Snapshot architecture

```text
Collection definition (live, mutable)
        ↓ resolve once at create
Resolved question selection (ordered, de-duplicated)
        ↓ transactional freeze
Immutable test snapshot (IDs + revision_number + stem/options/keys)
        ↓ future EXAM-P5
Attempt consumption (not implemented here)
```

### Lifecycle and immutability rules

| Event | Behaviour |
|---|---|
| Snapshot create | Resolve all (or selected) collections once; copy freeze payload into `test_snapshot_items` |
| Collection membership/filter change after create | Does **not** alter existing snapshot rows |
| Live question edit / new revision after create | Does **not** alter existing snapshot rows |
| Snapshot read (editor) | Returns frozen payload including answer keys (quiz edit access) |
| Snapshot read (learner) | Same freeze without answers/official/authoritative/review status |
| Attempt start/submit/score | Deferred to EXAM-P5 |

### Empty / unresolved / duplicates

| Condition | HTTP | Error |
|---|---|---|
| No collections on quiz | 400 | `test snapshot requires at least one collection` |
| DYNAMIC `METADATA_PENDING` | 400 | unresolved collection |
| Zero resolved questions | 400 | requires at least one resolved question |
| Same question in multiple collections | 400 | duplicate question across collections |

## Migration summary

| Migration | Purpose |
|---|---|
| `20260727170000_create_test_snapshots` | `test_snapshots` + `test_snapshot_items` |

Run: `gk-circle migrate up`

## API changes (additive)

| Method | Path | Notes |
|---|---|---|
| POST | `/v1/quizzes/:quiz_id/test-snapshots` | Create immutable composition snapshot |
| GET | `/v1/quizzes/:quiz_id/test-snapshots` | List snapshots |
| GET | `/v1/quizzes/:quiz_id/test-snapshots/:snapshot_id` | Editor inspection (includes keys) |
| GET | `/v1/quizzes/:quiz_id/test-snapshots/:snapshot_id/learner` | Answer-key-safe view |

Auth: `KratosAuthenticated` + `QuizPermission` + `VerifyQuizEditAccess` (contract hooks are editor-gated until P5 attempt wiring).

## Checks (2026-07-27)

```text
go test ./models/ -run "TestTestSnapshot" -count=1                 → PASS
go test ./services/ -run "TestTestSnapshot" -count=1               → PASS
go test ./controllers/api/v1/ -run "TestCreateTestSnapshot|TestGetLearnerTestSnapshot" -count=1 → PASS
go test ./... -count=1                                             → PASS
go build ./...                                                     → PASS
```

No frontend changes → frontend lint/test/build not required for T04.

## Security verification

- Create/list/get require quiz edit access.
- Learner projection strips `answers`, `official_answer`, `authoritative_answer`, `answer_review_status` (asserted in tests).
- Snapshot writes are transactional; no in-place overwrite API.

## Compatibility verification

- Collection CRUD/resolve (T01–T03) unchanged.
- Question bank create/import/revision paths unchanged.
- `assessment_attempts` / attempt scoring untouched (P5).

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
| Attempt start / resume / autosave | EXAM-P5-T02 |
| Submit + PCS scoring | EXAM-P5-T03 |
| Attempt-linked snapshot persistence wiring | EXAM-P5-T04 |
| Student player | EXAM-P6 |
| Phase 4 formal closure | Manual phase-closure review |

## Production source modified by EXAM-P4-T04: YES
