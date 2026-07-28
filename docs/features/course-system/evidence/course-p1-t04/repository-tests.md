# COURSE-P1-T04 Repository Tests

Focused CourseNode tests verify:

- unknown Course versus existing empty Course;
- non-nil empty root and hierarchy slices;
- Course-scoped root and child queries;
- missing and cross-Course parent handling;
- nested multi-level assembly without parent pointers;
- private path exclusion from nested JSON;
- explicit hierarchy preorder with child positions `2`, `10`, and `100`;
- malformed/disconnected hierarchy detection through reachable-count comparison;
- CTE query failure mapping to the existing persistence wrapper.

The isolated PostgreSQL test ran the same recursive CTE against a temporary
database. It returned `Root:0:4`, `Two:2:4`, `Ten:10:4`, and
`Hundred:100:4`, confirming the 10-digit position segment preserves numeric
order across the supported nonnegative PostgreSQL integer range.
