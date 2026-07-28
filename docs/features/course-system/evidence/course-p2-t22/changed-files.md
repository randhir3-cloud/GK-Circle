# COURSE-P2-T22 Changed Files

## Implementation

- `api/models/learning_item.go` — adapted internal helper `findAdjacentLearningItem`, implemented `GetAdjacentPublishedLearningItems`
- `api/models/learning_item_adjacent_test.go` — added `TestLearningItemPreviousNextPublished` and `TestAdjacentPublishedLearningItemSQLBoundary`
- `api/pkg/structs/learning_item_requests.go` — defined `LearnerLearningItemNavigation` and `LearnerLearningItemDetailResponse` DTOs
- `api/controllers/api/v1/learner_learning_item_controller.go` — updated `GetByID` to use all-or-nothing retrieval and return wrapped payload
- `api/controllers/api/v1/learner_learning_item_previous_next_test.go` (new) — covers learner adjacent HTTP endpoint behaviors
- `api/controllers/api/v1/learner_learning_item_visibility_test.go` — updated to expect wrapped detail response
- `api/controllers/api/v1/learning_item_draft_filtering_test.go` — updated `published_200` to expect wrapped detail response

No database migrations or frontend modifications.

## Ledger and evidence

- `docs/development/modules/course-system/phases/phase-02-learning-items.md` (T22 VERIFIED checkboxes + table status)
- Course System status / handoff / changelog / health / index surfaces
- `docs/features/course-system/evidence/course-p2-t22/**`

No DOCUMENTATION_FREEZE, ADR, architecture, or ROADMAP ownership edits.
