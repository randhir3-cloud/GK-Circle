# COURSE-P2-T02 Repository Tests

## Focused LearningItem coverage

- Create append positions `0`, `1`, `2` via `MAX(position)+1` / empty-node `0`
- Get / list / update / delete scoped by `course_id` + `course_node_id`
- Missing node → `ErrLearningItemNodeNotFound`
- Missing item → `ErrLearningItemNotFound`
- Cross-course node → `ErrLearningItemCrossCourse`
- Invalid type / blank title / invalid metadata
- Empty update rejected; metadata null rejected
- Unique position conflict → `ErrLearningItemConflict`
- Persistence errors do not leak SQL details

## Regression

- `go test ./models -count=1` — PASSED (includes Course / CourseNode suites)
- `go test ./controllers/api/v1 -count=1` — PASSED (T08–T10 HTTP unchanged)
- Docker full `go test ./...`, race on models/controllers/structs, build — PASSED
