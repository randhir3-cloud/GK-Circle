# EXAM-P3-T02 — Nuxt Import Wizard + Commit Workflow

Status: IMPLEMENTED (awaiting manual review)
- Started: 2026-07-27
- Implemented: 2026-07-27

## Standards loaded

AGENTS.md, CLAUDE.md, docs/standards/index.md, ADR-024, Product/Engineering roadmaps, EXAM-P3 phase ledger, EXAM-P3-T01 evidence, existing Question Bank creation (`AppendQuestionsToQuiz` / `registerQuestions`), Nuxt upload/API conventions.

## Task understanding

EXAM-P3-T02 delivers the **CSV import wizard and commit workflow**:

1. **Nuxt wizard** on the quiz manage page: CSV upload → validation preview → commit → success/failure summary.
2. **Commit API** that persists **eligible preview rows only** through the shared question-bank path (`QuestionModel.AppendQuestionsToQuiz` → `registerQuestions` with `ApplyAnswerAuthority` + `RecordRevision`), not raw bulk inserts.
3. **Import-job state machine** with transactional commit, idempotency, and concurrent-request protection.

Out of scope: duplicate detection (T03), XLSX, background jobs, collections, attempt/scoring changes.

## Frozen acceptance

<!-- TASK:EXAM-P3-T02:ACCEPTANCE:START -->

- [x] `QuizImportWizard` replaces direct legacy `/upload` in the quiz manage import modal; steps: upload → preview → commit → result.
- [x] Preview shows valid rows, invalid rows, and row-level validation messages from T01 job payload.
- [x] `POST /v1/quizzes/:quiz_id/questions/import-jobs/:import_job_id/commit` commits eligible preview rows (edit access + job belongs to quiz).
- [x] Commit uses `AppendQuestionsToQuiz` so each question gets `lineage_id`, revision 1, `question_revisions` entry, authority defaults, and `quiz_questions` link.
- [x] Commit policy: **valid-row-only** (invalid preview rows skipped); **all-or-nothing** transactional commit for the valid set.
- [x] Job states: `PREVIEWED`/`FAILED`(retryable) → `COMMITTING` → `COMMITTED` or `FAILED` (commit error); `COMMITTED` is idempotent.
- [x] Concurrent commit returns 409; no duplicate questions on repeated commit.
- [x] Migration extends `question_import_jobs` with `COMMITTING`/`COMMITTED`, `commit_result_json`, `committed_at`.
- [x] Focused Go + Vitest coverage; legacy `POST .../upload` retained.

<!-- TASK:EXAM-P3-T02:ACCEPTANCE:END -->

## Wizard and commit architecture

```text
Nuxt QuizImportWizard
  upload CSV → POST .../import-jobs (T01)
  preview    → GET  .../import-jobs/:id (optional refresh)
  commit     → POST .../import-jobs/:id/commit (T02, no body)

CommitPreviewJob (Go)
  verify quiz exists
  load job scoped by quiz_id
  if COMMITTED → return stored result (idempotent)
  if COMMITTING → 409 conflict
  if no valid rows → 400
  BEGIN tx
    SELECT … FOR UPDATE + status → COMMITTING
    QuestionsFromImportPreviewRows (re-apply ApplyAnswerAuthority)
    QuestionModel.AppendQuestionsToQuiz(tx, quiz_id, questions)
      → registerQuestions + RecordRevision + quiz_questions
    UPDATE job → COMMITTED + commit_result_json
  COMMIT tx
  on failure: ROLLBACK tx → MarkCommitFailed (FAILED + error, retry allowed)
```

## Import-job state transitions

| From | Event | To | Notes |
|---|---|---|---|
| PREVIEWED | commit (valid rows > 0) | COMMITTING | row lock |
| FAILED (preview, 0 valid) | commit | — | 400 not commitable |
| FAILED (commit error, valid > 0) | commit retry | COMMITTING | allowed |
| COMMITTING | success | COMMITTED | stores question_ids |
| COMMITTING | insert/finalize error | FAILED | after rollback |
| COMMITTING | concurrent request | — | 409 |
| COMMITTED | repeat commit | COMMITTED | idempotent 200 |

## Transaction and idempotency behaviour

- **Valid-row-only**: only `preview.valid_rows` are converted and inserted; `preview.errors` are never committed.
- **All-or-nothing**: question inserts + job finalization share one DB transaction; rollback leaves no partial questions.
- **Idempotency**: `COMMITTED` jobs return the stored `commit_result` without re-inserting.
- **Concurrency**: optimistic claim via `FOR UPDATE` + status check; second commit gets 409 while `COMMITTING`.
- **Retry**: commit failure sets `FAILED` with `commit_result.error`; user may retry commit (wizard shows error; backend accepts `FAILED` when `valid_row_count > 0`).

## Validation and error model

| Layer | Behaviour |
|---|---|
| Upload | T01 file/MIME/size + MaxRows validation |
| Preview | Per-row errors in job payload; UI lists row_number + messages |
| Commit | Re-applies `ApplyAnswerAuthority` from stored preview; no client authority body |
| Auth | Kratos session + quiz permission + `VerifyQuizEditAccess` |
| Job scope | `quiz_id` + `import_job_id` must match stored job |
| HTTP | 404 job not found; 400 no valid rows; 409 commit in progress; 500 on unexpected DB errors |

## Migration summary

| Migration | Purpose |
|---|---|
| `20260727150000_alter_question_import_jobs_commit` | `COMMITTING`/`COMMITTED` statuses, `commit_result_json`, `committed_at` |

Run: `gk-circle migrate up`

## API changes (additive)

| Method | Path | Notes |
|---|---|---|
| POST | `/v1/quizzes/:quiz_id/questions/import-jobs/:import_job_id/commit` | Commit preview rows (200, idempotent if already committed) |

Existing T01 create/get endpoints unchanged. Legacy `POST .../questions/upload` retained.

## Checks (2026-07-27)

```text
go test ./... -count=1                         → PASS
go build ./...                                 → PASS
npm test -- --run quiz_import_jobs.test.js     → PASS (3)
npm test -- --run QuizImportWizard.test.js     → PASS (4)
npx eslint (T02 changed files) --max-warnings 0 → PASS
npm run build                                  → PASS
```

## Compatibility verification

- Legacy immediate `/upload` endpoint unchanged.
- T01 preview jobs without commit remain `PREVIEWED`.
- Manual question create/editor paths untouched.
- Answer-key leak protections (P2-T03/T04) unchanged; commit requires edit access.

## Production impact assessment

| Item | Value |
|---|---|
| Breaking changes | **NO** |
| Migration required | **YES** (additive) |
| Downtime | None expected |
| Risk | Low–medium (new commit path; wizard replaces modal upload UX) |

## Deferred work

| Item | Task |
|---|---|
| Duplicate detection | EXAM-P3-T03 |
| XLSX import | Later phase |
| Background/async commit | Not required |

## Production source modified by EXAM-P3-T02: YES
