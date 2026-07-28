# Full Test Failure Repair — Evidence Package

Path: `docs/features/exam-platform/evidence/exam-p6-t02/repair-verification/`  
Date: 2026-07-28  
Task: EXAM PLATFORM — FULL TEST FAILURE REPAIR (no EXAM-P6-T03)

## Final status

**FAILED** — mandatory automated suites pass; mandatory real-browser attempt autosave cycle did not complete due to database migration ledger mismatch after API image update.

## Environment

| Item | Value |
|---|---|
| OS | Microsoft Windows NT 10.0.26200.0 |
| Node | v24.16.0 |
| npm | 11.16.0 |
| Go | go1.26.5 windows/amd64 |
| CWD | `E:\GK Circle v2` |
| Branch | `chore/ci-verification` |
| Commit | `eeac599f05eaf936c7f61db4a3deeac3c9063f59` |

## Standards loaded

AGENTS.md, CLAUDE.md, docs/standards/index.md (+ testing/frontend/security). Stop condition: EXAM-P6-T03 not authorised.

## Initial baseline

| Command | Result |
|---|---|
| `go test ./... -count=1` | PASS |
| `go build ./...` | PASS |
| `npm test -- --run` | 17 failed files / 34 failed tests / 161 passed |
| `npm run lint` | historically hung |
| `npm run build` | PASS (prior) |

## Complete failure table (baseline Vitest)

See parent `verification-addendum.md` for the 17-suite classification. Root class: stale Bootstrap/Vuetify selectors; canvas null context; Playwright specs collected by Vitest; missing `vuetify` dependency; lint hang on `playwright-report`.

## Root causes and fixes

| Failure class | Root cause | Fix |
|---|---|---|
| Markup/selector drift | Components redesigned to Tailwind/lucide; tests still used Bootstrap/Vuetify selectors | Rewrote tests to assert current semantic UI |
| Canvas / FinalScoreBoard | `getContext('2d')` null in happy-dom → `ctx.setTransform` crash | Production guard when `ctx` is null; stub confetti in FinalScoreBoard tests |
| Playwright under Vitest | `tests/e2e/**` collected by Vitest | Exclude `tests/e2e` in `vitest.config.ts`; run via Playwright |
| Vuetify imports | Tests imported removed dependency | Removed imports; assert real ScoreSpace/QuestionSpace/QuestionAnalysis markup |
| ESLint hang | `playwright-report/` minified JS not ignored | Added generated artefacts to `app/.gitignore` (eslint uses ignore-path) |
| Playwright browsers | Chromium not installed | `npx playwright install chromium` |
| Playwright env | Missing `.env.e2e.local` vars | Load documented local E2E env |

## Files changed (repair)

### Production
- `app/components/WinnerConfetti.vue` — null canvas context guard
- `app/.gitignore` — ignore playwright-report, test-results, coverage
- `app/vitest.config.ts` — exclude Playwright e2e from Vitest

### Tests (stale → current behaviour)
- Option, ConfirmModal, Pagination, PageLayout, WinnerCard, ListJoinUser, ScoreBoardTable
- Quiz: ListUserAnswered, OptionsAnalysis, UserAnalyticsSpace, WaitingSpace, QuestionAnalysis, QuestionSpace, SocreSpace
- FinalScoreBoard
- Prettier autofix: EditQuestion.test.js, CourseLearningItemsApi.test.js, quiz_questions.test.js, Option.test.js

### Smoke helpers (local)
- `app/tests/e2e/repair-browser-smoke.mjs`
- `app/tests/e2e/repair-smoke-attempt.mjs`
- `app/tests/e2e/repair-smoke-player.mjs`

## Final automated command results (clean repeat)

| Command | Result |
|---|---|
| `go test ./... -count=1` | **PASS** |
| `go build ./...` | **PASS** |
| `npm run lint` | **PASS** (EXIT 0, ~8s) |
| `npm test -- --run` | **PASS** — 45 files / 222 tests |
| T02 suites + integration | **PASS** — 26 tests |
| `npm run build` | **PASS** |
| `npx playwright test` (with `.env.e2e.local`) | **PASS** — 11/11 |
| Repeat Playwright | **PASS** — 11/11 |

## Real-browser smoke

| Step | Result |
|---|---|
| Login | PASS |
| Open quiz list | PASS |
| Open reports | PASS |
| Attempt instructions route (authz) | PASS — entitlement rejection for foreign IDs; no page errors |
| Create quiz for autosave smoke | PASS (201) |
| Add question / create snapshot / start attempt / autosave / refresh | **FAILED** |

Blocker detail for autosave cycle:

```text
ADD_QUESTION=500
pq: column "answer_review_status" of relation "questions" does not exist
```

`docker compose run --rm migration` then reported:

```text
Unable to create migration plan because of
20260727100000_create_course_enrollments_table.down.sql: unknown migration in database
```

So the rebuilt API expects schema the local DB migration ledger cannot currently plan. This is a repository/environment migration drift issue, not a skipped test.

## Remaining failures

1. Mandatory full attempt player smoke (select → autosave → reload restore → palette) — blocked by migration ledger mismatch.
2. No EXAM-P6-T03 work was started.

## Confirmations

- No tests skipped, disabled, or weakened for green results.
- No hardcoded bypasses or dummy production credentials committed.
- EXAM-P6-T03 not started.
- Vitest no longer executes Playwright specs; Playwright executes them and they pass.

## Production impact

Breaking Changes: **NO** (repair + defensive canvas guard + ignore generated lint paths)  
Migration status: **PENDING / BLOCKED** on local Compose DB ledger conflict after API image refresh.
