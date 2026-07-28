# EXAM-P6-T02 — Verification addendum

Status: VERIFICATION HOLD (implementation accepted; final approval withheld pending this addendum review)  
Date: 2026-07-28

## Standards loaded

AGENTS.md, CLAUDE.md, docs/standards/index.md (+ testing/frontend/security). No conflict with frozen T02 scope. EXAM-P6-T03 remains **not authorised**.

## Confirmation: no T03 code started

No timer, expiry, auto-submit, final-submission UX, scoring/results UI, or review work was begun. Changes in this addendum are diagnosis, integration tests, and evidence only.

---

## 1. Failure classification (all 17 failing suites)

**Method**

1. Re-ran the exact 17 failing files **with T02 present** → `17 failed` files, `34 failed | 18 passed` tests.
2. Quarantined the entire attempt surface (T01+T02 player/composables/pages/tests) by moving those paths out of `app/`.
3. Re-ran the **same 17 files** → identical summary: `17 failed` files, `34 failed | 18 passed` tests.
4. Grep of failing test files for `attempt/`, `attempt_player`, `attempt_autosave`, `assessment_attempts`, `AttemptPlayer` → **zero matches**.
5. Restored quarantined paths after the baseline run.

**Verdict:** T02 introduced **no new failing suites**. All 17 failures are pre-existing relative to the attempt-player surface.

### Classification table

| # | Test file / suite | Failure message (representative) | Affected feature | Baseline without T01/T02 attempt surface | With T02 | Why unrelated to T02 |
|---|---|---|---|---|---|---|
| 1 | `test/components/Reports/PageLayout.test.js` | Empty DOM / expected 2 nav tabs, got 0 | Live quiz analysis tabs (`QuizAnalysisTabs`) | FAIL (same) | FAIL | Selector/markup drift in reports UI; no attempt imports |
| 2 | `test/components/WinnerCard.test.js` | Empty DOM for medal/avatar; casing assertion | Live scoreboard winner card | FAIL (same) | FAIL | Live quiz scoreboard; no attempt imports |
| 3 | `test/components/Option.test.js` | Missing order/media/admin badge nodes | Live `Option.vue` (Vuetify-era markup) | FAIL (same) | FAIL | Player uses `AttemptQuestionPanel`, not this Option test path; Option.vue not modified by T02 |
| 4 | `test/components/Utils/ConfirmModal.test.js` | `#confirmModal` missing | Shared confirm modal markup | FAIL (same) | FAIL | Modal DOM contract stale; T02 has no ConfirmModal usage |
| 5 | `test/components/ListJoinUser.test.js` | chip count 0; ellipsis text mismatch (`..` vs `...`) | Live waiting-room participant chips | FAIL (same) | FAIL | Live join list; no attempt imports |
| 6 | `test/components/ScoreBoardTable.test.js` | row/chip selectors miss; avatar URL mismatch | Live scoreboard table | FAIL (same) | FAIL | Live scoreboard; no attempt imports |
| 7 | `test/components/Pagination.test.js` | empty page label / undefined button classes | Shared pagination control | FAIL (same) | FAIL | Pagination markup drift; not used by attempt player |
| 8 | `test/components/Quiz/ListUserAnswered.test.js` | expected nodes missing | Live answered-user chips | FAIL (same) | FAIL | Live quiz host UI |
| 9 | `test/components/Quiz/OptionsAnalysis.test.js` | options / correctness styles missing | Live options analysis | FAIL (same) | FAIL | Host analysis UI; correctness styling out of player scope |
| 10 | `test/components/Quiz/UserAnalyticsSpace.test.js` | mount/avatar assertions fail | Live user analytics space | FAIL (same) | FAIL | Live analytics UI |
| 11 | `test/components/Quiz/WaitingSpace.test.js` | render assertion fail (1 of 10) | Live waiting room | FAIL (same) | FAIL | Live waiting room |
| 12 | `test/components/FinalScoreBoard.test.js` | `ctx.setTransform` null in `WinnerConfetti.vue` | Live final scoreboard + canvas confetti | FAIL (same) | FAIL | Canvas API under happy-dom; live scoreboard only |
| 13 | `tests/e2e/learning-item-e2e.spec.ts` | Playwright `test.describe.configure()` unexpected under Vitest | Course learning-item E2E (Playwright) | FAIL (same) | FAIL | Vitest incorrectly collecting Playwright specs; not attempt player |
| 14 | `tests/e2e/runtime-verification.spec.ts` | Same Playwright-under-Vitest configure error | Runtime verification E2E | FAIL (same) | FAIL | Same Vitest/Playwright mismatch |
| 15 | `test/components/Quiz/QuestionAnalysis.test.js` | Cannot resolve `vuetify/components` | Live question analysis | FAIL (same) | FAIL | Legacy Vuetify import; Vuetify not a project dependency |
| 16 | `test/components/Quiz/QuestionSpace.test.js` | Cannot resolve `vuetify/components` | Live question space | FAIL (same) | FAIL | Legacy Vuetify import |
| 17 | `test/components/Quiz/SocreSpace.test.js` | Cannot resolve `vuetify/components` | Live score space | FAIL (same) | FAIL | Legacy Vuetify import |

### Baseline / regression comparison summary

| Run | Command scope | Result |
|---|---|---|
| A | 17 suites **with** attempt surface | 17 failed files / 34 failed tests / 18 passed |
| B | Same 17 suites **without** attempt surface (quarantined) | **Identical** 17 / 34 / 18 |
| C | Import coupling | Failing suites do not reference T02 modules |

**Regression conclusion:** No newly failing suites attributable to EXAM-P6-T02.

---

## 2. Lint-hang diagnosis

| Item | Value |
|---|---|
| Exact command | `npm run lint` → `eslint --ext ".ts,.js,.vue" --ignore-path .gitignore --max-warnings 0 .` |
| Node | `v24.16.0` |
| npm | `11.16.0` |
| Elapsed before termination (prior T02 run) | ~10+ minutes with **no eslint problem output** |
| Last observable activity | Process alive; CPU/IO continuing; no lint results emitted |
| Root cause | `app/playwright-report/` is **not** in `.gitignore` and contains minified generated JS (e.g. `defaultSettingsView-*.js` ~643 KB, `codeMirrorModule-*.js` ~306 KB, `sw.bundle.js` ~94 KB). Full-repo `.` scan includes these artefacts. |
| Direct proof | `npx eslint … playwright-report` **hung >60s** and was terminated |
| Bounded source tree | `components pages composables utils plugins store layouts lib config` completed in **~7.3s** |
| Intended source + tests | Completes in **~7.5s** but reports **pre-existing prettier** issues in unrelated files (`EditQuestion.test.js`, `CourseLearningItemsApi.test.js`, `quiz_questions.test.js`) — **not** introduced by T02; lint rules were **not** changed |
| T02-scoped eslint | PASS (`--max-warnings 0`) |

**Recommendation (not applied):** add `playwright-report/` and `test-results/` to `.gitignore` (or an `.eslintignore`) so `npm run lint` does not parse Playwright HTML report bundles. Deferred — do not alter ignore/lint config under this verification-only authorisation unless separately approved.

---

## 3. Additional integration coverage

New file: `app/test/components/attempt/AttemptPlayer.integration.test.js` (8 tests)

Covers:

| Scenario | Result |
|---|---|
| Resume → select → autosave → palette `answered` only after success | PASS |
| Remount/reload restores saved answer | PASS |
| Save failure → retry preserves draft | PASS |
| Navigate while save pending retains draft | PASS |
| Stale first response ignored after newer selection (serial queue + version) | PASS |
| Clear → remount without restored selection | PASS |
| Terminal-attempt autosave rejection surfaces alert | PASS |
| No answer-key fields in resume-derived player state | PASS |

Combined with prior T02 suites: **26 passed**.

---

## 4. T01 smoke-test status (phase-level)

Live Compose browser smoke for EXAM-P6 remains an **unresolved phase-level verification item**. It may stay deferred until Phase 6 closure, but the ledger must continue to identify it as open (see ledger update).

---

## Stop condition

Verification addendum complete. **EXAM-P6-T03 not started.** Awaiting manual review for final T02 approval.
