# COURSE-P2-T22 Navigation Contract

## API endpoint

`GET /api/v1/learner/courses/:course_id/nodes/:node_id/learning-items/:item_id`

## Authentication

Required. Learner identity must be authenticated via Ory Kratos context.

## Response payload wrapper

The response wraps properties inside a `learning_item` sub-object, adding minimal `previous` and `next` structures.

### Breaking Changes: YES

This is a response-shape contract change. Existing clients expecting flat data structures under the `"data"` field will receive the nested structure.

### Minimal Navigation DTO

The navigation entries return only the approved adjacent `id` and `title` fields. Position, publish state, node internals, and admin attributes are explicitly excluded.

## Schema shape

```json
{
  "status": "success",
  "data": {
    "learning_item": {
      "id": "019c02a0-1111-7000-8000-000000000031",
      "title": "Introduction to State PCS Preparation",
      "item_type": "ARTICLE",
      "description": "Getting started roadmap for Prelims & Mains",
      "metadata": {
        "version": 1,
        "blocks": [
          {
            "id": "b1",
            "type": "TEXT",
            "data": {
              "text": "Detailed syllabus analysis and study schedule."
            },
            "visibility": {
              "mode": "ALL"
            }
          }
        ]
      },
      "publish_state": "PUBLISHED"
    },
    "previous": null,
    "next": {
      "id": "019c02a0-1111-7000-8000-000000000032",
      "title": "History: Ancient Civilizations"
    }
  }
}
```

## Boundary conditions

- **First Item**: `previous` is `null`
- **Last Item**: `next` is `null`
- **Single Item**: Both `previous` and `next` are `null`
- **Draft Sibling**: Skipped; the link targets the next closest published sibling or `null`
- **Draft Current Item**: Excluded from lookup, returning a standard HTTP 404 (non-disclosing error, identical to a missing ID)
