# COURSE-P2-T09 UI Contract

- Route: `/admin/courses/learning-items`.
- Sidebar visibility reuses `canCreatePublicQuiz`; server authorization remains
  authoritative.
- Courses are explicitly selected. No first-record auto-selection occurs.
- Selecting a Course loads only roots.
- Selecting a node loads its node-local LearningItems and direct children.
- Each child selection appends one selector level. Changing an earlier
  selection discards deeper state, and request tokens ignore stale responses.
- No full-tree request, flattened tree, inferred parent, depth cap, or
  `child_node_ids[]` ownership exists.
- LearningItems render in the exact server array order. `position` is read-only.
- Create exposes title, type, description, and publish state; defaults are
  `ARTICLE` and `DRAFT`; blank description and metadata are omitted.
- Update uses PATCH scalar fields only, sends `description: null` when clearing,
  and preserves metadata by omission.
- Delete requires confirmation. Every successful mutation refreshes the list.
- Loading, error, prompt, and successful-empty states are distinct for Courses,
  CourseNodes, and LearningItems.
- There are no optimistic writes, autosave, sorting, move/reorder controls,
  metadata/block editing, or duplicated backend business validation.
