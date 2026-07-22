# GK Circle v1.0 — Exam Ready Mode
## MASTER PLAN — Phase 3.1.8

> **Source of Truth.** This document does not change except on formal scope changes.  
> **Last Scope Update:** 2026-07-04  
> **Approved By:** Project Owner  

---

## Overall Goal

Transform GK Circle into the best personal PCS/UPSC exam preparation platform — covering the complete lifecycle:

```
Create Questions → Build Tests → Simulate Exams → Analyze Performance
→ Track Syllabus Coverage → Generate Revision → Study Planning → Share & Compare
```

---

## Architecture Constraints (Non-Negotiable)

1. **No QuestionVersion table** — Publish Snapshot (`TestQuestion.publishedSnapshot`) solves historical integrity. QuestionVersion deferred to Phase 4.
2. **No destructive migrations** — All schema changes are additive only.
3. **Preserve all existing modules** — Extend, never rewrite.
4. **3 schema migrations only** — `m1_questions`, `m2_collections`, `m3_operations`.
5. **Standards compliance required** — All files in `docs/standards/` must be respected.

---

## Sprint Structure

| Sprint | Theme | Tickets | Estimated Days |
|---|---|---|---|
| **Sprint 1** | Core Exam Engine | TICKET-001 → TICKET-016 | 8–10 days |
| **Sprint 2** | Preparation Intelligence | TICKET-017 → TICKET-032 | 6–7 days |
| **Sprint 3** | Collaboration | TICKET-033 → TICKET-038 | 3 days |
| **Sprint 4** | Polish & Quality | TICKET-039 → TICKET-050 | 3–4 days |

---

## All Tickets

### Sprint 1 — Core Exam Engine

| Ticket | Title | Est. | Dependencies |
|---|---|---|---|
| TICKET-001 | Schema Migrations + Backfill | 0.5d | — |
| TICKET-002 | Publish Snapshot (Test → Freeze Questions) | 0.5d | TICKET-001 |
| TICKET-003 | Question Import — Backend (Parse, Validate, Deduplicate) | 1d | TICKET-001 |
| TICKET-004 | Question Import — Frontend (Upload, Mapping, Validation UI) | 1d | TICKET-003 |
| TICKET-005 | Import History + Rollback (Undo Import) | 0.5d | TICKET-003 |
| TICKET-006 | Image Support — Backend (Upload endpoint) | 0.5d | TICKET-001 |
| TICKET-007 | Image Support — Frontend (Editor + Exam/Review renderers) | 1d | TICKET-006 |
| TICKET-008 | PYQ Support + Sources UI | 0.5d | TICKET-001 |
| TICKET-009 | Question Locking Warning | 0.5d | — |
| TICKET-010 | Test Builder — Rich Filters (Never Used, Most Wrong, etc.) | 1d | — |
| TICKET-011 | Test Builder — Random Distribution Builder | 0.5d | TICKET-010 |
| TICKET-012 | Exam Simulation Mode (One-Click PPSC/UPSC) | 0.5d | — |
| TICKET-013 | Test Templates (Code constants + picker UI) | 0.5d | — |
| TICKET-014 | Test Cloning | 0.5d | — |
| TICKET-015 | Exam UI — Numbered Palette + Mark for Review + Timer | 2d | — |
| TICKET-016 | Autosave + Resume Later | 1d | — |

### Sprint 2 — Preparation Intelligence

| Ticket | Title | Est. | Dependencies |
|---|---|---|---|
| TICKET-017 | Performance Dashboard — Backend API (`/me/performance`) | 1d | — |
| TICKET-018 | Performance Dashboard — Frontend UI | 1d | TICKET-017 |
| TICKET-019 | Syllabus Coverage Tracker — Backend API | 1d | — |
| TICKET-020 | Syllabus Coverage Tracker — Frontend (Progress bars + Heatmap) | 0.5d | TICKET-019 |
| TICKET-021 | Mock Test Readiness Score | 1d | TICKET-017, TICKET-019 |
| TICKET-022 | Progress Targets + Exam Countdown Calendar | 0.5d | TICKET-001 |
| TICKET-023 | Daily Streak (Increment logic + Dashboard display) | 0.5d | — |
| TICKET-024 | Revision Cycle (First / Second / Final Revision lanes) | 0.5d | — |
| TICKET-025 | Smart Revision Queue + Most Forgotten Questions | 0.5d | — |
| TICKET-026 | Adaptive Daily Practice Generator | 1d | TICKET-025 |
| TICKET-027 | Study Planner (Exam countdown-driven daily plan) | 0.5d | TICKET-019, TICKET-026 |
| TICKET-028 | Confidence Rating UI | 0.5d | — |
| TICKET-029 | Confidence vs Reality + Time Wasted Analytics | 0.5d | TICKET-028 |
| TICKET-030 | Bookmarks During Exam + Practice Bookmarks | 0.5d | — |
| TICKET-031 | Personal Notes on Questions | 0.5d | TICKET-001 |
| TICKET-032 | Attempt History Per Question (Display UI) | 0.25d | — |

### Sprint 3 — Collaboration

| Ticket | Title | Est. | Dependencies |
|---|---|---|---|
| TICKET-033 | Test Sharing (Share link UI + `/tests/share/[token]` route) | 0.5d | — |
| TICKET-034 | Test Leaderboard | 0.5d | — |
| TICKET-035 | Friend Comparison (Side-by-side results) | 0.5d | TICKET-034 |
| TICKET-036 | Test Collections (Collection → Series → Tests hierarchy) | 1d | TICKET-001 |
| TICKET-037 | Friend Groups — Shared Tests + Group Leaderboard | 0.5d | TICKET-034 |
| TICKET-038 | Create Test From Mistakes (Results → new test) | 0.5d | — |

### Sprint 4 — Polish & Quality

| Ticket | Title | Est. | Dependencies |
|---|---|---|---|
| TICKET-039 | Question Statistics (Auto-maintained via QuestionStats) | 1d | TICKET-001 |
| TICKET-040 | Rich Question Metadata UI (Bloom, Importance, Est. Time) | 0.5d | TICKET-001 |
| TICKET-041 | Printable Export — PDF Question Paper + Answer Key | 1d | — |
| TICKET-042 | Printable Export — OMR Sheet + DOCX | 0.5d | TICKET-041 |
| TICKET-043 | Import History UI + Question Audit Views | 0.5d | TICKET-005 |
| TICKET-044 | Playwright Tests — Sprint 1 Full Suite | 1d | Sprint 1 done |
| TICKET-045 | Playwright Tests — Sprint 2 Full Suite | 1d | Sprint 2 done |
| TICKET-046 | Playwright Tests — Sprint 3 Full Suite | 0.5d | Sprint 3 done |
| TICKET-047 | Bulk Question Actions (Publish/Archive selection) | 0.5d | — |
| TICKET-048 | Subject/Topic Quick-Create from Question Form | 0.5d | — |
| TICKET-049 | Creator Question Preview (Exam-style modal) | 0.5d | — |
| TICKET-050 | Performance Optimization + Index Audit | 0.5d | All done |

---

## Schema Changes Summary

| Migration | File | Changes |
|---|---|---|
| m1_questions | `20260704_phase_3_1_8_m1_questions` | +13 fields to Question, +2 to TestQuestion, +1 to Note, +3 to Profile |
| m2_collections | `20260704_phase_3_1_8_m2_collections` | New: TestCollection, QuestionStats; Modified: TestSeries (+collectionId) |
| m3_operations | `20260704_phase_3_1_8_m3_operations` | New: ImportHistory |

---

## Milestones

| Milestone | Completion Criteria |
|---|---|
| **M1: Exam Engine Live** | Sprint 1 complete. Creator can import questions, build tests, take full exams with palette and autosave. |
| **M2: Preparation Intelligence Live** | Sprint 2 complete. Dashboard shows coverage, readiness, study plan, revision cycle. |
| **M3: Platform Complete** | Sprints 3+4 complete. Sharing, leaderboards, exports all working. |
| **FINAL: v1.0 Exam Ready** | All 50 tickets completed. All Playwright tests passing. All documentation updated. |

---

## Completion Criteria for v1.0

- [ ] All 50 tickets marked 🟢 Completed
- [ ] `npx jest --runInBand` passes ≥ 711 tests (no regressions)
- [ ] `npx tsc --noEmit` passes on backend and frontend
- [ ] `npm run build` completes on frontend
- [ ] All 15 Playwright spec files passing
- [ ] CHANGELOG.md documents every completed ticket
- [ ] `docs/active-context/latest-session.md` updated
- [ ] `docs/active-context/active-plan.md` updated to Phase 3.1.9

