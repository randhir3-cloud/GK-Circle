# COURSE-P2-T21 Query Contract

Canonical sibling order: `ORDER BY position ASC, id ASC`.

## Current

```sql
SELECT ...
FROM learning_items
WHERE course_id = ?
  AND course_node_id = ?
  AND id = ?
LIMIT 1;
```

## Previous (tuple less-than)

```sql
WHERE course_id = ?
  AND course_node_id = ?
  AND (
    position < current_position
    OR (position = current_position AND id < current_id)
  )
ORDER BY position DESC, id DESC
LIMIT 1;
```

## Next (tuple greater-than)

```sql
WHERE course_id = ?
  AND course_node_id = ?
  AND (
    position > current_position
    OR (position = current_position AND id > current_id)
  )
ORDER BY position ASC, id ASC
LIMIT 1;
```

## Required columns / predicates

Must include: `course_id`, `course_node_id`, `position`, `id`.

Must **not** include: `parent_id`, path/depth/`MAX_DEPTH`, recursive CTE, cross-node joins, `publish_state = PUBLISHED` filter.
