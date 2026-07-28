# EXAM-P4-T02 — Visual Test Builder UI

Status: IMPLEMENTED
- Started: 2026-07-27
- Implemented: 2026-07-27

## Standards loaded

AGENTS.md, CLAUDE.md, docs/standards/index.md (+ frontend/testing/security), ADR-024 §3/§8, Product/Engineering roadmaps, EXAM-P4 phase ledger, EXAM-P4-T01 evidence and collection APIs.

## Task understanding

EXAM-P4-T02 delivers the **Nuxt Visual Test Builder** on the quiz manage page. It consumes T01 collection CRUD/resolve APIs only — no new persistence, no parallel collection store, no inline question create (T03), no attempt snapshots (T04).

## Frozen acceptance

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

## UI architecture

```text
Quiz manage page (canEditQuiz)
  └── VisualTestBuilder
        ├── useQuizCollectionsApi → T01 routes only
        ├── Collection list (STATIC | DYNAMIC)
        ├── Create / edit details
        ├── STATIC: ordered membership from quiz questions → PUT /members
        ├── DYNAMIC: DynamicFilterFields → PATCH filter
        └── Resolve preview → GET /resolve
              ├── STATIC → "Ordered STATIC membership"
              └── DYNAMIC → METADATA_PENDING | ALL_QUIZ_QUESTIONS | IDs
```

Titles for resolved IDs are looked up from the existing quiz question list already loaded on the page. Resolve responses never surface answer keys.

## Migration summary

None. Reuses `20260727160000_create_question_collections` from T01.

## API changes

None. Frontend-only consumption of T01 APIs.

## Checks (2026-07-27)

```text
npm test -- --run quiz_collections / question_collection / VisualTestBuilder  → PASS (13)
eslint (changed T02 files)                                                    → PASS
npm run build                                                                 → PASS
```

## Compatibility verification

- Question Form / Import Wizard / revision history unchanged.
- Collection business rules remain server-owned.
- No backend schema or route changes.

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
| Inline Add New Question → bank + link | EXAM-P4-T03 |
| Attempt/test snapshot contract hooks | EXAM-P4-T04 |
| Full taxonomy filter resolution UI | EXAM-P10 |
| Browser runtime smoke (IronBee MCP unavailable in this session) | Manual / later |

## Production source modified by EXAM-P4-T02: YES
