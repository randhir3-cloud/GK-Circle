# Architectural Decisions — Phase 3.1.8

> Future agents: these decisions are **closed**. Do not reopen without explicit project owner approval.

---

## 2026-07-04 — Decision: No QuestionVersion Table

**Decision:** Do not implement a `QuestionVersion` table in Phase 3.1.8.

**Reason:**
The existing `TestAttemptQuestion.contentSnapshot` (JSON freeze at attempt start) already protects in-progress and completed attempts. The only gap — edits between test publish time and attempt start — is closed by `TestQuestion.publishedSnapshot` (TICKET-002), a single additive field that freezes question content at publish time. This covers 100% of the integrity requirement without the complexity of a versioning table, backfill scripts, or extra joins.

**What's deferred:** A full audit trail with diff views and "restore to version N" capability. This goes to Phase 4.

**Approved by:** Project Owner (2026-07-04)

---

## 2026-07-04 — Decision: Community Model Reused for Friend Groups

**Decision:** Friend Groups (TICKET-037) reuse the existing `Community` + `CommunityMember` models. No new "FriendGroup" or "StudyGroup" table.

**Reason:** The `Community` model already has everything needed: members, join policy, visibility, owner. Adding a separate model would be duplication. The distinction between "Study Circle" and "Friend Group" is purely a UI label and permission scope.

**Implementation:** New API endpoints filter Community members when computing test leaderboards and listing shared tests.

**Approved by:** Project Owner (2026-07-04)

---

## 2026-07-04 — Decision: Analytics Are Computed, Not Stored (for Sprint 2)

**Decision:** All Sprint 2 analytics (Syllabus Coverage, Mock Readiness Score, Revision Cycle, Most Forgotten, Time Wasted, Confidence vs Reality) are **computed on-demand from existing tables**. No new analytics tables are created.

**Reason:** The schema already has `UserQuestionHistory` (append-only event log), `UserQuestionMastery` (rollup), and `MCQAnswer` (with confidence, time). Computing from these eliminates synchronization complexity. If query performance becomes an issue post-exam, materialized views can be added in Phase 4.

**Exception:** `QuestionStats` (TICKET-039, Sprint 4) IS materialized because it's aggregated globally across all users, not per-user.

**Approved by:** Project Owner (2026-07-04)

---

## 2026-07-04 — Decision: Exam Simulation Mode = Code Constants, Not DB Records

**Decision:** PPSC Prelim, UPSC Prelim, and other simulation presets are defined as TypeScript constants in `TestsService`, not stored as database records.

**Reason:** These presets rarely change and don't need CRUD. They're behavioral configurations that belong in code, not data. Adding a `SimulationPreset` table would be over-engineering.

**Approved by:** Project Owner (2026-07-04)

---

## 2026-07-04 — Decision: Test Templates = Code Constants

**Decision:** Test templates (PPSC Prelim, UPSC Prelim, Daily Quiz, etc.) are TypeScript constants returned by `GET /api/v1/tests/templates`. No `TestTemplate` table.

**Same reasoning as Exam Simulation Mode.**

**Approved by:** Project Owner (2026-07-04)

---

## 2026-07-04 — Decision: Only 3 Schema Migrations in Phase 3.1.8

**Decision:** Exactly 3 additive migrations. No other schema changes are permitted in this phase.

```
m1_questions   — Question, TestQuestion, Note, Profile additions
m2_collections — TestCollection, QuestionStats, TestSeries.collectionId
m3_operations  — ImportHistory
```

Any feature that would require a 4th migration must either (a) be computed from existing data or (b) be deferred to Phase 4.

**Approved by:** Project Owner (2026-07-04)

---

## 2026-07-04 — Decision: Confidence UI Timing

**Decision:**
- In **Practice mode**: confidence selector shown BEFORE answer reveal (forces genuine self-assessment)
- In **Exam mode**: confidence selector shown at answer selection time (no reveal during exam anyway)
- Confidence is **optional** — student can skip without impact

**Approved by:** Project Owner (2026-07-04)

---

## 2026-07-04 — Decision: Daily Practice Size = Adaptive (Not Fixed)

**Decision:** Daily Practice question count adapts to user's available time input (15/30/60 min) using `estimatedTimeSecs` per question (default 90s). Not fixed at 20 questions.

**Approved by:** Project Owner (2026-07-04)

---

## 2026-07-04 — Decision: Publish Snapshot Implementation Pattern

**Decision:** When `TestsService.publish(testId)` is called, the service iterates all `TestQuestion` records for that test and writes `publishedSnapshot = { questionText, options, explanation, imageUrl }` from the live `Question` at that moment. When `TestAttemptsService.create()` runs, it reads from `TestQuestion.publishedSnapshot` (not the live `Question`) to build `TestAttemptQuestion.contentSnapshot`.

**This is the authoritative pattern. Do not deviate.**

**Approved by:** Project Owner (2026-07-04)

---

## 2026-07-04 — ADR-001: Question Import Moved into ImportModule

**Decision:** Refactor all bulk question import controller endpoints, validation routines, header parsing, mapping suggestions, and templates streaming into a dedicated `ImportModule` submodule (`backend/src/questions/import/`) under the route prefix `/questions/import`.

**Reason:**
The bulk question import feature is growing rapidly (to support custom header mapping wizard steps, profiles serialization, safety validations, rollback endpoints, and import history logs). Placing everything inside the single `QuestionsController` and `QuestionsService` would lead to an excessively large, non-modular, and difficult to maintain class. Isolating the logic into an `ImportModule` with dedicated services (e.g. `ValidationService`, `HeaderParserService`, `MappingService`, `TemplateService`) ensures clean domain separation, robust testing hooks, and zero cross-cutting code merge conflicts for separate agents.

**Consequences:**
- All import controllers and services are organized in subdirectories under `backend/src/questions/import/`.
- Routes are migrated to `/questions/import/validate`, `/questions/import/parse-headers`, `/questions/import/template`, and `/questions/import`.
- Deprecated routing wrappers are temporarily retained in the main `QuestionsController` to preserve backward compatibility.

**Approved by:** Project Owner (2026-07-04)


