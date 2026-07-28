# COURSE-P2-T26 Depth and Sequence Boundary

- A LearningItem may attach to a CourseNode at any hierarchy depth.
- LearningItems are ordered locally within one CourseNode.
- A LearningItem does not own child CourseNodes.
- LearningItems do not form the structural Course tree.
- `child_node_ids[]` is not a LearningItem hierarchy mechanism.
- Current previous/next behavior is node-local and backend-resolved.
- No course-wide sequence or prerequisite graph is claimed.
- No product maximum CourseNode depth is introduced.

These statements align with D-004, D-005, ADR-023, and the frozen Phase 2 /
Phase 5 ownership boundary.

