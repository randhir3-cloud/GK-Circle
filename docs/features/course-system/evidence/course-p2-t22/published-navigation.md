# COURSE-P2-T22 Published Navigation

## Repository-owned published adjacency

To satisfy security and separation of concerns, the learner adjacent resolution is executed entirely at the repository database level:

```go
func (model *LearningItemModel) GetAdjacentPublishedLearningItems(
    courseID, nodeID, currentItemID uuid.UUID,
) (LearningItemAdjacentResult, error)
```

No unrestricted collection is returned to the controller for filtering. Sibling traversal is bounded by SQL query predicates.

## Traversal query format

The SQL queries enforce published filter and deterministic tuple ordering:

### Fetch Current Item (Must be Published)
```sql
SELECT ... FROM "learning_items"
WHERE "id" = $1
  AND "course_id" = $2
  AND "course_node_id" = $3
  AND "publish_state" = 'PUBLISHED'
LIMIT 1
```

### Fetch Previous Published Sibling
```sql
SELECT ... FROM "learning_items"
WHERE "course_id" = $1
  AND "course_node_id" = $2
  AND "publish_state" = 'PUBLISHED'
  AND (
    ("position" < $3)
    OR (("position" = $4) AND ("id" < $5))
  )
ORDER BY "position" DESC, "id" DESC
LIMIT 1
```

### Fetch Next Published Sibling
```sql
SELECT ... FROM "learning_items"
WHERE "course_id" = $1
  AND "course_node_id" = $2
  AND "publish_state" = 'PUBLISHED'
  AND (
    ("position" > $3)
    OR (("position" = $4) AND ("id" > $5))
  )
ORDER BY "position" ASC, "id" ASC
LIMIT 1
```

## Guarantees

- **No Iteration**: Traversal executes with database index lookup, skipping draft items immediately without fetching them into Go memory or looping.
- **Node-Local**: Queries filter strictly by `course_node_id`, ensuring navigation never crosses into adjacent course nodes.
- **T21 Contract Safe**: The unrestricted admin/repository adjacency traversal (`GetAdjacentLearningItems`) remains completely unaffected and doesn't filter on `publish_state`.
