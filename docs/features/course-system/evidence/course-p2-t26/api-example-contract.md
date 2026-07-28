# COURSE-P2-T26 API Example Coverage

The normative examples are in
[`docs/api/course-learning-items-v1.md`](../../../../api/course-learning-items-v1.md).
This file is a verification matrix, not a competing API specification.

## Admin CRUD coverage

| Method | Route suffix | Success | Example coverage |
|---|---|---:|---|
| GET | `/admin/courses/{course_id}/nodes/{node_id}/learning-items` | 200 | ordered list and successful empty array |
| POST | `/admin/courses/{course_id}/nodes/{node_id}/learning-items` | 201 | supported create fields, defaults, full response |
| GET | `/admin/courses/{course_id}/nodes/{node_id}/learning-items/{item_id}` | 200 | complete admin representation |
| PATCH | `/admin/courses/{course_id}/nodes/{node_id}/learning-items/{item_id}` | 200 | scalar update, description null, metadata, publication |
| DELETE | `/admin/courses/{course_id}/nodes/{node_id}/learning-items/{item_id}` | 200 | `data: "success"` |

Admin examples expose the exact DTO fields and distinguish omitted PATCH fields
from explicit null values.

## Learner read coverage

| Method | Route suffix | Success | Example coverage |
|---|---|---:|---|
| GET | `/learner/courses/{course_id}/nodes/{node_id}/learning-items` | 200 | published-only ordered list and empty array |
| GET | `/learner/courses/{course_id}/nodes/{node_id}/learning-items/{item_id}` | 200 | `learning_item`, `previous`, and `next` wrapper |

Learner examples omit admin-only identity, position, and timestamp fields.
Navigation examples contain only backend-provided `id` and `title`.

## Error coverage

The canonical document contains source-confirmed JSend examples for HTTP 400,
401, 403, 404, 409, and 500. It contains no invented 422 contract, internal SQL
details, stack traces, trace IDs, or contractual timestamps.
