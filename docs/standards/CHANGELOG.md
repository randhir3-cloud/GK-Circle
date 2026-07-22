# GK Circle Standards Changelog

Central audit log for **governance and standards** changes. Individual standard files may carry inline version notes (e.g. v1.1); this file is the authoritative history of *why* rules changed.

Format for every entry:

| Field | Description |
|-------|-------------|
| **Version** | Governance release tag (standards bundle) |
| **Date** | ISO date |
| **Files Changed** | Paths touched |
| **Reason** | Why the change was made |
| **Breaking Changes** | YES / NO — what agents or workflows must do differently |
| **Migration Required** | YES / NO — concrete steps if yes |

---

## [2.5] - 2026-07-14 - Subject-level Course Content and hierarchical Course Tests

| Field | Value |
|-------|-------|
| **Files Changed** | `AGENTS.md`, `docs/standards/course-rules.md`, `docs/standards/CHANGELOG.md`, `docs/architecture/ADR/ADR-022-course-content-and-hierarchical-tests.md` |
| **Reason** | Approve ADR-022 and allow Course Content and Course Tests directly under a CourseSubject or under one of its CourseTopics while preserving the existing content, question, and assessment engines. |
| **Breaking Changes** | **YES** — Course hierarchy resolvers and learner queries must support an optional CourseTopic for Course Content and Course Tests; topic-level question ancestry remains exact. |
| **Migration Required** | **YES** — Migration Gate 1 authoring is approved; database execution remains prohibited until a target-specific Gate 2 approval. |

---

## [2.2] — 2026-07-12 — Governance freeze (maintainer constitution)

| Field | Value |
|-------|-------|
| **Files Changed** | `AGENTS.md`, `CLAUDE.md`, `docs/standards/index.md`, `docs/standards/testing-rules.md`, `docs/standards/security-rules.md`, `docs/standards/architecture-rules.md`, `docs/standards/ai-rules.md`, `docs/standards/documentation-rules.md`, `docs/standards/operations-rules.md`, `docs/standards/documentation-governance.md`, `docs/testing/qa-account-governance.md`, `docs/standards/CHANGELOG.md` (this file), `docs/architecture/ADR/README.md`, `docs/architecture/ADR/ADR-TEMPLATE.md` |
| **Reason** | Freeze cross-agent governance so Cursor, Claude, Copilot, Gemini behave consistently; add maintainer role, stop conditions, breaking-change disclosure, evidence conventions, ADR process, production verification gate |
| **Breaking Changes** | **YES** — Agents must confirm standards loaded before implementation; must declare breaking changes explicitly; evidence must use prescribed folder layout; commits must be single-purpose |
| **Migration Required** | **YES** — (1) Read `AGENTS.md` §AI Maintainer Constitution and §Before Coding Checklist. (2) Record new architectural decisions in `docs/architecture/ADR/`. (3) Move scattered bug-bash/release evidence into `docs/bug-bash/`, `docs/releases/`, or `docs/evidence/` as applicable |

---

## [2.1] — 2026-07-12 — RC-1 QA governance documentation

| Field | Value |
|-------|-------|
| **Files Changed** | `AGENTS.md`, `CLAUDE.md`, `docs/standards/index.md`, `docs/standards/testing-rules.md`, `docs/standards/security-rules.md`, `docs/standards/architecture-rules.md`, `docs/standards/ai-rules.md`, `docs/standards/documentation-rules.md`, `docs/testing/qa-account-governance.md` |
| **Reason** | Document immutable four-account QA system, single password source (`backend/.env`), Playwright preconditions, repository reuse, bug-bash discipline |
| **Breaking Changes** | **YES** — No legacy QA emails; no hardcoded QA passwords; Playwright must use `getQaUser()` / approved helpers |
| **Migration Required** | **YES** — Run `backend/scripts/qa/verify-qa-accounts.ts`; retire legacy seed scripts; see `docs/testing/qa-governance-migration-report.md` |

---

## [2.0] — 2026-07-09 — Standards index v2.0

| Field | Value |
|-------|-------|
| **Files Changed** | `docs/standards/index.md`, domain standards under `docs/standards/` |
| **Reason** | Consolidated mandatory reading order and execution pipeline for all agents |
| **Breaking Changes** | **NO** |
| **Migration Required** | **NO** |

---

## [2.3] - 2026-07-13 - Module certification standard

| Field | Value |
|-------|-------|
| **Files Changed** | `docs/standards/module-certification-standard.md`, `docs/standards/CHANGELOG.md` |
| **Reason** | Add reusable local module certification SOP covering entry criteria, root-cause verification, defect handling, shared-component regression, evidence, and exit criteria |
| **Breaking Changes** | **NO** |
| **Migration Required** | **NO** |

---

## [2.4] - 2026-07-13 - Course terminology standards alignment

| Field | Value |
|-------|-------|
| **Files Changed** | `docs/standards/index.md`, `docs/standards/course-rules.md`, `docs/standards/architecture-rules.md`, `docs/standards/backend-rules.md`, `docs/standards/admin-panel-rules.md`, `docs/standards/security-rules.md`, `docs/standards/testing-rules.md`, `docs/standards/live-exam-rules.md`, `docs/standards/creator-economy-rules.md`, `docs/standards/frontend-rules.md`, `docs/standards/rag-rules.md`, `docs/standards/devops-rules.md`, `docs/standards/operations-rules.md`, `docs/standards/documentation-rules.md`, `docs/standards/documentation-folder-rules.md`, `docs/standards/CHANGELOG.md` |
| **Reason** | Retire former educational catalog terminology from current standards after ADR-019 and make Course terminology authoritative before PHASE-C-002. |
| **Breaking Changes** | **YES** — agents must use Course terminology, Course routes, Course services, and `course:*` permissions for the active educational domain. |
| **Migration Required** | **NO** — documentation-only checkpoint; database permission terminology migration is tracked separately under the nested curriculum evidence package. |

---

## How to add an entry

When changing any file under `docs/standards/`, `AGENTS.md`, or `CLAUDE.md` governance sections:

1. Bump the **Governance Version** in `AGENTS.md` if the change is agent-facing or mandatory.
2. Add a row to this changelog **before** merging.
3. If the decision is architectural, add an ADR under `docs/architecture/ADR/`.
