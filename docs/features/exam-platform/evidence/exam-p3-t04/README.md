# EXAM-P3-T04 — Import Pipeline Regression + Phase Closure Evidence

Status: IMPLEMENTED (awaiting manual review)
- Started: 2026-07-27
- Implemented: 2026-07-27

## Standards loaded

AGENTS.md, CLAUDE.md, docs/standards/index.md, ADR-024, Product/Engineering roadmaps, EXAM-P3 phase ledger, EXAM-P3-T01–T03 evidence.

## Task understanding

The phase ledger listed only T01–T03; **EXAM-P3-T04 is the final Phase 3 task** (mirroring EXAM-P2-T04). It does **not** add new import features. It delivers:

1. **Import route inventory** — documents all CSV import HTTP surfaces and auth requirements.
2. **Regression proof suite** — verifies the full T01–T03 pipeline still behaves correctly.
3. **Phase closure evidence** — production audit, compatibility notes, deferred risks.

T04 depends on T01 (preview jobs), T02 (wizard + commit), T03 (duplicate detection).

## Frozen acceptance

See phase ledger `EXAM-P3-T04 acceptance (frozen)`.

## Architecture notes

| Component | Purpose |
|---|---|
| `api/security/import_routes.go` | Inventory of preview / commit / legacy upload routes |
| `question_import_pipeline_regression_test.go` | Cross-cutting regression for T01–T03 behaviours |
| Existing T01–T03 tests | Preserved; T04 adds inventory + pipeline regression |

Commit still flows: `CommitPreviewJob` → `QuestionsFromImportPreviewRows` → `AppendQuestionsToQuiz` → `registerQuestions` + `RecordRevision`.

## Validation / processing rules (regression scope)

| Check | Covered by |
|---|---|
| Preview creation | T01 controller + T04 regression |
| Row validation | `csv_import_preview` tests |
| Duplicate classification | T03 + T04 regression |
| Authorized preview GET | T04 regression |
| Wrong-quiz job 404 | T04 regression |
| Authority defaults at commit | T04 regression (`QuestionsFromImportPreviewRows`) |
| Legacy duplicate rejection | T04 regression |
| Idempotent commit | T02 controller test |
| Concurrent commit 409 | T02 controller test |
| Revision creation | P2 `registerQuestions` + `RecordRevision`; commit uses same path |

## State-transition changes

**None.** T04 is verification-only.

## Migration summary

**None.**

## API changes

**None.**

## Checks (2026-07-27)

```text
go test ./security/ -run "Import" -count=1                              → PASS
go test ./controllers/api/v1/ -run "ImportPipeline|QuestionImportJob" -count=1 → PASS
go test ./... -count=1                                                  → PASS
go build ./...                                                          → PASS
```

Frontend: no changes in T04.

## Compatibility verification

- T01 preview/create/get unchanged.
- T02 wizard + commit unchanged.
- T03 duplicate rules unchanged.
- Legacy `/upload` unchanged except existing T03 duplicate enforcement.

## Production impact assessment

| Item | Value |
|---|---|
| Breaking changes | **NO** |
| Migration required | **NO** |
| Risk | Very low (tests + inventory only) |

## Residual risks and deferred work

| Item | Phase |
|---|---|
| XLSX import | Later / separate acceptance |
| Cross-quiz duplicate detection | Future if required |
| E2E Playwright import workflow | EXAM-P10 |
| Phase 4 collections / test builder | EXAM-P4 |

## Production source modified by EXAM-P3-T04: YES (security inventory + regression tests)
