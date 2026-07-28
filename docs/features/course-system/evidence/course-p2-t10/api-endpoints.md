# COURSE-P2-T10 Admin Reorder API

## Endpoint

`POST /api/v1/admin/courses/:course_id/nodes/:node_id/learning-items/reorder`

Registered before `/:item_id`. Auth: Kratos + quiz-admin allowlist (`authorizeCourseAdmin`).

## Request

```json
{
  "ordered_item_ids": ["uuid", "uuid", "..."]
}
```

Bidirectional exact-set validation against siblings of `:node_id`:

- every existing sibling exactly once
- no missing IDs
- no extra IDs
- no duplicates

Empty array succeeds only when the node has zero LearningItems.

## Response (200)

```json
{
  "course_node_id": "...",
  "learning_item_count": 8,
  "positions_updated": 5,
  "noop": false
}
```

Canonical / empty noop:

```json
{
  "course_node_id": "...",
  "learning_item_count": 0,
  "positions_updated": 0,
  "noop": true
}
```

## Errors

| Condition | HTTP | Constant |
|---|---|---|
| Invalid UUID / empty entry | 400 | `ErrLearningItemReorderInvalid` |
| Duplicate / missing / extra / foreign-node IDs | 400 | `ErrLearningItemReorderMismatch` |
| Missing course / wrong course for node | 404 | course / course-node not found |
| Non-admin | 403 | course admin forbidden |
| Concurrency / unique conflict | 409 | `ErrLearningItemReorderConflict` |
