# COURSE-P2-T11 Move Algorithm

## UUID lock order

```text
course FOR UPDATE
→ min(sourceNodeID, destinationNodeID) FOR UPDATE
→ max(sourceNodeID, destinationNodeID) FOR UPDATE
→ siblings of min(node) FOR UPDATE (ORDER BY position, id)
→ siblings of max(node) FOR UPDATE (ORDER BY position, id)
```

When destination UUID is lexicographically less than source UUID, destination node and its siblings are locked first. This matches CourseNode dual-target move hygiene and avoids AB/BA deadlock pairs.

Post-move verification re-locks **source then destination** (ownership-oriented), not UUID order.

## Temporary position bases

```text
maxPosition = max(position across both locked sibling sets)
sourceTempBase = maxPosition + sourceCount + destCount + 1
destTempBase   = sourceTempBase + sourceCount
```

Source siblings stage to `[sourceTempBase, sourceTempBase+sourceCount)`.
Destination siblings stage to `[destTempBase, destTempBase+destCount)`.
Ranges are disjoint, so ownership change cannot collide on the unique `(course_node_id, position)` constraint.

## Six-step staging

1. Stage source siblings to source temps
2. Stage destination siblings to dest temps
3. UPDATE `course_node_id` for moved IDs (source → destination)
4. Compact remaining source to `0..n-1`
5. Compact destination (prior dest order + moved request order) to `0..m-1`
6. Verify re-read + COMMIT

Temps are written **before** any ownership change. Final writes touch `updated_at` via `goqu.L("now()")` (no extra bind arg). Position UPDATE bind order remains `(position, courseID, nodeID, itemID)`. Course-node UPDATE bind order is `(newNodeID, courseID, currentNodeID, itemID)`.

## Result counts

- Empty noop: counts are pre-move locked lengths.
- Mutating move: counts are post-verify re-lock lengths.
