# COURSE-P2-T07 Publish State

## Column

- Table: `learning_items`
- Column: `publish_state` TEXT NOT NULL DEFAULT `'DRAFT'`
- Constraint name (exact): `learning_items_publish_state_check`
- Allowed: `DRAFT`, `PUBLISHED` (uppercase only)

## Create matrix

| Request | Result |
|---|---|
| omitted | persist `DRAFT` |
| `"DRAFT"` | persist `DRAFT` |
| `"PUBLISHED"` | persist `PUBLISHED` |
| `""` / `null` / lowercase / unknown | 400 |

## PATCH matrix

| Request | Result |
|---|---|
| omitted | unchanged |
| `"DRAFT"` | update |
| `"PUBLISHED"` | update |
| `""` | 400 |
| `null` | 400 |
| `"draft"` / `"published"` / `"INVALID"` | 400 |

## Presence model

- HTTP: `structs.OptionalString` (`Present` / `Null` / `Value`)
- Repository update: `*LearningItemPublishState` only after exact-enum proof
- Omitted and explicit null never collapse

## Out of scope

`published_at`, scheduling, learner visibility, `/publish` endpoints, workflow, notifications, frontend.
