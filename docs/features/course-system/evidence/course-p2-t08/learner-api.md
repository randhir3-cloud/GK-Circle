# COURSE-P2-T08 Learner LearningItem APIs

## Routes

| Method | Path | Auth |
|---|---|---|
| GET | `/api/v1/learner/courses/:course_id/nodes/:node_id/learning-items` | `KratosAuthenticated` |
| GET | `/api/v1/learner/courses/:course_id/nodes/:node_id/learning-items/:item_id` | `KratosAuthenticated` |

No POST/PATCH/DELETE. No quiz-admin allowlist. No anonymous access.

## Repository contract

| Surface | Methods |
|---|---|
| Admin (unchanged) | `CreateLearningItem`, `UpdateLearningItem`, `DeleteLearningItem`, `GetLearningItemByID`, `ListLearningItemsByNode` |
| Learner (new) | `GetPublishedLearningItemByID`, `ListPublishedLearningItemsByNode` |

Private unexported helpers share SQL:

- `getLearningItem(courseID, nodeID, itemID, publishedOnly bool)`
- `listLearningItems(courseID, nodeID, publishedOnly bool)`

Publish filtering (`publish_state = 'PUBLISHED'`) is enforced in the repository SQL.
Controllers never filter `publish_state`.

## Security model

| Case | Result |
|---|---|
| Missing item | `ErrLearningItemNotFound` → 404 |
| Draft item | `ErrLearningItemNotFound` → 404 (no draft discovery) |
| Published item | 200 + learner DTO |
| Unauthenticated | 401 |
| Authenticated non-admin | allowed (enrollment deferred) |

## Response fields

`id`, `title`, `item_type`, `description`, `metadata`, `publish_state`

Omitted: `course_id`, `course_node_id`, `position`, `created_at`, `updated_at`

## Known limitation

Enrollment checks are deferred for this task. Publish state is the only content filter.
