# EXAM-P6-T01 — Instructions + Start Attempt UI

Status: IMPLEMENTED  
- Started: 2026-07-28  
- Implemented: 2026-07-28

## Standards loaded

AGENTS.md, CLAUDE.md, docs/standards/index.md (+ frontend/backend/testing), ADR-024 §3/§4/§10, Product/Engineering roadmaps, EXAM-P6 ledger, EXAM-P4/P5 evidence, existing attempt/snapshot/scoring services.

## Task understanding

EXAM-P6-T01 opens the **learner instructions + start/resume** surface for self-paced tests. It wires Nuxt to P5 attempt APIs (plus a key-safe instructions read) so learners can see exam rules and begin or resume an attempt. Question palette, autosave feedback, timer, and submit confirmation remain T02/T03.

## Frozen acceptance

<!-- TASK:EXAM-P6-T01:ACCEPTANCE:START -->
- [x] Authenticated learner instructions screen for a self-paced quiz + immutable `snapshot_id` (route/query), wired to real attempt/snapshot APIs — no mocked success payloads.
- [x] Instructions present exam rules without answer keys: quiz title, question count, duration, max attempts, negative marking, and start/resume eligibility.
- [x] Start Attempt calls P5 `POST /quizzes/:quiz_id/attempts` with `snapshot_id`; identity from session only; idempotent reuse of an owned `IN_PROGRESS` attempt is supported.
- [x] When an owned `IN_PROGRESS` attempt exists, Resume is offered and navigates into the attempt route using P5 list/get/resume contracts.
- [x] After start/resume, learner lands on a minimal attempt shell (no question palette, autosave UI, timer countdown, or submit confirmation — those remain T02/T03).
- [x] Unauthenticated callers are rejected by the API; learner UI never renders answer keys.
- [x] Analytics, answer review, leaderboards, live quiz player, and EXAM-P6-T02+ remain out of scope.
- [x] Frontend unit tests cover instructions/start/resume wiring; `go test ./...`, `go build ./...`, and applicable `npm` lint/test/build pass; evidence under `docs/features/exam-platform/evidence/exam-p6-t01/`.
<!-- TASK:EXAM-P6-T01:ACCEPTANCE:END -->

## Player architecture

```text
/attempt/quizzes/:quiz_id?snapshot_id=…
        │
        ▼
GET /v1/quizzes/:quiz_id/attempts/instructions?snapshot_id=…
  (key-safe rules + can_start / can_resume)
        │
        ├─ Start → POST /v1/quizzes/:quiz_id/attempts { snapshot_id }
        └─ Resume → GET …/attempts/:id/resume (ownership check)
                │
                ▼
/attempt/quizzes/:quiz_id/attempts/:attempt_id
  (minimal shell — T02 fills palette/autosave)
```

## Migration summary

None.

## API changes (additive)

| Method | Path | Notes |
|---|---|---|
| GET | `/v1/quizzes/:quiz_id/attempts/instructions?snapshot_id=` | Key-safe instructions; no items/keys |

Existing P5 create/resume/get used unchanged for start/resume/shell.

## Checks (2026-07-28)

```text
go test ./... -count=1 → PASS
go build ./... → PASS
npm test -- --run test/composables/assessment_attempts.test.js → PASS
eslint (T01 files) → PASS
npm run build → PASS
```

## Compatibility verification

- P5 attempt lifecycle unchanged.
- Shared snapshot immutability preserved; instructions never return items/keys.
- Live quiz player untouched.

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
| Question palette + autosave feedback | EXAM-P6-T02 |
| Timer + submit confirmation + expiry | EXAM-P6-T03 |
| Results / answer review | EXAM-P7 |
| Analytics / leaderboards | later |

## Production source modified by EXAM-P6-T01: YES

## Stop condition

**EXAM-P6-T02 not started.** Awaiting explicit manual review and approval.
