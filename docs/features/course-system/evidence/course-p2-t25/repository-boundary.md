# COURSE-P2-T25 Repository Boundary

Publish filtering remains repository-owned:

- `ListPublishedLearningItemsByNode` owns the SQL `publish_state=PUBLISHED`
  predicate and learner visibility projection.
- `GetPublishedLearningItemByID` owns the published-item lookup and projected
  metadata.
- `GetAdjacentPublishedLearningItems` owns published previous/next resolution.

The controller does not inspect `publish_state`, parse visibility metadata,
sort list results, loop to discover adjacent items, or perform list filtering.
The tests use sqlmock at the repository boundary to prove that returned values
are transported without an additional controller policy layer.
