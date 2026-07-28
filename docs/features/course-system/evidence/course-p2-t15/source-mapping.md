# COURSE-P2-T15 Source Mapping

| Contract area | Primary test sources |
|---|---|
| Admin HTTP CRUD/JSend/auth/errors | `api/controllers/api/v1/learning_item_backend_suite_test.go`, `learning_item_controller_test.go` |
| Learner HTTP publication/shape/errors | `learner_learning_item_controller_test.go`, `learner_learning_item_publish_contract_test.go`, `learning_item_draft_filtering_test.go` |
| Previous/next HTTP | `learner_learning_item_previous_next_test.go` |
| Create/update/repository CRUD | `api/models/learning_item_backend_suite_test.go`, `learning_item_test.go` |
| Metadata and placeholders | `learning_item_metadata_test.go`, `learning_item_placeholders_test.go` |
| Visibility normalization/runtime projection | `learning_item_visibility_test.go`, `learning_item_visibility_runtime_test.go`, controller visibility tests |
| Published reads | `learning_item_draft_filtering_test.go` |
| Reorder | `learning_item_reorder_test.go`, controller reorder cases |
| Move/rollback/conflict | `learning_item_move_test.go`, controller move cases |
| Deep ownership | `learning_item_ownership_test.go` |
| Node-local adjacency and draft skipping | `learning_item_adjacent_test.go`, `learning_item_chain_projection_test.go` |

The existing verified T10/T11/T13/T20/T21/T22/T23/T24/T25 suites are reused
as authoritative regression coverage; completed scenarios were not duplicated.

Enrollment has no mapped production authority or valid regression source. See
`enrollment-blocker.md`.
