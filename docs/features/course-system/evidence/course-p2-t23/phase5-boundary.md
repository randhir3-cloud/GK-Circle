# COURSE-P2-T23 Phase 5 Boundary

This projection is strictly a flat list projection within a **single CourseNode** (`node-local` only).

## Explicit Exclusions

The following concepts are not part of `COURSE-P2-T23` and belong entirely to Phase 5:
- **No recursive SQL / CTEs**: Traversal does not lookup parent/child nodes.
- **No `child_node_ids`**: The model does not store child nodes.
- **No prerequisite DAG / lesson graph**: No dependency relations are evaluated.
- **No continue learning / progress tracking**: Enrollment gates or user progress are ignored.
- **No navigation structures / edges**: This is not an adjacency or navigation graph.
- **No course-wide sequence**: No cross-node ordering is supported.

## Architectural Purpose

This model is intended for localized listings inside a single module node (e.g. state PCS topic indexes). It is decoupled from any structural tree routing, sequencing, or unlocking mechanisms.
