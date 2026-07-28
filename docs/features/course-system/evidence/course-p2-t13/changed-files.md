# COURSE-P2-T13 Changed Files

## Implementation

- `api/models/learning_item_visibility_runtime.go` — `LearnerVisibilityAccess`, `AuthenticatedLearnerVisibilityAccess`, `ProjectLearningItemForLearner`, `filterBlocksForLearner`
- `api/models/learning_item.go` — wire projection into `GetPublishedLearningItemByID` / `ListPublishedLearningItemsByNode` (atomic list failure)
- `api/models/learning_item_visibility.go` — comment update (runtime owned by T13)
- `api/models/learning_item_visibility_runtime_test.go` — `LearningItemVisibilityRuntime*` pack
- `api/controllers/api/v1/learner_learning_item_controller.go` — repository-owned projection comment only
- `api/controllers/api/v1/learner_learning_item_visibility_test.go` — learner/admin contrast HTTP tests

No migrations, routes, DTOs, or frontend.

## Ledger and evidence

- `docs/development/modules/course-system/phases/phase-02-learning-items.md` (T13 VERIFIED)
- Course System status / handoff / changelog surfaces
- `docs/features/course-system/evidence/course-p2-t13/**`
