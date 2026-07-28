# COURSE-P2-T21 Previous/Next Semantics

## Method

`LearningItemModel.GetAdjacentLearningItems(courseID, nodeID, currentItemID) (LearningItemAdjacentResult, error)`

```go
type LearningItemAdjacentResult struct {
    Previous *LearningItem
    Next     *LearningItem
}
```

## Node-local boundary

Resolution uses only siblings under the same `(course_id, course_node_id)`.

It does **not** traverse:

- parent / child CourseNodes
- ancestors / descendants
- another node's LearningItems
- another Course
- course-wide sequence (Phase 5)

## Chain ends

| Situation | Previous | Next | Error |
|---|---|---|---|
| First sibling | `nil` | next item or `nil` | none |
| Last sibling | previous item or `nil` | `nil` | none |
| Single-item node | `nil` | `nil` | none |
| Middle | previous | next | none |

Nil means boundary, not failure.

## Current item ownership

Current item is loaded with:

`WHERE course_id = ? AND course_node_id = ? AND id = ?`

Wrong course/node/missing item → `ErrLearningItemNotFound` (or existing node helper cross-course / node-not-found). No “exists elsewhere” signal.

## Publication

T21 resolves the **full** repository/admin sibling chain. It does **not** filter `publish_state`.

- **COURSE-P2-T22** owns authenticated learner previous/next and draft skipping.
- **Phase 5** owns course-wide sequence navigation.
