# COURSE-P2-T10 Transaction Timeline

Order of operations inside `ReorderLearningItems` (single PostgreSQL transaction):

1. **BEGIN**
2. **Lock course** — `SELECT id FROM courses WHERE id = $course FOR UPDATE`
3. **Lock CourseNode** — `ensureLearningItemNodeInCourse(..., forUpdate=true)`  
   `SELECT id, course_id FROM course_nodes WHERE id = $node AND course_id = $course FOR UPDATE`  
   (cross-course / missing node rejected before sibling work)
4. **Lock siblings** —  
   `SELECT ... FROM learning_items WHERE course_id = $course AND course_node_id = $node`  
   `ORDER BY position ASC, id ASC FOR UPDATE`
5. **Validate**
   - empty siblings ⇒ require empty `ordered_item_ids`; commit noop
   - bidirectional exact-set match (every sibling once; no missing/extra/duplicates)
6. **Idempotency check** — if each sibling already has canonical position matching request index `0..n-1`, **COMMIT noop** (no UPDATE)
7. **Phase 1 updates** — for each locked sibling (lock order), set temporary position  
   `temporaryBase + index` (avoids unique `(course_node_id, position)` collisions)
8. **Verify** phase-1 RETURNING IDs match expected sibling set
9. **Phase 2 updates** — for each locked sibling, set final position from request order; touch `updated_at`
10. **Verify** phase-2 RETURNING IDs match expected sibling set
11. **COMMIT**

## Failure / concurrency notes

- Any error before commit → **ROLLBACK** (no partial writes).
- Empty RETURNING or unique violation on position → `ErrLearningItemReorderConflict` (HTTP 409).
- Concurrent reorders on the same node serialize on course + node + sibling row locks (`FOR UPDATE`).
