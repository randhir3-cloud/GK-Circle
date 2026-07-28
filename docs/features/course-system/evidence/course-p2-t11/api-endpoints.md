# COURSE-P2-T11 Admin Move API

## Endpoint

`POST /api/v1/admin/courses/:course_id/nodes/:node_id/learning-items/move`

Path `:node_id` is the **source** node. Registered before `/:item_id`. Auth: Kratos + quiz-admin allowlist (`authorizeCourseAdmin`).

## Request

```json
{
  "destination_node_id": "uuid",
  "ordered_item_ids": ["uuid", "..."]
}
```

Subset validation against siblings of the source node:

- every requested ID must belong to the source sibling set
- duplicates rejected
- foreign-node / foreign-course IDs → mismatch without existence leak
- empty array is a successful noop after course + both nodes validate

## Response (200)

```json
{
  "source_node_id": "...",
  "destination_node_id": "...",
  "items_moved": 1,
  "source_item_count": 1,
  "destination_item_count": 2,
  "noop": false
}
```

Empty noop (counts = actual locked sibling lengths):

```json
{
  "source_node_id": "...",
  "destination_node_id": "...",
  "items_moved": 0,
  "source_item_count": 2,
  "destination_item_count": 1,
  "noop": true
}
```

## Errors

| Condition | HTTP | Constant |
|---|---|---|
| Invalid destination / ordered UUID | 400 | `ErrLearningItemMoveInvalid` |
| Duplicate / missing / foreign IDs | 400 | `ErrLearningItemMoveMismatch` |
| Source == destination | 400 | `ErrLearningItemMoveSameNode` |
| Missing course / wrong course for node | 404 | course / course-node not found |
| Non-admin | 403 | course admin forbidden |
| Concurrency / unique conflict | 409 | `ErrLearningItemMoveConflict` |
