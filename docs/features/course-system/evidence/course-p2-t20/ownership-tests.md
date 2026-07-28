# COURSE-P2-T20 Ownership Tests

File: `api/models/learning_item_ownership_test.go`

## What “≥10 nesting levels” means

A deterministic chain of **10** CourseNode UUIDs (depths 1..10). LearningItem SQL does not store or filter by depth/path; the leaf node ID is locked/looked up with `(course_id, id)` via `ensureLearningItemNodeInCourse`.

No hierarchy-depth validation and no product `MAX_DEPTH`.

## Cases

| Test | Proves |
|---|---|
| `TestLearningItemOwnershipAttachListGetAcrossTenNestingLevels` | Create + List + Get succeed at every depth 1..10 |
| `TestLearningItemOwnershipRootVsDeepParity` | Identical success + cross-Course failure semantics for root (depth 1) and deep (depth 10) |
| `TestLearningItemOwnershipDeepPublishedReads` | T08 published list/get on depth-10 leaf; draft published-get → not-found |
| `TestLearningItemOwnershipMissingDeepNode` | Missing deep node → `ErrLearningItemNodeNotFound` |
| `TestLearningItemOwnershipSQLShapeNoDepthPredicates` | Guard against accidental depth/path predicates in ownership SQL shapes |

## Cross-Course / no leak

- Create/List when node belongs to another course → `ErrLearningItemCrossCourse`
- Get with wrong `courseID` or wrong `nodeID` → `ErrLearningItemNotFound`
- Errors and empty results must not embed foreign course titles or unrelated item payloads
