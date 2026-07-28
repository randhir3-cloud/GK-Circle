# COURSE-P2-T23 Projection Contract

## Reusable Repository API

```go
type LearningChainItem struct {
	ID    uuid.UUID `db:"id"`
	Title string    `db:"title"`
}

func (model *LearningItemModel) ProjectPublishedLearningChain(
	courseID uuid.UUID,
	nodeID uuid.UUID,
) ([]LearningChainItem, error)
```

## Transport Neutrality

The `LearningChainItem` DTO is designed exclusively for repository-level queries and does not define JSON tags. It contains only structural elements of type `uuid.UUID` and `string`.

## Validation Contract

- **Nil Course ID**: `courseID == uuid.Nil` -> returns `nil, ErrCourseNotFound`.
- **Nil Node ID**: `nodeID == uuid.Nil` -> returns `nil, ErrLearningItemNodeNotFound`.
- **Missing Node**: Returns `nil, ErrLearningItemNodeNotFound` via `ensureLearningItemNodeInCourse` validation check.
- **Wrong Course**: Returns `nil, ErrLearningItemCrossCourse` via `ensureLearningItemNodeInCourse` validation check.
- **Database/Validation Failures**: Sourced error is returned directly wrapping `ErrLearningItemPersistence` without remapping.

## Output Specification

- Only items with `publish_state = 'PUBLISHED'` are selected.
- Sorted deterministically using `ORDER BY position ASC, id ASC`.
- An empty node returns a non-nil empty slice (`[]LearningChainItem{}, nil`).
- Failed queries return `nil, err`.
