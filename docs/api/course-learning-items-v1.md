# Course LearningItem API Contract (v1)

This document is the canonical API reference for the existing node-scoped
LearningItem admin CRUD and authenticated learner read endpoints.

The implementation is authoritative. These examples describe current behavior;
they do not add routes, validation, permissions, enrollment, or sequencing.

## Base URL and envelopes

All routes use the `/api/v1` prefix and return JSend-compatible JSON.

Successful response:

```json
{
  "status": "success",
  "data": {}
}
```

Client-visible failure:

```json
{
  "status": "fail",
  "data": "learning item not found"
}
```

Server or authentication-middleware error:

```json
{
  "status": "error",
  "message": "error while getting learning item",
  "code": 500
}
```

The JSend `status` field describes request execution. It is not
`CourseNode.status` or `LearningItem.publish_state`.

## Domain and hierarchy boundary

- A LearningItem may attach to a CourseNode at any hierarchy depth.
- LearningItems are ordered within one CourseNode. List order is defined by the
  server as `position` ascending, then `id` ascending.
- A LearningItem does not own child CourseNodes and does not form the structural
  Course tree. Clients must not introduce or use `child_node_ids[]` for
  LearningItem hierarchy.
- Current previous/next navigation is node-local. It does not define a
  course-wide learning sequence. Course-wide sequence and prerequisite graphs
  belong to a later phase.
- `CourseNode.status` is the CourseNode lifecycle field.
  `LearningItem.publish_state` is the LearningItem publication field and accepts
  only `DRAFT` or `PUBLISHED`. The two concepts are distinct.
- There is no product maximum CourseNode depth.

## LearningItem values and metadata

Supported LearningItem `item_type` values:

- `ARTICLE`
- `VIDEO`
- `PDF`
- `LINK`
- `QUIZ_REFERENCE`

Supported backend information-block `type` values:

- `TEXT`
- `HEADING`
- `IMAGE`
- `VIDEO`
- `PDF`
- `LINK`
- `CALLOUT`
- `CODE`
- `TABLE`
- `DIVIDER`

Backend persistence support does not imply that every block type has a
dedicated frontend renderer.

The current metadata envelope is:

```json
{
  "version": 1,
  "blocks": [
    {
      "id": "intro-heading",
      "type": "HEADING",
      "data": {
        "text": "Introduction",
        "level": 2
      },
      "visibility": {
        "mode": "ALL"
      }
    }
  ]
}
```

Metadata structure, unique block IDs, supported block types, placeholder syntax,
and visibility are validated when metadata is written. Supported visibility
modes are `ALL`, `AUTHENTICATED`, `PREMIUM`, `INSTRUCTOR`, and `HIDDEN`.
Omitted block visibility defaults to `ALL`; explicit null or invalid visibility
fails validation.

Learner reads receive a server-generated metadata projection. `ALL` and
`AUTHENTICATED` blocks are retained for the current authenticated learner
context. `PREMIUM`, `INSTRUCTOR`, and `HIDDEN` blocks are currently omitted
because no premium entitlement source is available and learners do not receive
instructor/hidden blocks. Admin reads remain unfiltered.

## Admin LearningItem CRUD

Admin routes require Ory Kratos authentication and the existing Course-admin
allowlist authorization. Course and CourseNode ownership come from path
parameters; body identifiers are not authoritative.

Route family:

```text
/api/v1/admin/courses/{course_id}/nodes/{node_id}/learning-items
```

### List LearningItems

```http
GET /api/v1/admin/courses/{course_id}/nodes/{node_id}/learning-items
```

The response includes draft and published items visible to the authorized
administrator, in server-defined order.

```json
{
  "status": "success",
  "data": [
    {
      "id": "019c02a0-1111-7000-8000-000000000101",
      "course_id": "019c01c6-b8b7-7f4a-9e7a-f62650d5e481",
      "course_node_id": "019c01c8-4bdd-78e2-a366-690bfd600280",
      "title": "Introduction to Indian Polity",
      "item_type": "ARTICLE",
      "description": "Overview of constitutional foundations.",
      "metadata": {
        "version": 1,
        "blocks": []
      },
      "position": 0,
      "publish_state": "DRAFT",
      "created_at": "2026-07-26T09:00:00Z",
      "updated_at": "2026-07-26T09:00:00Z"
    }
  ]
}
```

An existing node with no items returns HTTP 200:

```json
{
  "status": "success",
  "data": []
}
```

### Create a LearningItem

```http
POST /api/v1/admin/courses/{course_id}/nodes/{node_id}/learning-items
Content-Type: application/json
```

`title` and `item_type` are required. `description`, `metadata`, and
`publish_state` are optional.

For `QUIZ_REFERENCE`, `quiz_id` is required and must identify a Quiz the
administrator owns or a public Quiz they are allowed to manage. For every
other `item_type`, `quiz_id` must be omitted or null. The relationship is
enforced by an additive `ON DELETE RESTRICT` foreign key. Existing legacy
`QUIZ_REFERENCE` rows created before this relationship may remain unbound
until an administrator edits and repairs them; new writes cannot create
another unbound reference.

```json
{
  "title": "Introduction to Indian Polity",
  "item_type": "ARTICLE",
  "description": "Overview of constitutional foundations.",
  "metadata": {
    "version": 1,
    "blocks": []
  },
  "quiz_id": null,
  "publish_state": "DRAFT"
}
```

Create semantics:

- omitted or null `description` is stored as null;
- omitted or null `metadata` uses `{"version":1,"blocks":[]}`;
- omitted `publish_state` uses `DRAFT`;
- null, lowercase, empty, or unknown `publish_state` fails;
- the server appends the item after existing node-local siblings and assigns its
  `position`.

HTTP 201 response:

```json
{
  "status": "success",
  "data": {
    "id": "019c02a0-1111-7000-8000-000000000101",
    "course_id": "019c01c6-b8b7-7f4a-9e7a-f62650d5e481",
    "course_node_id": "019c01c8-4bdd-78e2-a366-690bfd600280",
    "title": "Introduction to Indian Polity",
    "item_type": "ARTICLE",
    "description": "Overview of constitutional foundations.",
    "metadata": {
      "version": 1,
      "blocks": []
    },
    "position": 0,
    "publish_state": "DRAFT",
    "created_at": "2026-07-26T09:00:00Z",
    "updated_at": "2026-07-26T09:00:00Z"
  }
}
```

The timestamp values above are illustrative response values, not fixed
contractual values.

### Get a LearningItem

```http
GET /api/v1/admin/courses/{course_id}/nodes/{node_id}/learning-items/{item_id}
```

HTTP 200 returns the complete admin representation shown in the create response,
including `position` and `publish_state`.

### Update a LearningItem

```http
PATCH /api/v1/admin/courses/{course_id}/nodes/{node_id}/learning-items/{item_id}
Content-Type: application/json
```

PATCH is presence-aware. Omitted fields remain unchanged; it does not replace
the whole resource.

Scalar update:

```json
{
  "title": "Constitutional Foundations",
  "item_type": "ARTICLE"
}
```

Clear the description:

```json
{
  "description": null
}
```

Update metadata:

```json
{
  "metadata": {
    "version": 1,
    "blocks": [
      {
        "id": "intro-heading",
        "type": "HEADING",
        "data": {
          "text": "Introduction",
          "level": 2
        },
        "visibility": {
          "mode": "ALL"
        }
      }
    ]
  }
}
```

Publish:

```json
{
  "publish_state": "PUBLISHED"
}
```

PATCH semantics:

- `description: null` clears the field;
- `metadata: null` is invalid;
- null `title`, `item_type`, or `publish_state` is invalid;
- omitted fields remain unchanged;
- an empty object fails because at least one field is required.

HTTP 200 returns the complete updated admin representation.

### Delete a LearningItem

```http
DELETE /api/v1/admin/courses/{course_id}/nodes/{node_id}/learning-items/{item_id}
```

HTTP 200:

```json
{
  "status": "success",
  "data": "success"
}
```

A missing item returns the same public not-found response documented below.

## Learner published reads

Learner routes require Ory Kratos authentication. They do not require the admin
allowlist and expose no write methods.

Enrollment enforcement is not part of the currently verified learner
LearningItem contract. Publication filtering and runtime visibility projection
are server-owned and must not be reproduced by clients.

Route family:

```text
/api/v1/learner/courses/{course_id}/nodes/{node_id}/learning-items
```

### List published LearningItems

```http
GET /api/v1/learner/courses/{course_id}/nodes/{node_id}/learning-items
```

The repository returns published items only, ordered by `position` and then
`id`. Draft items are not returned. Learner response objects intentionally omit
`course_id`, `course_node_id`, `position`, `created_at`, and `updated_at`.

```json
{
  "status": "success",
  "data": [
    {
      "id": "019c02a0-1111-7000-8000-000000000101",
      "title": "Introduction to Indian Polity",
      "item_type": "ARTICLE",
      "description": "Overview of constitutional foundations.",
      "metadata": {
        "version": 1,
        "blocks": []
      },
      "publish_state": "PUBLISHED"
    }
  ]
}
```

An accessible node with no published items returns HTTP 200:

```json
{
  "status": "success",
  "data": []
}
```

### Get a published LearningItem

```http
GET /api/v1/learner/courses/{course_id}/nodes/{node_id}/learning-items/{item_id}
```

The backend supplies node-local `previous` and `next` values. It skips draft
siblings and returns null at a boundary. Clients must not calculate adjacency
or infer a course-wide sequence.

```json
{
  "status": "success",
  "data": {
    "learning_item": {
      "id": "019c02a0-1111-7000-8000-000000000101",
      "title": "Introduction to Indian Polity",
      "item_type": "ARTICLE",
      "description": "Overview of constitutional foundations.",
      "metadata": {
        "version": 1,
        "blocks": []
      },
      "publish_state": "PUBLISHED"
    },
    "previous": null,
    "next": {
      "id": "019c02a0-1111-7000-8000-000000000102",
      "title": "Fundamental Rights"
    }
  }
}
```

A missing item and a draft current item both return the same HTTP 404 response,
preventing draft discovery.

## Error responses

Only stable public fields are contractual. Internal errors, SQL details, stack
traces, generated identifiers, and timestamps are not exposed as examples.

### 400 Invalid request

Examples include invalid UUID path parameters, an empty PATCH, invalid metadata,
and invalid publication state.

```json
{
  "status": "fail",
  "data": "publish_state must be DRAFT or PUBLISHED"
}
```

### 401 Unauthenticated

Kratos middleware rejects a missing session before controller or repository
work:

```json
{
  "status": "error",
  "message": "error no session_id found in kratos cookie",
  "code": 401
}
```

### 403 Admin authorization denied

This response applies to authenticated users who are not in the existing
Course-admin allowlist:

```json
{
  "status": "fail",
  "data": "access denied. You do not have course administration permissions."
}
```

### 404 LearningItem not found

```json
{
  "status": "fail",
  "data": "learning item not found"
}
```

Node/course ownership mismatches use the implementation's corresponding
non-disclosing Course or CourseNode not-found response.

### 409 Conflict

```json
{
  "status": "fail",
  "data": "learning item conflict"
}
```

### 500 Internal error

The public message is operation-specific. For example:

```json
{
  "status": "error",
  "message": "error while getting learning item",
  "code": 500
}
```

## Authority and non-goals

- Admin authorization is enforced by Kratos authentication plus the existing
  Course-admin allowlist.
- Learner reads require Kratos authentication.
- Publication filtering and runtime visibility projection are repository-owned.
- Clients are transport and presentation consumers only.
- This contract does not introduce enrollment enforcement, media upload, AI
  generation, placeholder resolution, metadata authoring, reorder/move UI,
  GraphQL, OpenAPI generation, or course-wide navigation.
