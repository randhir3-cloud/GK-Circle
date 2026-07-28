# COURSE-P2-T23 Changed Files

## Implementation

- `api/models/learning_item.go` — added `LearningChainItem` DTO (without JSON tags) and implemented `ProjectPublishedLearningChain` under `LearningItemModel`.
- `api/models/learning_item_chain_projection_test.go` (new) — covers the model projection, validation errors, nil UUID checks, mock failure propagation, and strict SQL boundary/keywords assertions.

No controller changes, no route configurations, no migrations, and no frontend modifications.

## Ledger and evidence

- `docs/development/modules/course-system/phases/phase-02-learning-items.md` (T23 VERIFIED checkboxes + table status)
- Course System status / handoff / changelog / health / index surfaces
- `docs/features/course-system/evidence/course-p2-t23/**`

No DOCUMENTATION_FREEZE, ADR, architecture, or ROADMAP ownership edits.
