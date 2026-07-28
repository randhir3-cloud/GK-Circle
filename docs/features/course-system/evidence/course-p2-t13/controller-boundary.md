# COURSE-P2-T13 Controller Boundary

## Learner controller

`LearnerLearningItemController` remains transport-only:

- Auth via `requireAuthenticatedLearner`
- Calls `ListPublishedLearningItemsByNode` / `GetPublishedLearningItemByID`
- Maps repository results with `toLearnerLearningItemResponse`
- Does **not** iterate blocks, switch on visibility modes, or filter `publish_state`

Comment on the controller documents that publish filtering and visibility projection are repository-owned.

## Admin controller

Admin LearningItem GET/list paths unchanged; full metadata including HIDDEN / INSTRUCTOR / PREMIUM blocks is returned.

## DTO / routes

- No new DTOs
- No new routes
- Existing learner response shape unchanged (`metadata` already carries projected JSON)
