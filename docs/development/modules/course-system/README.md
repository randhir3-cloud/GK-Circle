# Universal Chained Course System — Development Tracker

This directory is the persistent implementation record for the Universal Chained Course System.

## Start here

Every implementation agent must read these files in order:

1. `../../PROJECT_INDEX.md`
2. `../../SYSTEM_ARCHITECTURE.md`
3. `MASTER_PLAN.md`
4. `ROADMAP.md` (implementation-order index; task tables stay in `phases/`)
5. `DOCUMENTATION_FREEZE.md` (freeze certificate; DOC-CS-T04)
6. `CURRENT_STATUS.md`
7. `HANDOFF.md`
8. The active phase document under `phases/`
9. `DECISIONS.md` (module summaries; **D-004**, **D-005** required)
10. `architecture/current.md` (recursive tree + Relationship Contract)
11. `../../../architecture/ADR/ADR-023-canonical-course-hierarchy-model.md`
12. `../../../api/course-learning-items-v1.md` (canonical LearningItem API examples)

Canonical ADRs live under `docs/architecture/ADR/`. Module decision files are
summaries and must not become competing authoritative copies of ADR-023.

**Mandatory agent rule:** Before implementing any Course hierarchy task, read
`architecture/current.md`, **D-004**, **D-005**, ADR-023, the relevant phase
file, and all dependency tasks.

## Aggregate model (frozen)

```text
Course
└── CourseNode
    ├── parent_id → CourseNode | NULL   # AUTHORITATIVE structural edge
    ├── children  → derived WHERE parent_id = id   # PROJECTION only
    └── LearningItem*                  # ordered content on any node
```

## Course Node Relationship Contract (summary)

Normative architecture lives in **one** place — do not maintain a second independent copy here.

* **Architecture authority:** [`architecture/current.md`](architecture/current.md) (Relationship Contract + three-graph model)
* **Decision authority:** [`DECISIONS.md`](DECISIONS.md) (**D-004** Unlimited Logical Course Tree Depth; **D-005** CourseNode Relationship Contract)
* **ADR authority:** [`ADR-023`](../../../architecture/ADR/ADR-023-canonical-course-hierarchy-model.md) (body not forked here)
* **Ownership authority:** [`ROADMAP.md`](ROADMAP.md)
* **Next-action authority:** [`HANDOFF.md`](HANDOFF.md)

Frozen summary: `CourseNode.parent_id` is authoritative; `children[]` / breadcrumbs / ancestors are derived projections; unlimited logical depth; no product `MAX_*_DEPTH`; no authoritative `child_node_ids[]`; CourseNode.`status` ≠ LearningItem.`publish_state`; structure ≠ learning sequence ≠ prerequisite DAG.

## Source of truth

`CURRENT_STATUS.md` records:

* active phase;
* active work item;
* overall completion;
* phase completion;
* completed work;
* remaining work;
* blockers;
* verification state;
* next safe action.

`status.json` is the authoritative machine-readable numeric state.

Conversation history is not an authoritative implementation record.

## Status values

* `NOT_STARTED`
* `IN_PROGRESS`
* `BLOCKED`
* `IMPLEMENTED`
* `VERIFIED`
* `DEFERRED`
* `SUPERSEDED`

A task is counted as complete only when its status is `VERIFIED` in the corresponding phase file under `phases/`.

Ledger checklist columns (Evidence at columns[7]):

`| ID | Task | Points | Status | Difficulty | Est. Time | Evidence | Dependencies |`

## Agent protocol

### Before implementation

1. Read this README (contract summary + canonical links above), then the Start-here list in order.
2. Confirm the selected task ID is explicitly approved; do not start adjacent tasks.
3. Compare `CURRENT_STATUS.md` / `HANDOFF.md` with the repository; resolve discrepancies before coding.
4. Re-read ADR-023 plus D-004/D-005; reject any design that adds authoritative `child_node_ids[]`, product max depth, or a depth column without ADR amendment.
5. Mark the selected task `IN_PROGRESS` in the active phase file.
6. Record the agent/session identifier and start time in `CURRENT_STATUS.md`.

### During implementation

1. Keep hierarchy mutations Course-scoped and transactional; children always derived from `parent_id`.
2. Prefer iterative/CTE traversal over unsafe deep recursion; operational page/payload limits are not product depth.
3. Keep CourseNode `status` and LearningItem `publish_state` distinct.
4. Update the active phase checklist after meaningful milestones.
5. Record architectural decisions in `DECISIONS.md` (summaries only; do not fork ADR-023).
6. Record blockers immediately.
7. Do not mark tests as passed without running them.
8. Do not fabricate evidence or mark tasks `VERIFIED` without proof under `docs/features/course-system/evidence/`.
9. Do not edit applied migrations, ADR-023 body, or production/NUC data.

### Before stopping

1. Update all task statuses in the phase files.
2. Recalculate completion using `npm run course-system:status:sync`.
3. Run `npm run course-system:status:check` and ensure it is a no-op after sync.
4. Record verification commands and results in evidence / `HANDOFF.md`.
5. Write the exact next action and Resume Prompt in `HANDOFF.md`.
6. Ensure `CURRENT_STATUS.md` matches the real repository state.
7. Leave no false `VERIFIED` claims for unfinished work.

### Forbidden without explicit approval

* Introducing authoritative `child_node_ids[]` (or equivalent) persistence.
* Product `MAX_*_DEPTH` / application depth rejection rules.
* Silent overwrite of ADR-023; inventing a `depth` column; renaming CourseNode `status` to `publish_state`.
* Marking new `COURSE-P*` tasks `VERIFIED` without evidence.
* Production/NUC destructive operations.
