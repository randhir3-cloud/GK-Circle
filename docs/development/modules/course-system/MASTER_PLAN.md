# Universal Chained Course System â€” Master Plan

Implementation blueprint index. Task checklists: `phases/`. Order narrative: `ROADMAP.md`.
Architecture: `architecture/current.md` (**D-004**, **D-005**, ADR-023).

Documentation governance: DOC-CS-T01 (architecture) â†’ DOC-CS-T02 (draft) â†’
**DOC-CS-T02-R1** (roadmap freeze) â†’ **DOC-CS-T03** (normalization) â†’ **DOC-CS-T04** (freeze certification; done). See `DOCUMENTATION_FREEZE.md`.

## Overall status

| Phase | Name | Weight | Status | Completion |
|---|---|---:|---|---:|
| 1 | Unlimited Recursive Course Tree Foundation | 15% | IN_PROGRESS | 32.26% |
| 2 | Learning Items and Information Blocks | 15% | VERIFIED | 100.00% |
| 3 | Resources and Native Content | 15% | NOT_STARTED | 0% |
| 4 | Assessments (any-depth bindings) | 15% | NOT_STARTED | 0% |
| 5 | Navigation, Unlocking, and Learning Sequence | 10% | NOT_STARTED | 0% |
| 6 | Recursive Progress and Completion | 15% | NOT_STARTED | 0% |
| 7 | Templates and Builder | 10% | NOT_STARTED | 0% |
| 8 | Production Hardening | 5% | NOT_STARTED | 0% |

## Overall completion formula

Overall completion is calculated as:
$$\text{Overall Completion} = \sum (\text{Phase Weight} \times \text{Verified Checklist Percentage of that Phase})$$

Do not calculate overall completion by averaging the phase percentages unless all phase weights are equal.

## Phase order

Phases must normally execute in order.
A later phase may begin only when:
* the preceding phase is `VERIFIED`; or
* an explicit architectural decision records why parallel work is safe (see D-003).

## Product model (all phases)

```text
Course â†’ CourseNode(parent_id authoritative; children derived) â†’ LearningItem*
```

Unlimited logical depth. No fixed curriculum schema levels.

## Global acceptance conditions

The system is complete only when:
* all eight phases are verified;
* canonical backend verification passes;
* frontend lint and build pass;
* Playwright passes where applicable;
* migration safety is documented;
* unresolved critical blockers equal zero;
* documentation matches the implementation;
* no production deployment was performed as part of development.
