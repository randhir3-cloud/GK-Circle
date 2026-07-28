# COURSE-P2-T11 Mutation Tests

Focused model tests (`LearningItemMove`) cover:

- parameter validation without database access (nil course/nodes, same-node, nil/duplicate ordered IDs);
- subset mismatch (missing-from-source, foreign-node, foreign-course treated as mismatch without existence leak);
- missing source/destination nodes;
- empty noop with real locked sibling counts;
- empty destination success;
- single-item and multi-item moves with canonical source/dest renumber;
- staging order: both nodes stage to disjoint temps before any `course_node_id` change;
- UUID lock order when destination UUID < source UUID;
- concurrent empty-RETURNING conflict and unique-violation conflict → `ErrLearningItemMoveConflict`;
- rollback on injected update failure;
- int32 overflow rejection for temporary position ranges;
- `verifyLearningItemMoveResult` ownership and contiguity invariants.

Controller tests (`AdminLearningItemMove`) cover auth (401/403), invalid JSON/UUID, same-node, mismatch, conflict (409), success payload, and empty-noop payload with frozen fields.
