# COURSE-P2-T11 Transaction Timeline

Order of operations inside `MoveLearningItems` (single PostgreSQL transaction):

1. **BEGIN**
2. **Lock course** — `SELECT id FROM courses WHERE id = $course FOR UPDATE`
3. **Lock CourseNodes by ascending UUID** — both source and destination via `ensureLearningItemNodeInCourse(..., forUpdate=true)` in UUID byte order (prevents AB/BA deadlocks)
4. **Lock siblings in the same node-UUID order** —  
   `SELECT ... FROM learning_items WHERE course_id = $course AND course_node_id = $node`  
   `ORDER BY position ASC, id ASC FOR UPDATE` for each node
5. **Validate**
   - subset match: every `ordered_item_ids` entry belongs to the source sibling set
   - empty `ordered_item_ids` ⇒ **COMMIT noop** with `source_item_count` / `destination_item_count` = actual locked lengths
6. **Stage temps (both nodes)** — move all source then destination siblings to disjoint temporary positions (`sourceTempBase` / `destTempBase`) without touching `course_node_id`
7. **Ownership change** — UPDATE `course_node_id` for each moved item (source → destination) while still at source temp positions
8. **Final source positions** — remaining source siblings compacted to canonical `0..n-1` (touch `updated_at`)
9. **Final destination positions** — existing destination order + moved request order compacted to `0..m-1` (touch `updated_at`)
10. **Verify** — re-lock both sibling sets (source then dest) and assert IDs, positions, and `course_node_id` ownership
11. **COMMIT**

## Failure / concurrency notes

- Any error before commit → **ROLLBACK** (no partial writes).
- Empty RETURNING or unique violation on `(course_node_id, position)` → `ErrLearningItemMoveConflict` (HTTP 409).
- Concurrent moves/reorders serialize on course + node + sibling row locks (`FOR UPDATE`).
