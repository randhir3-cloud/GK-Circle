# EXAM-P4-T01 — Collection Entities (STATIC + DYNAMIC Filters)

Status: IMPLEMENTED
- Started: 2026-07-27
- Verified: 2026-07-27

## Standards loaded

AGENTS.md, CLAUDE.MD, docs/standards/index.md (+ backend/testing/security), ADR-024 §3, Product/Engineering roadmaps, EXAM-P4 phase ledger, EXAM-P1–P3 evidence.

## Task understanding

EXAM-P4-T01 introduces **Question Collection entities** for quiz-scoped test composition:

- **STATIC** collections with ordered question membership (`question_collection_members`).
- **DYNAMIC** collections with validated filter JSON (subject, topic, year, difficulty, PYQ status per ADR-024 §3).
- CRUD + membership replace + resolve preview APIs (editor auth).
- No Visual Test Builder UI, inline question create, or attempt snapshot persistence (T02–T04).

Reuse: quiz edit access (`VerifyQuizEditAccess`), `quiz_questions` linkage validation, P2 revision/authority untouched, P3 import pipeline untouched.

## Frozen acceptance

- [x] Additive migration for `question_collections` and `question_collection_members`.
- [x] `STATIC` and `DYNAMIC` kinds with DB constraints (STATIC has no filter; DYNAMIC requires `filter_json`).
- [x] Quiz-scoped CRUD APIs with editor authentication.
- [x] STATIC ordered membership replace validates questions are linked to the quiz.
- [x] DYNAMIC filter schema validation and resolve preview endpoint.
- [x] Metadata filters return `METADATA_PENDING` until question bank taxonomy ships (EXAM-P10).
- [x] Empty dynamic filter resolves to all quiz-linked questions.
- [x] Model + controller tests; `go test ./...` and `go build ./...` pass.
- [x] Evidence pack under `docs/features/exam-platform/evidence/exam-p4-t01/`.

## Architecture notes

```text
quiz (existing)
  └── question_collections (kind STATIC | DYNAMIC)
        ├── STATIC → question_collection_members → questions (via quiz_questions validation)
        └── DYNAMIC → filter_json (stored criteria)
              └── resolve preview → question IDs (or METADATA_PENDING)
```

| Concept | Implementation |
|---|---|
| Scope | Collections belong to one quiz (`quiz_id` FK, CASCADE delete) |
| STATIC membership | Ordered `question_id` list; members must exist in `quiz_questions` |
| DYNAMIC filters | JSON: `subject`, `topic`, `year`, `difficulty`, `pyq_status` |
| Resolution | Read-only preview; does not persist attempt snapshots (EXAM-P4-T04 / P5) |
| Bank path | No parallel question store; members reference existing `questions` rows |

## Migration summary

| Migration | Purpose |
|---|---|
| `20260727160000_create_question_collections` | `question_collections` + `question_collection_members` |

Run: `gk-circle migrate up`

## API changes (additive)

| Method | Path | Notes |
|---|---|---|
| GET | `/v1/quizzes/:quiz_id/collections` | List collections for quiz |
| POST | `/v1/quizzes/:quiz_id/collections` | Create STATIC or DYNAMIC collection |
| GET | `/v1/quizzes/:quiz_id/collections/:collection_id` | Get collection (+ STATIC members) |
| PATCH | `/v1/quizzes/:quiz_id/collections/:collection_id` | Update title, position, DYNAMIC filter |
| DELETE | `/v1/quizzes/:quiz_id/collections/:collection_id` | Delete collection |
| PUT | `/v1/quizzes/:quiz_id/collections/:collection_id/members` | Replace STATIC membership (ordered) |
| GET | `/v1/quizzes/:quiz_id/collections/:collection_id/resolve` | Resolve preview (no snapshot persist) |

Auth: `KratosAuthenticated` + `QuizPermission` + `VerifyQuizEditAccess`.

### Example: create DYNAMIC collection

```json
POST /v1/quizzes/{quiz_id}/collections
{
  "title": "All linked questions",
  "kind": "DYNAMIC",
  "position": 0,
  "filter": {}
}
```

### Example: create STATIC collection + members

```json
POST /v1/quizzes/{quiz_id}/collections
{ "title": "Section A", "kind": "STATIC", "position": 0 }

PUT /v1/quizzes/{quiz_id}/collections/{collection_id}/members
{ "question_ids": ["uuid-1", "uuid-2"] }
```

## Checks (2026-07-27)

```text
go test ./models/ -run "QuestionCollection|CollectionDynamic" -count=1      → PASS
go test ./controllers/api/v1/ -run "QuestionCollection" -count=1          → PASS
go test ./... -count=1                                                    → PASS
go build ./...                                                            → PASS
```

## Compatibility verification

- Existing question create/edit/import flows unchanged.
- No Nuxt UI changes (T02).
- No attempt snapshot tables or hooks (T04).
- DYNAMIC metadata filters stored but not resolved until bank taxonomy (EXAM-P10).

## Production impact assessment

| Item | Value |
|---|---|
| Breaking changes | **NO** |
| Migration required | **YES** (additive) |
| Downtime | None expected |
| Risk | Low |

## Out of scope (per authorisation)

| Item | Task |
|---|---|
| Visual Test Builder UI | EXAM-P4-T02 |
| Inline Add New Question → bank + link | EXAM-P4-T03 |
| Attempt/test snapshot contract hooks | EXAM-P4-T04 |
| Full PYQ/taxonomy resolution | EXAM-P10 |
| XLSX import, scoring, player, analytics | Later phases |

## Production source modified by EXAM-P4-T01: YES
