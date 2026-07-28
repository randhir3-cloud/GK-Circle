# COURSE-P1-T03 Model Test Results

The focused tests cover initialization, pre-query input validation, Course and
parent locking, missing and cross-Course parents, top-level SECTION/SUBJECT and
child TOPIC creation, exact positions, DRAFT status, private logical-path
invariants, JSON path exclusion, deterministic/default UUID generation,
semantic position conflicts, rollback paths, commit failure, and Course-scoped
lookup.

Final results:

- `go test ./models -count=1`: PASSED, exit 0.
- Docker Go 1.23 `go vet ./...`: PASSED, exit 0.
- Docker Go 1.23 `go test ./... -count=1`: PASSED, exit 0.
- Docker Go 1.23 `go test -race ./... -count=1`: PASSED, exit 0.
- Docker Go 1.23 `go build ./...`: PASSED, exit 0.
- `gofmt -l models/course_node.go models/course_node_test.go`: no output.

The tests assert path behavior through the private codec and ADR invariants;
callers cannot provide a path and normal JSON never exposes it.
