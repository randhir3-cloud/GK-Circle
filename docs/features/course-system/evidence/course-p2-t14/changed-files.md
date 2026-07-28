# COURSE-P2-T14 Changed Files

## Production frontend

- Learner list and detail pages under `app/pages/courses/`.
- Reusable renderer and block presentation components under
  `app/components/learning-items/`.
- `app/composables/learner_learning_items.js`.
- `app/utils/content_url.js`.

## Tests

- Renderer and content URL tests under
  `app/test/components/LearningItems/`.
- Learner list/detail page tests under `app/test/pages/`.
- Learner transport tests under `app/test/composables/`.

## Dependency-completion changes

- `app/composables/learner_learning_items.js` — formatting only.
- `app/test/pages/LearnerLearningItemsPages.test.js` — T12 enrollment mocks,
  enrollment success/failure regressions, wrapper cleanup, and responsive root
  assertions.
- Both learner list/detail page roots — `w-full flex-1` responsive sizing after
  the browser audit exposed signed-out shrink-to-fit behavior.
- This evidence bundle and the authorized ledger/status surfaces.

No T14-owned backend, DTO, schema, migration, Compose, or deployment file
changed. The repository already contained extensive pre-existing untracked
Course-system API/app work; those files remain outside this closure run.
