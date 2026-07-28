# EXAM-P3-T01 — CSV Import Job API + Validation Preview

Status: VERIFIED
- Started: 2026-07-27
- Verified: 2026-07-27

## Standards loaded

AGENTS.md, CLAUDE.MD, docs/standards/index.md (+ backend/testing/security), ADR-024, Product/Engineering roadmaps, EXAM-P3 phase ledger, EXAM-P2 closure state, existing CSV upload conventions.

## Task understanding

EXAM-P3-T01 delivers the **backend CSV import job + validation preview** API. Editors upload a CSV; the API validates each row, applies answer-authority defaults for preview, stores a job record, and returns a structured preview. **Questions are not inserted** into the bank. Commit/wizard UI and duplicate detection are T02/T03.

Reuse: existing CSV headers/parsers (`ValidateCSVFileFormat`), shared row validation, `ApplyAnswerAuthority`, quiz edit access + `ValidateCsv` middleware. Legacy `POST .../upload` remains for immediate import.

## Frozen acceptance

- [x] Import-job create/get APIs with quiz edit access.
- [x] Preview without question persistence.
- [x] Valid rows include ADR-024 authority defaults; row errors are structured.
- [x] File size + MaxRows=500 limits.
- [x] Shared validation with legacy upload path.
- [x] Migration + tests + evidence.

## Import architecture notes

```text
CSV upload (ValidateCsv)
        ↓
ValidateCSVFileFormat (headers / unmarshal)
        ↓
PreviewQuestionsFromCSV (per-row validate + BuildImportPreviewRow)
        ↓
question_import_jobs (PREVIEWED | FAILED)  ← no questions insert
        ↓
GET import-jobs/:id → stored preview
```

Commit path (out of scope): future T02 must call `QuizService.AppendQuestionsToQuiz` / `CreateQuestion` so lineage, revisions, and authority are recorded.

## Validation and error model

| Layer | Behaviour |
|---|---|
| File | MIME + `MAX_QUIZ_FILE_SIZE`; empty file rejected |
| Rows | Max 500 (`constants.MaxRows`) |
| Per-row | Question text, type, ≥2 options, correct answer refs, media, points |
| Preview | Valid rows + `{row_number, messages[]}` errors; does not fail whole file |
| Job status | `PREVIEWED` if any valid rows; `FAILED` if only errors |
| Legacy `/upload` | Still all-or-nothing reject via `ExtractQuestionsFromCSV` |

## Migration summary

| Migration | Purpose |
|---|---|
| `20260727140000_create_question_import_jobs` | Job + preview JSON storage |

Run: `gk-circle migrate up`

## API changes (additive)

| Method | Path | Notes |
|---|---|---|
| POST | `/v1/quizzes/:quiz_id/questions/import-jobs` | Multipart CSV → create preview job (201) |
| GET | `/v1/quizzes/:quiz_id/questions/import-jobs/:import_job_id` | Fetch preview job |

Auth: `KratosAuthenticated` + `QuizPermission` + `VerifyQuizEditAccess` (+ `ValidateCsv` on POST).

## Checks (2026-07-27)

```text
go test ./utils/ -run "PreviewQuestions|ExtractQuestionsFromCSV" -count=1 → PASS
go test ./models/ -run "ImportPreview|QuestionImportJob" -count=1         → PASS
go test ./controllers/api/v1/ -run "QuestionImportJob" -count=1           → PASS
go test ./... -count=1                                                    → PASS
go build ./...                                                            → PASS
```

## Compatibility verification

- Existing `POST .../questions/upload` unchanged in behaviour (now also enforces MaxRows).
- Manual question create / T01–T02 editor paths untouched.
- No Nuxt changes.

## Production impact assessment

| Item | Value |
|---|---|
| Breaking changes | **NO** |
| Migration required | **YES** (additive) |
| Downtime | None expected |
| Risk | Low |

## Deferred items

| Item | Task |
|---|---|
| Nuxt import wizard | EXAM-P3-T02 |
| Commit/persist previewed rows | EXAM-P3-T02 |
| Duplicate detection | EXAM-P3-T03 |
| XLSX | Later phase/acceptance |

## Production source modified by EXAM-P3-T01: YES
