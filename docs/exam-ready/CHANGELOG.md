# Changelog — Phase 3.1.8

> Log every completed ticket here. No code explanations — only what changed, what files, what tests.

---

## 2026-07-10 — RC-1 Product Management Verification (pre-commit gate)

**Scope:** Verification-only pass for product hardening (dashboard, dependency delete UX, path tabs, breadcrumbs). No new features.

**Verified:**
- Frontend `tsc --noEmit` PASS; production build PASS
- Backend `tsc --noEmit` PASS; Jest `products.service.spec.ts` **31/31**
- Playwright product suite **9/9** (`product-management`, `product-dashboard`, `products-sidebar`)
- Playwright navigation suite **6/6** (`navigation`, `navigation-contract`, `content-navigation`)
- Browser dialog audit: **0** `window.alert` / `window.confirm` / `window.prompt` in production source
- Manual (IronBee): delete-blocked dialog with dependency counts + Manage Test Series/Enrollments navigation; product dashboard tabs + stat cards on QA acceptance product

**Verification-only change:** `frontend/playwright/product-dashboard.spec.ts` — wait for single search result before click (strict-mode flake fix).

**Docs:** `CURRENT_STATUS.md`, `CHANGELOG.md`

---

## 2026-07-10 — Content Management Navigation (Educational OS sidebar)

**Scope:** Unified Creator/Admin navigation aligned to content authoring pipeline.

**Navigation:**
- Sidebar groups: **Content**, **People**, **Moderation**, **System**
- Content section (both roles): Products, Test Series, Subjects, Topics, Question Bank, Tests
- `super_admin` included in admin nav permission checks

**Routes:**
- New admin pages: `/admin/subjects`, `/admin/topics`, `/admin/questions`, `/admin/tests`
- Shared management components with `basePath` (`/creator` | `/admin`)
- Creator/Admin platform quick actions for content pipeline

**Verified:** Frontend tsc + build; Playwright `navigation-contract.spec.ts` + `content-navigation.spec.ts` **5/5**.

**Docs:** `CURRENT_STATUS.md`, `BOARD.md`, `ROADMAP.md`

---

## 2026-07-10 — Product Management (Phase A+B)

**Scope:** Foundational product catalog management before Punjab PCS content creation.

**Backend:**
- Paginated `GET /products/owned` with search, filters, sort
- Lifecycle routes: publish, unpublish, archive, restore
- Delete returns **409** when test series, enrollments, circles, or creator earnings exist
- `ProductOwnershipGuard` — `super_admin` admin bypass aligned with `admin`

**Frontend:**
- Shared `ProductManagementPage` at `/creator/products` and `/admin/products`
- Products added to Creator and Admin navigation
- Test Series: empty-state CTA when no products; admin uses managed-products API
- Select dropdown z-index fix inside modals

**Verified:** Backend tsc + Jest **881/881**; Frontend tsc + production build; Playwright `product-management.spec.ts` **3/3**.

**Docs:** `CURRENT_STATUS.md`, `BOARD.md`, `ROADMAP.md`

---

## 2026-07-10 — RC-1: Punjab PCS Personal Study Release Candidate

**Tag:** `rc-1-personal-study`

**P1 fixes:**
- Exam answer save race — immediate PATCH + `flushPendingExamAnswerSaves()` before submit (`ExamRenderer`, `attempt/page.tsx`, `exam-pending-saves.ts`)
- Test mode dialog — Practice/Exam visible while eligibility loads (`TestModeDialog.tsx`)
- Legacy Playwright — submit confirmation helper, answer-save waits, dialog loading waits (`helpers/exam-submit.ts`, `assessment-engine.spec.ts`, `test-attempt-modes.spec.ts`, `helpers/auth.ts`)

**Exam Fast Track (accepted prior to RC tag):** Subject/Topic/Test Series management, Test Builder, Exam Player, Results, Wrong Notebook, PYQ, question images.

**Verified:** Backend tsc + Jest **869/869**; Prisma validate + 32 migrations up to date; Frontend tsc + production build; Exam Fast Track Playwright **41/41** (workers=1).

**Docs:** `CURRENT_STATUS.md`, `BOARD.md`, `ROADMAP.md`, `releases/rc-1-personal-study.md`

---

## 2026-07-05 — TICKET-006: Image Support — Backend (Gate 2)

**Endpoint:** `POST /questions/:id/image` — `QuestionOwnershipGuard`; **MIME allowlist** (jpeg/png/webp/gif → 400 otherwise); **5 MB** Multer cap (→ 413); reuses the shared **`StorageService`** (NOT the product-coupled `UploadsController`), storing under the **`questions/`** directory; persists via `QuestionsService.setImageUrl(id, url)` (thin controller); returns `{ imageUrl }`.
**DTOs:** optional `imageUrl` added to Create/Update (auto-passthrough via `...rest`; no migration — `Question.imageUrl` exists since TICKET-001). `findOne` already returns `imageUrl` (via `include`).
**Snapshot immutability (v3):** `CURRENT_SNAPSHOT_VERSION 2 → 3`; `buildContentSnapshot` freezes `imageUrl`; it **participates in the content hash** (`canonicalizeSnapshot`) but is **omitted when null**, so every pre-existing v1/v2 content hash stays valid; `deserializeSnapshot` gains a v3 branch (v1/v2 default `imageUrl: null`). Publish (`buildContentSnapshot`) therefore freezes the image the candidate saw.
**Decisions honored:** Snapshot v3 (hashed, not display-only); StorageService not UploadsController; 5 MB; jpeg/png/webp/gif; `setImageUrl` service method; `questions/` dir.

**Files:** `questions.controller.ts`, `questions.service.ts`, `dto/create-question.dto.ts`, `dto/update-question.dto.ts`, `tests/services/snapshot-serializer.ts`; specs: `questions.controller.spec.ts` [NEW], `questions.service.spec.ts`, `tests/services/snapshot-serializer.spec.ts` (+ fixture updates in grading/assessment/test-attempts specs).
**Verified:** `tsc` 0; backend Jest **791/791** (new: controller upload/no-file/bad-mime/allowed-types; service persist/404; snapshot image-in-hash/null-omitted/v1-v2-v3-compat). **Live:** upload→201 `{imageUrl}` under `questions/`; persisted (GET); bad-mime→400; oversized→413; cross-owner→403.

## 2026-07-04 — TICKET-005-RB: Import History + Rollback

**Backend (canonical `import/` module, per CONF-01):**
- `GET /questions/import/history` — paginated history for the requesting creator (id, fileName, createdAt, importedCount, errorCount, duplicateCount, skippedCount, totalRows, status, rolledBackAt).
- `DELETE /questions/import/:historyId/rollback` — soft-deletes the questions the import created (`ImportHistory.importedIds`), sets `status='rolled_back'` + `rolledBackAt`. Owner (or admin/super_admin) only → 403; 404 if not found; **idempotent** (second call = no-op); rejects imports older than **24h** (400).
- `ImportRepository`: `findImportHistoryById`, `listImportHistory`, `softDeleteQuestions` (in-tx).

**Frontend:**
- `questionsService.getImportHistory` / `rollbackImport` + types.
- `/creator/questions/import/history/page.tsx` — DS Table (Date/File/Imported/Errors/Status/Actions), status Badges, per-row **Undo** (disabled when >24h or already rolled back) → `ConfirmDialog` → `useToast` → refresh. All Phase-1 primitives.

**RB-01 — View Errors:** added `GET /questions/import/history/:historyId` (owner/admin; 404/403) returning the full record incl. `errorReport` — kept OUT of the list so list responses stay lightweight (per REST review). History page rows with errors show an expandable **View Errors** panel (lazy-loaded row-level `{rowNumber, questionText, errors[]}`).
**RB-02 — Playwright:** `frontend/playwright/tests/question-import-history.spec.ts` — seeds an import (valid+invalid rows) via API, then drives the UI: history row appears → View Errors expands → Undo → confirm → success toast → status "rolled back" → Undo disabled; plus an empty/load-safe test.

**Files:** `import/{controllers,services,repositories}/*`, `frontend/.../import/history/page.tsx` [NEW], `frontend/src/services/questions.service.ts`, `playwright/tests/question-import-history.spec.ts` [NEW].
**Verified:** BE+FE `tsc` 0; backend Jest **777/777** (incl. 10 RB unit tests: rollback 404/403/admin/idempotent/24h/soft-delete + detail 404/403/owner/admin); **live** import→history→rollback→idempotent; **Playwright 2/2**.
**Known limitation:** ISSUE-004 (uploaded images not cleaned on rollback) already tracked.

## 2026-07-04 — TICKET-005-CONF-Gate1-FIX-01: Preserve Structured Error Envelope + Gate 1 PASS

**Problem (found in live release verification):** the global `HttpExceptionFilter` flattened every error to `{success, statusCode, timestamp, error}`, dropping the structured `code`/`suggestedAction`; and NestJS raised the Multer size overflow as a `PayloadTooLargeException` before the controller-scoped `ImportUploadExceptionFilter` (`@Catch(MulterError)`) could map it.

**Fixed:**
- `HttpExceptionFilter` now preserves structured payloads **backward-compatibly** — keeps the existing `error` field and adds `code`/`message`/`suggestedAction` when the thrown response object carries them.
- `ImportUploadExceptionFilter` now `@Catch(MulterError, PayloadTooLargeException)` and emits the aligned envelope, so oversized uploads on import routes return `IMPORT_FILE_TOO_LARGE`.

**Files:** `common/filters/http-exception.filter.ts`, `import/filters/import-upload-exception.filter.ts` (+ spec).
**Verified:** `tsc` clean; filter (4) + `import.controller.spec` (8) pass. **Live:** oversized→413 `IMPORT_FILE_TOO_LARGE`, unsupported→400 `IMPORT_UNSUPPORTED_FORMAT`, dup-headers→400 `IMPORT_DUPLICATE_HEADERS`, empty→400 `IMPORT_EMPTY_FILE` — all carry `code`+`message`+`suggestedAction` (and `error`).

### ✅ GATE 1 RELEASE VERIFICATION — PASS
CONF-01…06 + FIX-01 verified against a running stack (PG/Redis/backend/frontend):
- Compilation: BE+FE `tsc` 0 errors. Unit: **766/766 pass**. Playwright `question-import-mapping`: **2/2 pass**.
- Live: old route→404 / new route→200 (CONF-01); import executes with real counts (CONF-02); cross-owner replace→**403** (CONF-03); **two concurrent imports → exactly 1 created** vs real Postgres (CONF-04); re-validate reset (CONF-05); predictable upload failures with full structured contract (CONF-06 + FIX-01).
- Out of scope / separate track: 12 pre-existing e2e suite failures in live-tests/battle/products (`RCF-Regression-01`) — not attributable to questions-only changes (app boots, unit green, those suites don't import `questions/`).

## 2026-07-04 — TICKET-005-CONF-06: Upload Hardening

**Verified already present (structured `{code,message,suggestedAction}` in `header-parser.service.ts`):** file-size (`IMPORT_FILE_TOO_LARGE`), empty workbook (`IMPORT_EMPTY_FILE` / no-sheet), duplicate headers (`IMPORT_DUPLICATE_HEADERS`), unsupported type (`IMPORT_UNSUPPORTED_FORMAT`), row limit (`IMPORT_TOO_MANY_ROWS`), column limit (`IMPORT_TOO_MANY_COLUMNS`).

**Completed (closed cert finding P1-2 — post-buffer size check = OOM/DoS vector):**
- Added HTTP-layer Multer limits `{ fileSize: 25MB, files: 1 }` to all three `ImportController` `FileInterceptor`s so oversized/multi-file uploads abort before unbounded in-memory buffering.
- Added `ImportUploadExceptionFilter` (`@Catch(MulterError)`) mapping `LIMIT_FILE_SIZE` → 413 `IMPORT_FILE_TOO_LARGE`, `LIMIT_*_FILE` → 400 `IMPORT_TOO_MANY_FILES`, else 400 `IMPORT_UPLOAD_FAILED` — same structured contract (no opaque 500s).

**Files:** `import.controller.ts`, `filters/import-upload-exception.filter.ts` [NEW], `filters/import-upload-exception.filter.spec.ts` [NEW].
**Tests:** `tsc` clean; filter spec (3) + `import.controller.spec` (8) pass. **Independent Verification ⏳** (live oversized-upload returns 413 structured via interceptor→filter wiring).

## 2026-07-04 — TICKET-005-CONF-05: Wizard State Reset on Re-validate

**Changed:** `useQuestionImport.validate()` now clears `selectedRowNumbers`, `hiddenRowNumbers`, and `activeFilter` on every (re)validation; the page's new `handleValidate` also clears page-local `expandedRows` before delegating. Errors/warnings come from the fresh report. No stale row-level UI state survives a re-validation.
**Files:** `hooks/useQuestionImport.ts`, `import/page.tsx`. **Tests:** frontend `tsc` clean.

## 2026-07-04 — TICKET-005-CONF-04: Transaction-Safe Duplicate Handling

**Changed:**
- Added `ImportRepository.findQuestionByNormalizedText(text, tx)` — authoritative in-transaction dedup lookup.
- `ImportService.importQuestions` now re-checks each NEW row inside the transaction (validation is advisory); a late DB match is treated as a duplicate honoring the `skip`/`replace` strategy, and `replace` reuses the CONF-03 ownership guard (creator → own only; admin/super_admin → any; else `ForbiddenException`). Rows already flagged during validation skip the extra lookup (no redundant query).
- Persisted a PostgreSQL isolation-level note (READ COMMITTED default; SERIALIZABLE would need retry handling; durable guarantee is the DB uniqueness layer) as a code comment at the re-check site.

**Wording correction (per review):** the unit test **deterministically simulates the expected concurrent outcome** — it is explicitly NOT a proof of true concurrency; comment updated in both `import.service.ts` and `import.service.spec.ts`.

**Files:** `import.service.ts`, `import.repository.ts`, `import.service.spec.ts`.
**Tests:** `tsc --noEmit` clean; `import.service.spec` → 34/34 pass (late replace, admin bypass, creator forbidden, flag, skip-known-duplicate, deterministic concurrency simulation).
**Status:** Implementation ✅ · Hostile Review ✅ · **Independent Verification ⏳ Required** (run two simultaneous imports of the same question against a live PostgreSQL DB) · Release Certification Pending.

## 2026-07-04 — Planning Phase

**Completed:** Full implementation planning (4 rounds of review)

**Documents Created:**
- `docs/exam-ready/MASTER_PLAN.md`
- `docs/exam-ready/ROADMAP.md`
- `docs/exam-ready/CURRENT_STATUS.md`
- `docs/exam-ready/DECISIONS.md`
- `docs/exam-ready/KNOWN_ISSUES.md`
- `docs/exam-ready/AGENT_HANDOFF.md`
- `docs/exam-ready/BOARD.md`
- `docs/exam-ready/tickets/TICKET-001.md` through `TICKET-050.md`

## 2026-07-04 — TICKET-005-CONF-01: Resolve Import Route Collision

**Problem:** `POST /questions/import` was registered by TWO live controllers — the legacy
`QuestionsController` (`@Post('import')`, `@Post('validate-import')`) and the canonical
`ImportController` (`@Controller('questions/import')`) — a duplicate runtime route registration.

**Changed:**
- Removed the duplicate import handlers (`validateImport`, `importQuestions`) and the
  `QuestionImportService` dependency from `backend/src/questions/questions.controller.ts`.
- Removed the now-unused `QuestionImportService` provider/export from `backend/src/questions/questions.module.ts`.
- `ImportController` is now the single implementation: `POST /questions/import/parse-headers`,
  `POST /questions/import/validate`, `POST /questions/import`, `GET /questions/import/template`.

**Backward compatibility:** `POST /questions/import` (write) still works — now served only by
`ImportController` with identical parse→validate→import behavior. The legacy
`POST /questions/validate-import` is removed; canonical validate path is `POST /questions/import/validate`
(already the path the frontend uses).

**Files:** `questions.controller.ts` [MODIFY], `questions.module.ts` [MODIFY].
Deprecated `question-import.service.ts` + `dto/import-questions.dto.ts` are now unreferenced by
runtime code (kept only by the service's own unit spec); flagged for a later cleanup ticket.

**Tests:** `npx tsc --noEmit` clean; `import.controller.spec` + `question-import.service.spec` → 24/24 pass.
Confirmed exactly one `@Controller('questions/import')` registration remains; no e2e drove the removed routes.

## 2026-07-04 — TICKET-006A: Column Mapping Wizard

**Implemented:**
- Modularized backend question import into specialized sub-services (`HeaderParserService`, `MappingService`, `TemplateService`, `ValidationService`, and `ImportRepository`).
- Created dynamic Excel import template generator with versioned metadata tags (`v1`) and support for target `examType` / `pyqExam` fields in the `Question` DB model.
- Created `ColumnMapper` component displaying mapping fields with searchable selectors, color-coded border confidence mappings, duplicate mapping warnings, and progress stats.
- Created `ProfileManager` component to save/load/delete custom named mapping profiles from `localStorage` (`gk-circle-question-import-profiles`).
- Created `PreviewCard` displaying cell contents of the first 3 workbook rows mapped dynamically.
- Created `IgnoredColumns` listing workbook columns that are not selected.
- Updated `useQuestionImport` hook and page controller endpoints to leverage the new structured mapping parser.

**Files Changed:**
- `backend/src/questions/questions.module.ts` — [Register import submodule module]
- `backend/src/questions/question-import.service.ts` — [Refactor to use new sub-services]
- `backend/src/questions/import/` — [New submodule folder with controller, services, DTOs, repository, and shared constants/aliases/types]
- `frontend/src/services/questions.service.ts` — [Add parseHeaders API types and client methods]
- `frontend/src/app/(creator)/creator/questions/import/hooks/useQuestionImport.ts` — [Update hook logic with mapping profiles and parsing states]
- `frontend/src/app/(creator)/creator/questions/import/page.tsx` — [Update wizard pages and components mounting]
- `frontend/src/app/(creator)/creator/questions/import/components/ColumnMapper.tsx` — [New UI component]
- `frontend/src/app/(creator)/creator/questions/import/components/ProfileManager.tsx` — [New UI component]
- `frontend/src/app/(creator)/creator/questions/import/components/PreviewCard.tsx` — [New UI component]
- `frontend/src/app/(creator)/creator/questions/import/components/IgnoredColumns.tsx` — [New UI component]
- `frontend/scripts/run-playwright.mjs` — [Forward process.argv to Playwright CLI]

**Tests Added:**
- `backend/src/questions/import/controllers/import.controller.spec.ts` — [Unit tests checking controller endpoints and validation boundaries]
- `frontend/playwright/tests/question-import-mapping.spec.ts` — [E2E Playwright test covering import mapping workflow]

**Migration:** none

---

## 2026-07-04 — TICKET-005: Import Validation Screen

**Implemented:**
- Decoupled React hook `useQuestionImport` for managing import states (file selection, validation, filtering, selection, hidden sets).
- Drag-and-drop file upload zone with visual dragging highlights and immediate file metadata overview card (name, size, extension) before validation.
- Integration stepper progress headers (`1 Upload → 2 Validate → 3 Confirm → 4 Import`).
- Metrics overview cards (Total Rows, Valid Rows, Warnings, Errors, Selected).
- List filtering by All, Errors, Warnings, and Valid states.
- Virtualized list rendering using `@tanstack/react-virtual` handling 10,000+ records.
- Expandable detail panel for table rows (keyboard & click) exposing subject, topic, options list, and raw import row JSON data.
- Row selection checklist checks (Select All, Errors, Warnings, individual rows).
- "Hide Invalid Rows" soft-hide state with Undo recovery banner.
- Verified dark-mode high contrast contrast, mobile responsive layouts, and ARIA keyboard focus states.
- Appended "Import Questions" navigation link next to "New Question" on the Question Bank page.

**Files Changed:**
- `frontend/package.json` — [Add @tanstack/react-virtual dependency]
- `frontend/src/services/questions.service.ts` — [Declare validation types and validateImport method]
- `frontend/src/app/(creator)/creator/questions/import/hooks/useQuestionImport.ts` — [Implement state hook]
- `frontend/src/app/(creator)/creator/questions/import/page.tsx` — [Implement validation list UI]
- `frontend/src/app/(creator)/creator/questions/page.tsx` — [Add navigation button link]

**Tests Added:** none (Playwright verification scheduled for TICKET-005-CONF).

**Migration:** none

---

## 2026-07-04 — TICKET-004A: Structured Validation Response

**Implemented:**
- Refactored `RowError` in backend validator to return structured validation errors with `message` and `suggestedAction`.
- Added columns count and sheet name parameter returns to the validation report to avoid client-side workbook re-parsing.
- Configured 15 backend error check blocks (missing options, type validation, difficulty normalization, correct answers, subject/topic database matching) to return suggested actions.
- Updated spec tests in `question-import.service.spec.ts` to assert on structured errors and `ParsedFile` responses.
- Verified NestJS compiles cleanly and unit tests pass (**740/740** total).

**Files Changed:**
- `backend/src/questions/question-import.service.ts` — [Refactor validation output structures]
- `backend/src/questions/questions.controller.ts` — [Update controller endpoints parsing calls]
- `backend/src/questions/question-import.service.spec.ts` — [Update spec test cases]

**Tests Added:**
- `backend/src/questions/question-import.service.spec.ts` — Updated error asserts and file parsed structures.

**Migration:** none

---

## 2026-07-04 — TICKET-004: CSV Import Backend

**Implemented:**
- State-machine CSV parser supporting quoted strings, double-quotes for escaping (`""`), and embedded line breaks.
- Delimiter auto-detection on the header row outside quotes supporting comma `,`, semicolon `;`, and tab `\t`.
- UTF-8 BOM detection and auto-stripping.
- Automatic routing between Excel and CSV based on file name extension with buffer compound file/zip magic bytes auto-detection fallback.
- Empty row filtering.
- Reused the core validation, duplicate check, transaction, and history logging pipeline.

**Files Changed:**
- `backend/src/questions/question-import.service.ts` — [Implement parseCSV, detectDelimiter, and auto-route parseFile]
- `backend/src/questions/questions.controller.ts` — [Pass originalName from controller endpoints]
- `backend/src/questions/question-import.service.spec.ts` — [Add unit/integration tests for CSV parsing, auto-routing, delimiters, quotes, line breaks, empty row filtering, and BOM stripping]

**Tests Added:**
- `backend/src/questions/question-import.service.spec.ts` — [7 unit/integration tests]. Full suite **740/740** passing (0 regressions).

**Migration:** none

---

## 2026-07-04 — TICKET-003: Question Import — Backend

**Implemented:**
- Bulk question import logic with Excel/CSV parsing via SheetJS (`xlsx`).
- Row-level validation checking option count, question type validation, difficulty normalization, and correct answer flags.
- Case-insensitive database lookup for Subject and Topic names.
- Duplicate detection including exact duplicate checks (SHA-256 hash of normalized text) and near-duplicate checks (Levenshtein distance matching with >= 90% similarity and length-difference filtering).
- DB transaction import supporting `skip`, `flag`, and `replace` duplicate strategies.
- `ImportHistory` audit logging logging total, imported, skipped, and error rows.

**Files Changed:**
- `backend/package.json` — [Add xlsx dependency]
- `backend/package-lock.json` — [Updated dependencies]
- `backend/src/questions/dto/import-questions.dto.ts` — [New DTOs for import and mapping request validation]
- `backend/src/questions/question-import.service.ts` — [New service implementing parsing, validation, deduplication, and transactions]
- `backend/src/questions/questions.controller.ts` — [Add validate-import and import endpoints]
- `backend/src/questions/questions.module.ts` — [Wire up and export QuestionImportService]

**Tests Added:**
- `backend/src/questions/question-import.service.spec.ts` — [12 unit tests for helpers, parsing, validation, duplicate checks, strategies, and failure logging]

**Migration:** none

---

## 2026-07-04 — TICKET-002: Publish Snapshot

**Implemented:**
- `TestsService.publish()` freezes each question into `TestQuestion.publishedSnapshot` (canonical `buildContentSnapshot`), atomic with the status flip via `$transaction`.
- `TestAttemptsService.startOrResumeAttempt()` builds `contentSnapshot` from `publishedSnapshot` (via `deserializeSnapshot`) when present, else from the live question (draft/practice fallback preserved).
- Reuses the single snapshot serializer — publish-time and attempt-time snapshots cannot drift.

**Files Changed:**
- `backend/src/tests/tests.service.ts` — [publish() writes publishedSnapshot per TestQuestion in a transaction]
- `backend/src/tests/test-attempts.service.ts` — [attempt creation prefers publishedSnapshot over live question]
- `backend/src/tests/tests.service.spec.ts` — [publish freezes snapshot]
- `backend/src/tests/test-attempts.service.spec.ts` — [attempt uses frozen snapshot, ignores live edits]

**Tests Added:**
- 2 unit tests (publish-freezes-snapshot; attempt-reads-frozen-snapshot). Full suite **721/721** passing (0 regressions).

**Migration:** none (uses `TestQuestion.publishedSnapshot` from TICKET-001).

---

## 2026-07-04 — TICKET-001: Schema Migrations + Backfill

**Implemented:**
- Applied 3 schema migrations containing additions to Question, TestQuestion, Note, Profile, TestSeries and new TestCollection, QuestionStats, ImportHistory models.
- Executed one-time backfill script to populate QuestionStats for 129 existing questions.
- Excluded playwright directory from frontend tsconfig to fix compilation.
- Fixed TypeScript compile errors on backend related to nullable productId in TestSeries.

**Files Changed:**
- `backend/prisma/schema.prisma` — [Schema additions]
- `backend/src/auth/guards/test-access.guard.ts` — [Null check for productId]
- `backend/src/auth/guards/test-series-ownership.guard.ts` — [Ownership check via Collection creator fallback]
- `backend/src/tests/tests.service.ts` — [Nullable check for productId]
- `frontend/tsconfig.json` — [Exclude playwright from typecheck]

**Tests Added:**
- All 720 Jest tests verified passing (0 regressions).

**Migration:** `20260703223308_phase_3_1_8_exam_ready`

---

<!-- Future entries follow this format:

## YYYY-MM-DD — TICKET-XXX: [Title]

**Implemented:**
- Brief description of what was built

**Files Changed:**
- `backend/src/path/to/file.ts` — [what changed]
- `frontend/src/path/to/file.tsx` — [what changed]

**Tests Added:**
- `backend/src/path/to/file.spec.ts` — [N tests]
- `frontend/playwright/acceptance/xxx.spec.ts` — [N tests]

**Migration:** [migration file name, if any]

-->

