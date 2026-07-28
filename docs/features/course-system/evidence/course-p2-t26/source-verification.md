# COURSE-P2-T26 Source Verification

The repository implementation was inspected before writing examples.

| Contract | Source |
|---|---|
| Admin and learner route registration | `api/routes/main.go`, `setupCourseController` |
| Admin CRUD behavior and status codes | `api/controllers/api/v1/learning_item_controller.go` |
| Learner list/detail wrapper and errors | `api/controllers/api/v1/learner_learning_item_controller.go` |
| Request presence/null semantics and response fields | `api/pkg/structs/learning_item_requests.go`, `optional_string.go`, `optional_json.go` |
| Item types, publication states, defaults, ordering | `api/models/learning_item.go` |
| Metadata envelope and backend block enums | `api/models/learning_item_metadata.go` |
| Visibility modes and write-time normalization | `api/models/learning_item_visibility.go` |
| Learner runtime projection | `api/models/learning_item_visibility_runtime.go` |
| JSend response shapes | `api/utils/json_response.go`, `clevergo.tech/jsend` v1.1.3 |
| Public error strings | `api/constants/constant.go` |

## Verified task evidence consulted

- T06: admin CRUD routes, DTOs, authorization, and repository mapping.
- T07: `DRAFT`/`PUBLISHED` create and PATCH semantics.
- T08: authenticated learner published-read contract and enrollment deferral.
- T13: repository-owned runtime visibility projection.
- T22: wrapped learner detail response and node-local published adjacency.

## Important corrections applied

- Learner DTOs omit `course_id`, `course_node_id`, `position`, `created_at`, and
  `updated_at`.
- Backend metadata supports `CALLOUT`, `CODE`, and `TABLE`; backend persistence
  support is not described as equivalent to dedicated frontend rendering.
- Validation uses HTTP 400, not an invented 422 response.
- Enrollment enforcement is explicitly not part of the verified learner
  contract.

