# COURSE-P1-T10 HTTP Tests

## Route inventory

| Method | Path | AuthN | AuthZ |
|---|---|---|---|
| PATCH | `/api/v1/admin/courses/:course_id/nodes/:node_id/move` | KratosAuthenticated | allowlist |
| POST | `/api/v1/admin/courses/:course_id/nodes/reorder` | KratosAuthenticated | allowlist |
| DELETE | `/api/v1/admin/courses/:course_id/nodes/:node_id` | KratosAuthenticated | allowlist |

`/reorder` is registered before `/:node_id`. Still absent: `POST .../move`,
`DELETE .../subtree`, `PATCH .../:node_id` (non-move field update).

## Request DTO inventory

- `ReqMoveAdminCourseNode` — presence-aware `position`, optional `new_parent_id`
- `ReqReorderAdminCourseNodes` — optional `parent_id`, structural `ordered_node_ids`

## Success envelope

All mutations return `utils.JSONSuccess(c, 200, "success")` (T08/T09 jsend shape).
No 204 / empty-body responses.

## Coverage

- Auth 401 / 403 on mutation routes
- Move position presence (omit/null/negative) and parent parsing
- Move missing parent → 404; root noop → 200 then GET tree
- Reorder structural rejects (missing/null/empty/duplicates/empty/whitespace/malformed)
- Reorder sibling mismatch → 400; valid child reorder → 200 then GET children
- Delete missing node/course → 404; leaf delete → 200 then GET tree
- Absence regression for still-unowned routes only
