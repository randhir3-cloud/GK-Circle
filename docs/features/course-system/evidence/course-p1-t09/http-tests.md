# COURSE-P1-T09 HTTP Tests

## Route inventory

| Method | Path | AuthN | AuthZ |
|---|---|---|---|
| POST | `/api/v1/admin/courses/:course_id/nodes` | KratosAuthenticated | allowlist |
| GET | `/api/v1/admin/courses/:course_id/nodes` | KratosAuthenticated | allowlist |
| GET | `/api/v1/admin/courses/:course_id/nodes/tree` | KratosAuthenticated | allowlist |
| GET | `/api/v1/admin/courses/:course_id/nodes/:node_id` | KratosAuthenticated | allowlist |
| GET | `/api/v1/admin/courses/:course_id/nodes/:node_id/children` | KratosAuthenticated | allowlist |

`/tree` is registered before `/:node_id`. Hierarchy mutation routes are absent.

## Request/response DTO inventory

- `ReqCreateAdminCourseNode` — title, node_type, required `OptionalInteger` position, `OptionalString` parent_id
- `OptionalInteger` — omit / null / value (omit and null rejected for position)
- `AdminCourseNodeResponse` — UUID fields, no `path`, RFC3339Nano timestamps
- `AdminCourseHierarchyResponse` — snake_case `course_id` + nested `roots`

## Create contract

| Field | Rule |
|---|---|
| position omitted / null / negative | 400 |
| position `0` | valid |
| parent_id omitted / null | root |
| parent_id valid UUID | child |
| parent_id empty / whitespace / malformed | 400 |
| node_type | `SECTION` / `SUBJECT` / `TOPIC` via model constants |

## Error mapping matrix

| Condition | Status |
|---|---|
| Missing Kratos session / identity | 401 |
| Authenticated non-allowlisted | 403 |
| Malformed JSON / UUID / omitted position / empty parent | 400 |
| Domain validation / cross-Course parent on create | 400 |
| Position conflict | 409 |
| Missing Course / scoped node missing / cross-Course read | 404 |
| Unexpected persistence | logged 500 |

## Coverage

- Unauthenticated 401 via real `KratosAuthenticated`
- Non-admin 403 with Course-admin wording
- Create root/child; parent omit/null/UUID/empty/malformed; position omit/0/-1/null
- Body injection: `course_id` and `owner_id` ignored; persisted course from route
- Cross-Course get and children → 404
- Empty tree for existing Course → `{ course_id, roots: [] }`
- Missing Course list → 404
- Mutation absences assert 404 or 405 for PATCH/DELETE/move/reorder/delete-subtree
- Path never exposed in JSON responses
