# COURSE-P2-T15 Behavioral Coverage

| Frozen area | Result | Regression source |
|---|---|---|
| Create defaults, required fields, position, timestamps | PASSED | `learning_item_backend_suite_test.go`, existing controller/model tests |
| Ordered/empty list and full admin get | PASSED | T15 controller tests |
| Presence-aware PATCH and null rejection | PASSED | T15 controller/model tests |
| Delete and missing item | PASSED | T15 controller tests |
| 401, 403, 404, 409, 500 and JSend shapes | PASSED | T15 controller tests |
| Metadata version, blocks, IDs, placeholders, visibility | PASSED | Existing metadata/placeholder/visibility tests plus enum gap test |
| Append ordering and deterministic `(position,id)` order | PASSED | Existing repository tests and T15 list assertion |
| Deep ownership and course/node isolation | PASSED | Existing T20/T21 tests |
| Transaction rollback and conflicts | PASSED | Existing reorder/move/model tests |
| Published-only learner reads and server order | PASSED | Existing publication tests and strengthened exact-shape assertions |
| Learner DTO excludes admin-only fields | PASSED | Strengthened exact key-set assertions |
| Runtime metadata visibility projection | PASSED | Existing T13 tests |
| Reorder and move | PASSED | Existing T10/T11 suites |
| Previous/next, draft skipping, node-local chain | PASSED | Existing T22/T24 controller/model suites |
| Enrollment enforcement regression | PASSED | `TestLearnerLearningItemRoutesDenyUnenrolled`, enrollment model/controller tests, enrolled delivery suites |

No numerical coverage threshold was introduced. `PASSED` means the mapped
behavior executed successfully in the focused and full suites; it does not
override the blocked enrollment category.
